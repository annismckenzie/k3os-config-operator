package nodes

import (
	"fmt"
	"strings"

	"github.com/annismckenzie/k3os-config-operator/pkg/consts"
	"github.com/annismckenzie/k3os-config-operator/pkg/errors"
	internalConsts "github.com/annismckenzie/k3os-config-operator/pkg/internal/consts"
	corev1 "k8s.io/api/core/v1"
)

// Labeler interface allows the reconciliation of node labels.
type Labeler interface {
	Reconcile(*corev1.Node, map[string]string) error
	GetUpdatedLabels() map[string]string
}

// labeler implements the Labeler interface.
var _ Labeler = (*labeler)(nil)

type labeler struct {
	updatedLabels map[string]string
}

// NewLabeler returns an initialized label reconciler.
func NewLabeler() Labeler {
	return &labeler{
		updatedLabels: map[string]string{},
	}
}

// Reconcile updates a node's labels according to the provided node labels.
// It will return errors.ErrSkipUpdate if no updates to the node are required.
// The provided Node object is updated and the caller must persist the updated
// Node with the Kubernetes API server on their own.
func (l *labeler) Reconcile(node *corev1.Node, configNodeLabels map[string]string) error {
	if node == nil {
		return fmt.Errorf("node: %w", errors.ErrNilObjectPassed)
	}
	nodeLabels := node.GetLabels()
	if nodeLabels == nil {
		nodeLabels = map[string]string{}
	}

	var update bool
	addedLabelsMap := getAddedLabels(node)
	for addedLabel := range addedLabelsMap {
		if _, ok := configNodeLabels[addedLabel]; !ok { // a label that we added was removed, drop it
			delete(nodeLabels, addedLabel)
			delete(addedLabelsMap, addedLabel)
			update = true
			l.updatedLabels[addedLabel] = "(removed)"
		}
	}

	if len(configNodeLabels) > 0 {
		update = true
		for labelKey, labelValue := range configNodeLabels {
			addedLabelsMap[labelKey] = struct{}{}
			nodeLabels[labelKey] = labelValue
			l.updatedLabels[labelKey] = labelValue
		}
	}

	if update {
		node.Labels = nodeLabels
		updateAddedLabels(node, addedLabelsMap)
		return nil
	}

	return errors.ErrSkipUpdate
}

// GetUpdatedLabels returns the updated (added, removed, changed) labels after Reconcile was called.
func (l *labeler) GetUpdatedLabels() map[string]string {
	return l.updatedLabels
}

func getAddedLabels(node *corev1.Node) map[string]struct{} {
	if node == nil {
		return nil
	}

	addedLabelsMap := map[string]struct{}{}
	if addedLabelsAnnotation := node.GetAnnotations()[consts.GetAddedLabelsNodeAnnotation()]; addedLabelsAnnotation != "" {
		for _, addedLabel := range strings.Split(addedLabelsAnnotation, internalConsts.NodeAnnotationValueSeparator) {
			addedLabelsMap[addedLabel] = struct{}{}
		}
	}
	return addedLabelsMap
}

func updateAddedLabels(node *corev1.Node, addedLabelsMap map[string]struct{}) {
	addedLabels := make([]string, len(addedLabelsMap))
	var i int
	for addedLabel := range addedLabelsMap {
		addedLabels[i] = addedLabel
		i++
	}
	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[consts.GetAddedLabelsNodeAnnotation()] = strings.Join(addedLabels, internalConsts.NodeAnnotationValueSeparator)
	node.Annotations = annotations
}
