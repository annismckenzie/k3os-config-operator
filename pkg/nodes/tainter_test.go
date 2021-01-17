package nodes

import (
	"sort"
	"strings"
	"testing"

	"github.com/annismckenzie/k3os-config-operator/pkg/errors"
	internalConsts "github.com/annismckenzie/k3os-config-operator/pkg/internal/consts"
	"github.com/annismckenzie/k3os-config-operator/pkg/util/taints"
	corev1 "k8s.io/api/core/v1"
)

func defaultTaintedNode() *corev1.Node {
	node := defaultNode()
	node.Spec = corev1.NodeSpec{
		Taints: []corev1.Taint{
			{Key: "existingTaint", Value: "existingTaintValue", Effect: corev1.TaintEffectNoSchedule},
		},
	}
	return node.DeepCopy()
}

func taintedNode(nodeTaints []string) *corev1.Node {
	node := defaultTaintedNode()
	l := NewTainter()
	l.Reconcile(node, nodeTaints)
	return node
}

func Test_tainter_Reconcile(t *testing.T) {
	type args struct {
		node             *corev1.Node
		configNodeTaints []string
	}
	tests := []struct {
		name                          string
		args                          args
		wantErr                       error
		expectedTaints                []string
		updatedTaints                 map[string]string // stringified taint => added/removed/changed
		expectedAddedTaintsAnnotation string
	}{
		{
			name:    "passing a nil Node object",
			wantErr: errors.ErrNilObjectPassed,
		},
		{
			name: "Node has no existing taints and no taints are added",
			args: args{
				node:             &corev1.Node{},
				configNodeTaints: nil,
			},
			wantErr: errors.ErrSkipUpdate,
		},
		{
			name: "Node has existing taints and no taints are added",
			args: args{
				node:             defaultTaintedNode(),
				configNodeTaints: nil,
			},
			wantErr: errors.ErrSkipUpdate,
		},
		{
			name: "Node has no existing taints and we add some taints",
			args: args{
				node: &corev1.Node{},
				configNodeTaints: []string{
					"key1=value1:NoSchedule",
					"key1=value1:NoExecute",
				},
			},
			expectedTaints: []string{
				"key1=value1:NoSchedule",
				"key1=value1:NoExecute",
			},
			updatedTaints: map[string]string{
				"key1=value1:NoSchedule": "added",
				"key1=value1:NoExecute":  "added",
			},
			expectedAddedTaintsAnnotation: "key1=value1:NoExecute,key1=value1:NoSchedule",
		},
		{
			name: "Node has existing taints and we add some taints",
			args: args{
				node: defaultTaintedNode(),
				configNodeTaints: []string{
					"key1=value1:NoExecute",
				},
			},
			expectedTaints: []string{
				"existingTaint=existingTaintValue:NoSchedule",
				"key1=value1:NoExecute",
			},
			updatedTaints: map[string]string{
				"key1=value1:NoExecute": "added",
			},
			expectedAddedTaintsAnnotation: "key1=value1:NoExecute",
		},
		{
			name: "Node has existing taints, we added some taints and we don't change them",
			args: args{
				node: taintedNode([]string{"key1=value1:NoExecute"}),
				configNodeTaints: []string{
					"key1=value1:NoExecute",
				},
			},
			expectedTaints: []string{
				"existingTaint=existingTaintValue:NoSchedule",
				"key1=value1:NoExecute",
			},
			updatedTaints:                 map[string]string{},
			expectedAddedTaintsAnnotation: "key1=value1:NoExecute",
			wantErr:                       errors.ErrSkipUpdate,
		},
		{
			name: "Node has existing taints and we add and update some taints",
			args: args{
				node: defaultTaintedNode(),
				configNodeTaints: []string{
					"existingTaint=updatedTaintValue:NoSchedule",
					"key2=value2:NoExecute",
				},
			},
			expectedTaints: []string{
				"existingTaint=updatedTaintValue:NoSchedule",
				"key2=value2:NoExecute",
			},
			updatedTaints: map[string]string{
				"existingTaint=updatedTaintValue:NoSchedule": "changed",
				"key2=value2:NoExecute":                      "added",
			},
			expectedAddedTaintsAnnotation: "existingTaint=updatedTaintValue:NoSchedule,key2=value2:NoExecute",
		},
		{
			name: "Node has existing taints that we added and we remove them",
			args: args{
				node: taintedNode([]string{"key2=value2:NoExecute", "key3=value3:NoSchedule"}),
				configNodeTaints: []string{
					"key2=value2:NoExecute-",
					// this version also needs to take care of removed taints in the list so
					// key3=value3:NoSchedule isn't listed here and it should still be removed
				},
			},
			expectedTaints: []string{
				"existingTaint=existingTaintValue:NoSchedule",
			},
			updatedTaints: map[string]string{
				"key2:NoExecute":         "removed", // FIXME: this should have `key2=value2:NoExecute` but taints are unique by key and effect only
				"key3=value3:NoSchedule": "removed",
			},
			expectedAddedTaintsAnnotation: "",
		},
		{
			name: "Removed taint stays in the list of taints (issue #10)",
			args: args{
				node: taintedNode([]string{"key2=value2:NoExecute"}),
				configNodeTaints: []string{
					"key2=value2:NoExecute-",
					"key4=value4:NoExecute-", // simulate removal of a taint that was removed in a previous reconciliation
				},
			},
			expectedTaints: []string{
				"existingTaint=existingTaintValue:NoSchedule",
			},
			updatedTaints: map[string]string{
				"key2:NoExecute": "removed", // FIXME: this should have `key2=value2:NoExecute` but taints are unique by key and effect only
			},
			expectedAddedTaintsAnnotation: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewTainter()
			err := l.Reconcile(tt.args.node, tt.args.configNodeTaints)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("tainter.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expectedTaints != nil {
				nodeTaints := tt.args.node.Spec.Taints
				expectedTaints, deleteTaints, err := taints.ParseTaints(tt.expectedTaints)
				if err != nil || len(deleteTaints) > 0 {
					t.Errorf("ParseTaints() on tt.expectedTaints returned error %v or taints to delete which is invalid: %v",
						err, deleteTaints)
				}
				added, removed := taints.TaintSetDiff(expectedTaints, nodeTaints)
				if len(added) > 0 || len(removed) > 0 {
					t.Errorf("tainter.Reconcile() expected taints %v but got %v (added: %v, removed: %v)",
						tt.expectedTaints, nodeTaints, added, removed)
				}
			}

			addedTaintsMap := getAddedTaints(tt.args.node)
			var addedTaints []string
			for addedTaint := range addedTaintsMap {
				addedTaints = append(addedTaints, addedTaint.ToString())
			}
			sort.Strings(addedTaints)
			addedTaintsAnnotation := strings.Join(addedTaints, internalConsts.NodeAnnotationValueSeparator)
			if tt.expectedAddedTaintsAnnotation != addedTaintsAnnotation {
				t.Errorf("tainter expected added taints annotation = %q, got %q", tt.expectedAddedTaintsAnnotation, addedTaintsAnnotation)
			}

			updatedTaints := l.UpdatedTaints()
			if len(tt.updatedTaints) != len(updatedTaints) {
				t.Errorf("tainter.UpdatedTaints() expected updated taints = %v (len: %d), got %v (len: %d)",
					tt.updatedTaints, len(tt.updatedTaints), updatedTaints, len(updatedTaints))
			}

			for expectedUpdatedTaint, expectedEffect := range tt.updatedTaints {
				effect, ok := updatedTaints[expectedUpdatedTaint]
				if !ok {
					t.Errorf("tainter expected updated taint %q but did not find it", expectedUpdatedTaint)
				} else if expectedEffect != effect {
					t.Errorf("tainter expected updated taint %q to be %q, got %q instead", expectedUpdatedTaint, expectedEffect, effect)
				}
			}
		})
	}
}
