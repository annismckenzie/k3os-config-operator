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

	configv1alpha1 "github.com/annismckenzie/k3os-config-operator/apis/config/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// response is a helper struct to cut down on the amount of if and switch statements.
type response struct {
	result reconcile.Result
	err    error
}

// allow operator to handle K3OSConfig CR objects in its namespace
// +kubebuilder:rbac:groups=config.operators.annismckenzie.github.com,resources=k3osconfigs,verbs=get;list;watch;create;update;patch;delete,namespace=k3os-config-operator-system
// +kubebuilder:rbac:groups=config.operators.annismckenzie.github.com,resources=k3osconfigs/status,verbs=get;update;patch,namespace=k3os-config-operator-system

// allow operator to get and watch Secret objects in its namespace
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;watch,namespace=k3os-config-operator-system

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

	switch r.leader {
	case true: // this instance of the operator won the leader election and can update the K3OSConfig CR
		response, err = r.handleK3OSConfigAsLeader(ctx, config)
	default: // handle k3os config file
		response, err = r.handleK3OSConfig(ctx, config)
	}

	if err != nil {
		r.logger.Error(err, "failed to handle K3OSConfig")
	}
	return response.result, response.err
}

func (r *K3OSConfigReconciler) handleK3OSConfigAsLeader(ctx context.Context, config *configv1alpha1.K3OSConfig) (*response, error) {
	return &response{result: r.defaultRequeueResponse, err: nil}, nil
}

func (r *K3OSConfigReconciler) handleK3OSConfig(ctx context.Context, config *configv1alpha1.K3OSConfig) (*response, error) {
	return &response{result: r.defaultRequeueResponse, err: nil}, nil
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
