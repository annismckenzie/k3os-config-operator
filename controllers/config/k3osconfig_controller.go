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
	"github.com/annismckenzie/k3os-config-operator/pkg/errors"
	"github.com/annismckenzie/k3os-config-operator/pkg/nodes"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// allow operator to handle K3OSConfig CR objects in its namespace
// +kubebuilder:rbac:groups=config.operators.annismckenzie.github.com,resources=k3osconfigs,verbs=get;list;watch;create;update;patch;delete,namespace=k3os-config-operator-system
// +kubebuilder:rbac:groups=config.operators.annismckenzie.github.com,resources=k3osconfigs/status,verbs=get;update;patch,namespace=k3os-config-operator-system

// allow operator to get, list and watch Secret objects in its namespace
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch,namespace=k3os-config-operator-system

// allow operator to update Node objects (the verbs deliberately do not include create and delete)
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;update;patch

// allow operator to use the k3os-config-operator-manager PodSecurityPolicy
// +kubebuilder:rbac:groups=policy,resources=podsecuritypolicies,verbs=use,resourceNames=k3os-config-operator-manager,namespace=k3os-config-operator-system

// Reconcile handles K3OSConfig CRs.
func (r *K3OSConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	config := &configv1alpha1.K3OSConfig{}
	if err := r.client.Get(ctx, req.NamespacedName, config); err != nil {
		if apierrors.IsNotFound(err) { // request object not found, could have been deleted after reconcile request, return and don't requeue
			return result, nil
		}
		r.logger.Error(err, "failed to fetch K3OSConfig")
		return result, err
	}
	config = config.DeepCopy()
	r.logger.V(1).Info("successfully fetched K3OSConfig", "spec", config.Spec)

	if r.leader { // this instance of the operator won the leader election and can update the K3OSConfig CR
		result, err = r.handleK3OSConfigAsLeader(ctx, config)
	} else { // handle k3os config file
		result, err = r.handleK3OSConfig(ctx, config)
	}

	if err != nil {
		r.logger.Error(err, "failed to handle K3OSConfig")
	}
	return result, err
}

func (r *K3OSConfigReconciler) handleK3OSConfigAsLeader(ctx context.Context, config *configv1alpha1.K3OSConfig) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func resultError(err error, logger logr.Logger) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errors.ErrSkipUpdate):
		return nil
	case apierrors.IsNotFound(err), apierrors.IsGone(err):
		logger.Info("object is gone, not requeuing")
		return nil
	case apierrors.IsForbidden(err):
		logger.Error(err, "failed to execute operation, did you forget to apply some RBAC rules?")
		return nil
	case errors.Is(err, os.ErrPermission):
		logger.Error(err, "failed to execute operation, permission denied was returned")
		return nil
	default:
		return err
	}
}

func (r *K3OSConfigReconciler) handleK3OSConfig(ctx context.Context, k3OSConfig *configv1alpha1.K3OSConfig) (ctrl.Result, error) {
	// 1. get node name we're running
	nodeName := r.configuration.NodeName

	// 2. get node config
	nodeConfig, err := r.getNodeConfig(ctx, nodeName)
	if err != nil {
		return ctrl.Result{}, resultError(err, r.logger)
	}
	r.logger.V(1).Info("successfully fetched node config", "config", nodeConfig)

	// 3. get node
	node, err := r.getNode(ctx, nodeName)
	if err != nil {
		return ctrl.Result{}, resultError(err, r.logger)
	}

	var updateNode bool

	// 4. sync node labels
	labeler := nodes.NewLabeler()
	if k3OSConfig.Spec.SyncNodeLabels {
		if err = labeler.Reconcile(node, nodeConfig.K3OS.Labels); err == nil {
			updateNode = true
		} else if err = resultError(err, r.logger); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 5. sync node taints
	tainter := nodes.NewTainter()
	if k3OSConfig.Spec.SyncNodeTaints {
		if err = tainter.Reconcile(node, nodeConfig.K3OS.Taints); err == nil {
			updateNode = true
		} else if err = resultError(err, r.logger); err != nil {
			return ctrl.Result{}, err
		}
	}

	// 6. update node only on changes
	if updateNode {
		err = r.updateNode(ctx, node)
		switch {
		case err == nil:
			r.logger.Info("successfully updated node", "updatedLabels", labeler.UpdatedLabels(), "updatedTaints", tainter.UpdatedTaints())
		case apierrors.IsConflict(err): // check for conflict which happens from time to time and should not blow up the log
			return ctrl.Result{}, errors.New("node object was changed, requeuing")
		default:
			return ctrl.Result{}, err // bail and pass error to caller
		}
	} else {
		r.logger.V(1).Info("skipped updating node")
	}

	// 7. update the config file on disk (if enabled – which is checked inside the updater)
	configFileUpdater := nodes.NewK3OSConfigFileUpdater(r.configuration)
	updateErr := configFileUpdater.Update(nodeConfig)
	if err = resultError(updateErr, r.logger); err != nil {
		return ctrl.Result{}, err
	}
	switch {
	case updateErr == nil: // would equal errors.ErrSkipUpdate if the update was skipped
		r.logger.Info("successfully updated node config on disk")
	case errors.Is(updateErr, errors.ErrSkipUpdate):
		r.logger.V(1).Info("skipped updating node config on disk")
	default:
		r.logger.Error(updateErr, "failed to update node config on disk")
	}

	return ctrl.Result{}, nil
}

func (r *K3OSConfigReconciler) getNodeConfig(ctx context.Context, nodeName string) (*configv1alpha1.K3OSConfigFileSpec, error) {
	secret, err := r.clientset.CoreV1().Secrets(r.namespace).Get(ctx, r.configuration.NodeConfigSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	nodeConfigBytes, ok := secret.Data[nodeName]
	if !ok {
		return nil, fmt.Errorf("failed to find node %q in config (keys: %v)", nodeName, secretKeys(secret))
	}
	return configv1alpha1.ParseConfigYAML(nodeConfigBytes)
}

func (r *K3OSConfigReconciler) getNode(ctx context.Context, nodeName string) (*corev1.Node, error) {
	node, err := r.nodeLister.Get(nodeName)
	if err != nil {
		return nil, err
	}
	return node.DeepCopy(), nil
}

func (r *K3OSConfigReconciler) updateNode(ctx context.Context, node *corev1.Node) error {
	if _, err := r.clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}

func secretKeys(secret *corev1.Secret) []string {
	keys := make([]string, 0, len(secret.Data))
	for key := range secret.Data {
		keys = append(keys, key)
	}
	return keys
}
