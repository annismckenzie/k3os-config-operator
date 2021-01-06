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
	"github.com/annismckenzie/k3os-config-operator/pkg/errors"
	"github.com/annismckenzie/k3os-config-operator/pkg/nodes"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

// allow operator to handle K3OSConfig CR objects in its namespace
// +kubebuilder:rbac:groups=config.operators.annismckenzie.github.com,resources=k3osconfigs,verbs=get;list;watch;create;update;patch;delete,namespace=k3os-config-operator-system
// +kubebuilder:rbac:groups=config.operators.annismckenzie.github.com,resources=k3osconfigs/status,verbs=get;update;patch,namespace=k3os-config-operator-system

// allow operator to get and watch Secret objects in its namespace
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;watch,namespace=k3os-config-operator-system

// allow operator to update Node objects (the verbs deliberately do not include create and delete)
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;update;patch

// Reconcile handles K3OSConfig CRs.
func (r *K3OSConfigReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()

	config := &configv1alpha1.K3OSConfig{}
	if err := r.client.Get(ctx, req.NamespacedName, config); err != nil {
		if apierrors.IsNotFound(err) { // request object not found, could have been deleted after reconcile request, return and don't requeue
			return result, nil
		}
		r.logger.Error(err, "failed to fetch K3OSConfig")
		return result, err
	}
	config = config.DeepCopy()
	r.logger.Info("successfully fetched K3OSConfig", "spec", config.Spec)

	switch r.leader {
	case true: // this instance of the operator won the leader election and can update the K3OSConfig CR
		result, err = r.handleK3OSConfigAsLeader(ctx, config)
	default: // handle k3os config file
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

func (r *K3OSConfigReconciler) handleK3OSConfig(ctx context.Context, config *configv1alpha1.K3OSConfig) (ctrl.Result, error) {
	// 1. get node name we're running on from the environment
	nodeName := os.Getenv(consts.NodeNameEnvName)
	if nodeName == "" {
		err := fmt.Errorf("failed to find node name in %q environment variable", consts.NodeNameEnvName)
		return ctrl.Result{}, err
	}

	// 2. fetch secret with node configs
	secret, err := r.clientset.CoreV1().Secrets(config.GetNamespace()).Get(ctx, consts.NodeConfigSecretName, metav1.GetOptions{})
	if err != nil {
		return ctrl.Result{}, err
	}

	// 3. get node config
	nodeConfigBytes, ok := secret.Data[nodeName]
	if !ok {
		err = fmt.Errorf("failed to find node %q in config (keys: %v)", nodeName, secretKeys(secret))
		return ctrl.Result{}, err
	}
	nodeConfig, err := nodes.ParseNodeConfig(nodeConfigBytes)
	if err != nil {
		return ctrl.Result{}, err
	}
	r.logger.Info("successfully fetched config of node", "config", nodeConfig)

	node, err := r.nodeLister.Get(nodeName)
	if err != nil {
		return ctrl.Result{}, err
	}

	var updateNode bool

	// 4. sync node labels
	if config.Spec.SyncNodeLabels {
		if err = nodes.NewLabeler().Reconcile(node, nodeConfig.K3OS.Labels); err == nil {
			updateNode = true
		} else if !errors.Is(err, errors.ErrSkipUpdate) { // error is non-nil but it's not the one telling us to skip the update, bail
			return ctrl.Result{}, err
		}
	}

	// 5. sync node taints
	if config.Spec.SyncNodeTaints {
		if err = syncNodeTaints(node, map[string]string{}); err == nil { // FIXME: this does nothing for now
			updateNode = true
		} else if !errors.Is(err, errors.ErrSkipUpdate) { // error is non-nil but it's not the one telling us to skip the update, bail
			return ctrl.Result{}, err
		}
	}

	if updateNode {
		r.logger.Info("updating node", "labels", node.GetLabels(), "addedLabels", node.GetAnnotations()[consts.GetAddedLabelsNodeAnnotation()], "addedTaints", node.GetAnnotations()[consts.GetAddedTaintsNodeAnnotation()])
		// move out into updateNode so it's testable
		if _, err = r.clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{}); err != nil { // FIXME: switch to a better way that tries a couple times!
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// syncNodeLabels syncs node labels between the existing set of labels and the ones given in the configuration.
// This method returns errSkipUpdate if no updates should be done.
func syncNodeLabels(node *corev1.Node, configNodeLabels map[string]string) error {
	//nodeLabels := node.GetLabels()

	// put added labels (as a slice) into the annotation addedLabelsNodeAnnotation

	// FIXME: if we added some labels before and they were removed from the configuration then we wouldn't delete them from the node labels ever
	// that needs to be fixed
	if len(configNodeLabels) == 0 {
		return errors.ErrSkipUpdate
	}

	// check node labels

	return errors.ErrSkipUpdate
}

// syncNodeTaints syncs node taints between the existing set of taints and the ones given in the configuration.
// This method returns errSkipUpdate if no updates should be done.
func syncNodeTaints(node *corev1.Node, configNodeTaints map[string]string) error {
	//nodeTaints := node.Spec.Taints

	// put added taints (as a slice) into the annotation addedTaintsNodeAnnotation

	// FIXME: if we added some taints before and they were dropped from the configuration then we wouldn't delete them from the node taints ever
	// that needs to be fixed
	if len(configNodeTaints) == 0 {
		return errors.ErrSkipUpdate
	}

	// check node taints

	return errors.ErrSkipUpdate
}

func secretKeys(secret *corev1.Secret) []string {
	keys := make([]string, 0, len(secret.Data))
	for key := range secret.Data {
		keys = append(keys, key)
	}
	return keys
}
