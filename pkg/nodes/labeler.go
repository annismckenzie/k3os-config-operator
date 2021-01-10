package nodes

import (
	"fmt"
	"strings"

	"github.com/annismckenzie/k3os-config-operator/pkg/consts"
	"github.com/annismckenzie/k3os-config-operator/pkg/errors"
	internalConsts "github.com/annismckenzie/k3os-config-operator/pkg/internal/consts"
	corev1 "k8s.io/api/core/v1"
)

// Labeler allows reconciling node labels.
type Labeler interface {
	Reconcile(*corev1.Node, map[string]string) error
	UpdatedLabels() map[string]string
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
	addedLabelsMap := addedLabels(node)
	for addedLabel := range addedLabelsMap {
		if _, ok := configNodeLabels[addedLabel]; !ok { // a label that we added was removed, drop it
			delete(nodeLabels, addedLabel)
			delete(addedLabelsMap, addedLabel)
			update = true
			l.updatedLabels[addedLabel] = "(removed)"
		}
	}

	for labelKey, labelValue := range configNodeLabels {
		if _, ok := addedLabelsMap[labelKey]; ok && labelValue == nodeLabels[labelKey] { // label already exists and hasn't been changed, skip
			continue
		}

		update = true
		addedLabelsMap[labelKey] = struct{}{}
		nodeLabels[labelKey] = labelValue
		l.updatedLabels[labelKey] = labelValue
	}

	if update {
		node.Labels = nodeLabels
		updateAddedLabels(node, addedLabelsMap)
		return nil
	}

	return errors.ErrSkipUpdate
}

// UpdatedLabels returns the updated (added, removed, changed) labels after Reconcile was called.
func (l *labeler) UpdatedLabels() map[string]string {
	return l.updatedLabels
}

func addedLabels(node *corev1.Node) map[string]struct{} {
	if node == nil {
		return nil
	}

	addedLabelsMap := map[string]struct{}{}
	if addedLabelsAnnotation := node.GetAnnotations()[consts.AddedLabelsNodeAnnotation()]; addedLabelsAnnotation != "" {
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
	annotations[consts.AddedLabelsNodeAnnotation()] = strings.Join(addedLabels, internalConsts.NodeAnnotationValueSeparator)
	node.Annotations = annotations
}
