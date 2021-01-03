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

package controllers

import (
	"context"
	"os"
	"time"

	configv1alpha1 "github.com/annismckenzie/k3os-config-operator/apis/config/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

// response is a helper struct to cut down on the amount of if and switch statements.
type response struct {
	result reconcile.Result
	err    error
}

// K3OSConfigReconciler reconciles a K3OSConfig object.
type K3OSConfigReconciler struct {
	client                 client.Client
	logger                 logr.Logger
	scheme                 *runtime.Scheme
	leader                 bool
	defaultRequeueResponse ctrl.Result
}

// +kubebuilder:rbac:groups=config.operators.annismckenzie.github.com,resources=k3osconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=config.operators.annismckenzie.github.com,resources=k3osconfigs/status,verbs=get;update;patch

// allow operator to update Node objects (the verbs deliberately do not include create and delete)
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;update;patch

// Reconcile handles K3OSConfig CRs.
func (r *K3OSConfigReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	config, response, err := r.fetchK3OSConfig(ctx, req.NamespacedName)
	if err != nil {
		r.logger.Error(err, "failed to fetch K3OSConfig")
		return response.result, response.err
	}

	r.logger.Info("successfully fetched K3OSConfig", "spec", config.Spec)

	return r.defaultRequeueResponse, nil
}

func (r *K3OSConfigReconciler) fetchK3OSConfig(ctx context.Context, name types.NamespacedName) (*configv1alpha1.K3OSConfig, *response, error) {
	config := &configv1alpha1.K3OSConfig{}
	if err := r.client.Get(ctx, name, config); err != nil {
		if errors.IsNotFound(err) { // request object not found, could have been deleted after reconcile request, return and don't requeue
			return nil, &response{result: ctrl.Result{}, err: nil}, err
		}
		return nil, &response{result: r.defaultRequeueResponse, err: nil}, err
	}
	return config.DeepCopy(), nil, nil
}

// SetupWithManager is called in main to setup the K3OSConfig reconiler with the manager as a non-leader.
func (r *K3OSConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.defaultRequeueResponse = ctrl.Result{RequeueAfter: time.Second * 30}

	// cannot inject via inject.LoggerInto because `leader` field isn't set at that point
	r.logger = mgr.GetLogger().
		WithName("controllers").
		WithName("K3OSConfig").
		WithValues("podName", os.Getenv("HOSTNAME"), "leader", r.leader)

	return ctrl.NewControllerManagedBy(mgr).For(&configv1alpha1.K3OSConfig{}).Complete(r)
}

// SetupWithManagerAsLeader is called in main to setup the K3OSConfig reconiler with the manager as a leader.
func (r *K3OSConfigReconciler) SetupWithManagerAsLeader(mgr ctrl.Manager) error {
	r.leader = true

	return r.SetupWithManager(mgr)
}

// NeedLeaderElection satisfies manager.LeaderElectionRunnable interface.
func (r *K3OSConfigReconciler) NeedLeaderElection() bool {
	return r.leader
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
