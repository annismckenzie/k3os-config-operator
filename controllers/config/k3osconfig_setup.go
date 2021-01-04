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
	"os"
	"time"

	configv1alpha1 "github.com/annismckenzie/k3os-config-operator/apis/config/v1alpha1"
	"github.com/annismckenzie/k3os-config-operator/pkg/nodes"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// K3OSConfigReconciler reconciles a K3OSConfig object.
type K3OSConfigReconciler struct {
	client                 client.Client
	clientset              *kubernetes.Clientset
	logger                 logr.Logger
	scheme                 *runtime.Scheme
	leader                 bool
	defaultRequeueResponse ctrl.Result
	nodeLister             listersv1.NodeLister
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
func (r *K3OSConfigReconciler) SetupWithManager(mgr ctrl.Manager, options ...Option) error {
	r.defaultRequeueResponse = ctrl.Result{RequeueAfter: time.Second * 30}
	r.nodeLister = nodes.NewNodeLister()

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

	if !r.leader { // wrap manager so this can run without a leader lease
		mgr = &nonLeaderLeaseNeedingManagerWrapper{Manager: mgr}
	}

	// cannot inject via inject.LoggerInto because `leader` field isn't set at that point
	r.logger = mgr.GetLogger().
		WithName("controllers").
		WithName("K3OSConfig").
		WithValues("podName", os.Getenv("HOSTNAME"), "leader", r.leader)

	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.K3OSConfig{}).
		//Watches()  // TODO: can I watch a secret named k3os-nodes with this? That'd be rad.
		Complete(r)
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
