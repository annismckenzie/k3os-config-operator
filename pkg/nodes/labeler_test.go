package nodes

import (
	"sort"
	"strings"
	"testing"

	"github.com/annismckenzie/k3os-config-operator/pkg/errors"
	internalConsts "github.com/annismckenzie/k3os-config-operator/pkg/internal/consts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultNodeLabels = map[string]string{
	"k3s.io/hostname":                  "n2-node",
	"kubernetes.io/arch":               "arm64",
	"node.kubernetes.io/instance-type": "k3s",
	"someExistingLabel":                "existingValue",
}

func defaultNode() *corev1.Node {
	return (&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "node",
			Labels: defaultNodeLabels,
		},
	}).DeepCopy()
}

func labeledNode(updateLabels map[string]string) *corev1.Node {
	node := defaultNode()
	l := NewLabeler()
	l.Reconcile(node, updateLabels)
	return node
}

func Test_labeler_Reconcile(t *testing.T) {
	type args struct {
		node             *corev1.Node
		configNodeLabels map[string]string
	}
	tests := []struct {
		name                          string
		args                          args
		wantErr                       error
		expectedLabels                map[string]string
		updatedLabels                 map[string]string
		expectedAddedLabelsAnnotation string
	}{
		{
			name:    "passing a nil Node object",
			wantErr: errors.ErrNilObjectPassed,
		},
		{
			name: "Node has no existing labels and no labels are added",
			args: args{
				node:             &corev1.Node{},
				configNodeLabels: nil,
			},
			wantErr: errors.ErrSkipUpdate,
		},
		{
			name: "Node has existing labels and no labels are added",
			args: args{
				node:             defaultNode(),
				configNodeLabels: nil,
			},
			wantErr: errors.ErrSkipUpdate,
		},
		{
			name: "Node has no existing labels and we add some labels",
			args: args{
				node: &corev1.Node{},
				configNodeLabels: map[string]string{
					"someNewLabel": "value",
					"another":      "anotherValue",
				},
			},
			expectedLabels: map[string]string{
				"someNewLabel": "value",
				"another":      "anotherValue",
			},
			updatedLabels: map[string]string{
				"someNewLabel": "value",
				"another":      "anotherValue",
			},
			expectedAddedLabelsAnnotation: "another,someNewLabel",
		},
		{
			name: "Node has existing labels and we add some labels",
			args: args{
				node: defaultNode(),
				configNodeLabels: map[string]string{
					"someNewLabel": "value",
				},
			},
			expectedLabels: map[string]string{
				"k3s.io/hostname":                  "n2-node",
				"kubernetes.io/arch":               "arm64",
				"node.kubernetes.io/instance-type": "k3s",
				"someExistingLabel":                "existingValue",
				"someNewLabel":                     "value",
			},
			updatedLabels: map[string]string{
				"someNewLabel": "value",
			},
			expectedAddedLabelsAnnotation: "someNewLabel",
		},
		{
			name: "Node has existing labels, we added some labels, without any further changes it should skip updates",
			args: args{
				node: labeledNode(map[string]string{"addedLabel": "value"}),
				configNodeLabels: map[string]string{
					"addedLabel": "value",
				},
			},
			expectedLabels: map[string]string{
				"k3s.io/hostname":                  "n2-node",
				"kubernetes.io/arch":               "arm64",
				"node.kubernetes.io/instance-type": "k3s",
				"someExistingLabel":                "existingValue",
				"addedLabel":                       "value",
			},
			wantErr:                       errors.ErrSkipUpdate,
			expectedAddedLabelsAnnotation: "addedLabel",
		},
		{
			name: "Node has existing labels and we add and update some labels",
			args: args{
				node: defaultNode(),
				configNodeLabels: map[string]string{
					"someNewLabel":      "value",
					"someExistingLabel": "newValue",
				},
			},
			expectedLabels: map[string]string{
				"k3s.io/hostname":                  "n2-node",
				"kubernetes.io/arch":               "arm64",
				"node.kubernetes.io/instance-type": "k3s",
				"someExistingLabel":                "newValue",
				"someNewLabel":                     "value",
			},
			updatedLabels: map[string]string{
				"someNewLabel":      "value",
				"someExistingLabel": "newValue",
			},
			expectedAddedLabelsAnnotation: "someExistingLabel,someNewLabel",
		},
		{
			name: "Node has existing labels that we added and we remove them",
			args: args{
				node:             labeledNode(map[string]string{"someExistingLabel": "newValue"}),
				configNodeLabels: map[string]string{},
			},
			expectedLabels: map[string]string{
				"k3s.io/hostname":                  "n2-node",
				"kubernetes.io/arch":               "arm64",
				"node.kubernetes.io/instance-type": "k3s",
			},
			updatedLabels: map[string]string{
				"someExistingLabel": "(removed)",
			},
			expectedAddedLabelsAnnotation: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLabeler()
			err := l.Reconcile(tt.args.node, tt.args.configNodeLabels)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("labeler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expectedLabels != nil {
				nodeLabels := tt.args.node.GetLabels()
				if len(tt.expectedLabels) != len(nodeLabels) {
					t.Errorf("labeler.Reconcile() expected labels = %v (len: %d), got %v (len: %d)",
						tt.expectedLabels, len(tt.expectedLabels), nodeLabels, len(nodeLabels))
				}
				for labelKey, labelValue := range tt.expectedLabels {
					if value, ok := nodeLabels[labelKey]; !ok || labelValue != value {
						t.Errorf("labeler.Reconcile() expected %q key to have value %q, but got %q (has key: %v)", labelKey, labelValue, value, ok)
					}
				}
			}
			updatedLabels := l.UpdatedLabels()
			if len(tt.updatedLabels) != len(updatedLabels) {
				t.Errorf("labeler.UpdatedLabels() expected updated labels = %v (len: %d), got %v (len: %d)",
					tt.updatedLabels, len(tt.updatedLabels), updatedLabels, len(updatedLabels))
			}
			for expectedLabelKey, expectedLabelValue := range tt.updatedLabels {
				value, ok := updatedLabels[expectedLabelKey]
				if !ok {
					t.Errorf("labeler expected updated label %q but did not find it", expectedLabelKey)
				} else if expectedLabelValue != value {
					t.Errorf("labeler expected updated label %q with value %q, got %q instead", expectedLabelKey, expectedLabelValue, value)
				}
			}

			addedLabelsMap := addedLabels(tt.args.node)
			var addedLabels []string
			for addedLabel := range addedLabelsMap {
				addedLabels = append(addedLabels, addedLabel)
			}
			sort.Strings(addedLabels)
			addedLabelsAnnotation := strings.Join(addedLabels, internalConsts.NodeAnnotationValueSeparator)
			if tt.expectedAddedLabelsAnnotation != addedLabelsAnnotation {
				t.Errorf("labeler expected added labels annotation = %v, got %v", tt.expectedAddedLabelsAnnotation, addedLabelsAnnotation)
			}
		})
	}
}
