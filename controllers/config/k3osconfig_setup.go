/*
MIT License

Copyright (c) 2021 Daniel Lohse

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package config

import (
	"context"
	"fmt"
	"os"

	configv1alpha1 "github.com/annismckenzie/k3os-config-operator/apis/config/v1alpha1"
	"github.com/annismckenzie/k3os-config-operator/pkg/consts"
	"github.com/annismckenzie/k3os-config-operator/pkg/nodes"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// K3OSConfigReconciler reconciles a K3OSConfig object.
type K3OSConfigReconciler struct {
	client      client.Client
	clientset   *kubernetes.Clientset
	logger      logr.Logger
	scheme      *runtime.Scheme
	leader      bool
	nodeLister  listersv1.NodeLister
	shutdownCtx context.Context
	namespace   string
}

// Option denotes an option for configuring this controller.
type Option interface{}

type requireLeaderElectionOpt struct {
	requireLeaderElection bool
}

// RequireLeaderElection returns an option that requires the operator being the leader to run this controller instance.
func RequireLeaderElection() Option {
	return &requireLeaderElectionOpt{}
}

// https://github.com/kubernetes-sigs/controller-runtime/pull/921#issuecomment-662187521 doesn't work
// but there's always another way. 🥁 🥁 🥁
type nonLeaderLeaseNeedingManagerWrapper struct {
	manager.Manager
}

func (w *nonLeaderLeaseNeedingManagerWrapper) Add(r manager.Runnable) error {
	return w.Manager.Add(&nonLeaderLeaseNeedingRunnableWrapper{Runnable: r})
}

type nonLeaderLeaseNeedingRunnableWrapper struct {
	manager.Runnable
}

// NeedLeaderElection satisfies manager.LeaderElectionRunnable interface.
func (w *nonLeaderLeaseNeedingRunnableWrapper) NeedLeaderElection() bool {
	return false
}

// SetupWithManager is called in main to setup the K3OSConfig reconiler with the manager as a non-leader.
func (r *K3OSConfigReconciler) SetupWithManager(shutdownCtx context.Context, mgr ctrl.Manager, options ...Option) error {
	r.nodeLister = nodes.NewNodeLister()
	r.shutdownCtx = shutdownCtx
	r.namespace = consts.Namespace()

	clientset, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}
	r.clientset = clientset

	for _, option := range options {
		if _, ok := option.(*requireLeaderElectionOpt); ok {
			r.leader = true
		}
	}

	// cannot inject via inject.LoggerInto because `leader` field isn't set at that point
	r.logger = mgr.GetLogger().
		WithName("controllers").
		WithName("K3OSConfig").
		WithValues("podName", os.Getenv("HOSTNAME"), "leader", r.leader)

	if r.leader { // if we're building the controller for the leader we can bail here
		return ctrl.NewControllerManagedBy(mgr).For(&configv1alpha1.K3OSConfig{}).Complete(r)
	}

	// wrap manager so this can run without a leader lease
	mgr = &nonLeaderLeaseNeedingManagerWrapper{Manager: mgr}
	c := ctrl.NewControllerManagedBy(mgr).For(&configv1alpha1.K3OSConfig{})

	// construct a watch on the Secret resource that contains the node config.yaml files
	opts := []builder.WatchesOption{
		builder.OnlyMetadata, // only watch and cache the metadata of the secrets because we don't need the contents
		builder.WithPredicates(labelSelectorPredicateForSecret()), // filter the list of secrets using a label selector
	}
	c.Watches(&source.Kind{Type: &corev1.Secret{}}, handler.EnqueueRequestsFromMapFunc(r.enqueueObjectsOnChanges), opts...)

	// construct a watch on the Node this operator is running on
	opts = []builder.WatchesOption{
		builder.OnlyMetadata, // only watch and cache the metadata of the nodes because we don't need the contents
		builder.WithPredicates(namePredicateForNode()),
	}
	c.Watches(&source.Kind{Type: &corev1.Node{}}, handler.EnqueueRequestsFromMapFunc(r.enqueueObjectsOnChanges), opts...)

	return c.Complete(r)
}

func namePredicateForNode() predicate.Predicate {
	nodeName := consts.NodeName()

	return predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetName() == nodeName
	})
}

func labelSelectorPredicateForSecret() predicate.Predicate {
	p, err := predicate.LabelSelectorPredicate(consts.LabelSelectorForNodeConfigFileSecret())
	if err != nil {
		// we're panicking here in order to crash the operator because if this doesn't work there's no
		// recourse (and indicates a programmer error when building the label selector above)
		panic(fmt.Sprintf("failed to build label selector predicate for secret: %v", err))
	}
	return p
}

// enqueueObjectsOnChanges is used to enqueue all K3OSConfig resources in the operator's namespace when
// changes happen to the watched resources (secrets, nodes).
func (r *K3OSConfigReconciler) enqueueObjectsOnChanges(object client.Object) []reconcile.Request {
	r.logger.V(1).Info("change to a watched object noticed", "namespace/name", client.ObjectKeyFromObject(object).String())

	// construct a PartialObjectMetadataList for a list of K3OSConfig resources in the operator's namespace
	var k3osconfigs metav1.PartialObjectMetadataList
	k3osconfigs.SetGroupVersionKind(configv1alpha1.GroupVersion.WithKind("K3OSConfigList"))
	if err := r.client.List(r.shutdownCtx, &k3osconfigs, client.InNamespace(r.namespace)); err != nil {
		r.logger.Error(err, "failed to PartialObjectMetadataList all K3OSConfig resources in this namespace")
	}

	requests := make([]reconcile.Request, len(k3osconfigs.Items))
	for i, item := range k3osconfigs.Items {
		requests[i] = reconcile.Request{NamespacedName: types.NamespacedName{Name: item.GetName(), Namespace: item.GetNamespace()}}
	}
	r.logger.V(1).Info("enqueuing requests for all K3OSConfig resources in this namespace", "namespace", r.namespace, "requests", requests)
	return requests
}

// Interface implementations for dependency injection
var _ inject.Client = (*K3OSConfigReconciler)(nil)
var _ inject.Scheme = (*K3OSConfigReconciler)(nil)

// InjectClient satisfies the inject.Client interface.
func (r *K3OSConfigReconciler) InjectClient(client client.Client) error {
	r.client = client
	return nil
}

// InjectScheme satisfies the inject.Scheme interface.
func (r *K3OSConfigReconciler) InjectScheme(scheme *runtime.Scheme) error {
	r.scheme = scheme
	return nil
}
