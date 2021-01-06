package nodes

import (
	"github.com/annismckenzie/k3os-config-operator/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

// Tainter interface allows the reconciliation of node labels.
type Tainter interface {
	Reconcile(*corev1.Node, []string) error
	GetUpdatedTaints() []string
}

// tainter implements the Tainter interface.
var _ Tainter = (*tainter)(nil)

type tainter struct {
	updatedTaints []string
}

// NewTainter returns an initialized taint reconciler.
func NewTainter() Tainter {
	return &tainter{}
}

// Reconcile updates a node's taints according to the provided node taints.
// It will return errors.ErrSkipUpdate if no updates to the node are required.
// The provided Node object is updated and the caller must persist the updated
// Node with the Kubernetes API server on their own.
func (t *tainter) Reconcile(node *corev1.Node, configNodeTaints []string) error {
	return errors.ErrSkipUpdate
}

// GetUpdatedTaints returns the updated (added, removed, changed) taints after Reconcile was called.
func (t *tainter) GetUpdatedTaints() []string {
	return nil
}
