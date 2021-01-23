package nodes

import (
	"fmt"
	"strings"

	"github.com/annismckenzie/k3os-config-operator/pkg/consts"
	"github.com/annismckenzie/k3os-config-operator/pkg/errors"
	internalConsts "github.com/annismckenzie/k3os-config-operator/pkg/internal/consts"
	"github.com/annismckenzie/k3os-config-operator/pkg/util/taints"
	corev1 "k8s.io/api/core/v1"
)

// Tainter allows reconciling node taints.
type Tainter interface {
	Reconcile(*corev1.Node, []string) error
	UpdatedTaints() map[string]string // stringified taint => added/removed/changed
}

// tainter implements the Tainter interface.
var _ Tainter = (*tainter)(nil)

type tainter struct {
	updatedTaints map[string]string
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
	if node == nil {
		return fmt.Errorf("node: %w", errors.ErrNilObjectPassed)
	}

	taintsToAdd, taintsToRemove, err := taints.ParseTaints(configNodeTaints)
	if err != nil {
		return err
	}

	addedTaintsMap := getAddedTaints(node)
	for existingTaint := range addedTaintsMap {
		var skipCheckingTaintsToAdd bool
		for _, taintToRemove := range taintsToRemove {
			taintToRemove := taintToRemove
			if existingTaint.MatchTaint(&taintToRemove) {
				skipCheckingTaintsToAdd = true
				delete(addedTaintsMap, existingTaint)
			}
		}

		if skipCheckingTaintsToAdd {
			continue
		}

		var found bool
		for _, taintToAdd := range taintsToAdd {
			taintToAdd := taintToAdd
			if existingTaint.MatchTaint(&taintToAdd) {
				found = true
				break
			}
		}
		if !found {
			delete(addedTaintsMap, existingTaint)
			taintsToRemove = append(taintsToRemove, existingTaint)
		}
	}
	for _, taintToAdd := range taintsToAdd {
		addedTaintsMap[taintToAdd] = struct{}{}
	}

	updatedTaints := map[string]string{}

	// here we check whether any of the taints that should be removed
	// can even be removed from the node's taints or whether they were
	// potentially already removed in a previous reconciliation
	tmp := make([]corev1.Taint, 0, len(taintsToRemove))
	for _, taintToRemove := range taintsToRemove {
		// check whether the taint to remove even still exists on the node
		taintToRemove := taintToRemove
		if taints.TaintExists(node.Spec.Taints, &taintToRemove) {
			tmp = append(tmp, taintToRemove)
			updatedTaints[taintToRemove.ToString()] = "removed"
		}
	}
	taintsToRemove = tmp

	tmp = make([]corev1.Taint, 0, len(taintsToAdd))
outer:
	for _, taintToAdd := range taintsToAdd {
		taintToAdd := taintToAdd
		for _, taint := range node.Spec.Taints {
			if taint.MatchTaint(&taintToAdd) {
				if taint.ToString() != taintToAdd.ToString() {
					tmp = append(tmp, taintToAdd)
					updatedTaints[taintToAdd.ToString()] = "changed"
				}
				continue outer
			}
		}

		tmp = append(tmp, taintToAdd)
		updatedTaints[taintToAdd.ToString()] = "added"
	}
	taintsToAdd = tmp

	if len(taintsToAdd) == 0 && len(taintsToRemove) == 0 {
		return errors.ErrSkipUpdate
	}

	t.updatedTaints = updatedTaints

	_, newNodeTaints, err := taints.ReorganizeTaints(node, false, taintsToAdd, taintsToRemove)
	if err != nil {
		return err
	}

	node.Spec.Taints = newNodeTaints
	updateAddedTaints(node, addedTaintsMap)

	return nil
}

// UpdatedTaints returns the updated (added, removed, changed) taints after Reconcile was called.
func (t *tainter) UpdatedTaints() map[string]string {
	return t.updatedTaints
}

func getAddedTaints(node *corev1.Node) map[corev1.Taint]struct{} {
	if node == nil {
		return nil
	}

	addedTaintsMap := map[corev1.Taint]struct{}{}
	if addedTaintsAnnotation := node.GetAnnotations()[consts.AddedTaintsNodeAnnotation()]; addedTaintsAnnotation != "" {
		for _, addedTaint := range strings.Split(addedTaintsAnnotation, internalConsts.NodeAnnotationValueSeparator) {
			parsed, _, err := taints.ParseTaints([]string{addedTaint})
			if len(parsed) != 1 || err != nil {
				panic(fmt.Sprintf("could not parse stored taint %q: %v", addedTaint, err))
			}
			addedTaintsMap[parsed[0]] = struct{}{}
		}
	}
	return addedTaintsMap
}

func updateAddedTaints(node *corev1.Node, addedTaintsMap map[corev1.Taint]struct{}) {
	addedTaints := make([]string, len(addedTaintsMap))
	var i int
	for addedTaint := range addedTaintsMap {
		addedTaints[i] = addedTaint.ToString()
		i++
	}
	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[consts.AddedTaintsNodeAnnotation()] = strings.Join(addedTaints, internalConsts.NodeAnnotationValueSeparator)
	node.Annotations = annotations
}
