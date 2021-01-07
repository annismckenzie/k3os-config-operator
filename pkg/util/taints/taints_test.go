/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package taints

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestAddOrUpdateTaint(t *testing.T) {
	node := &corev1.Node{}

	taint := &corev1.Taint{
		Key:    "foo",
		Value:  "bar",
		Effect: corev1.TaintEffectNoSchedule,
	}

	checkResult := func(testCaseName string, newNode *corev1.Node, expectedTaint *corev1.Taint, result, expectedResult bool, err error) {
		if err != nil {
			t.Errorf("[%s] should not raise error but got %v", testCaseName, err)
		}
		if result != expectedResult {
			t.Errorf("[%s] should return %t, but got: %t", testCaseName, expectedResult, result)
		}
		if len(newNode.Spec.Taints) != 1 || !reflect.DeepEqual(newNode.Spec.Taints[0], *expectedTaint) {
			t.Errorf("[%s] node should only have one taint: %v, but got: %v", testCaseName, *expectedTaint, newNode.Spec.Taints)
		}
	}

	// Add a new Taint.
	newNode, result, err := AddOrUpdateTaint(node, taint)
	checkResult("Add New Taint", newNode, taint, result, true, err)

	// Update a Taint.
	taint.Value = "bar_1"
	newNode, result, err = AddOrUpdateTaint(node, taint)
	checkResult("Update Taint", newNode, taint, result, true, err)

	// Add a duplicate Taint.
	node = newNode
	newNode, result, err = AddOrUpdateTaint(node, taint)
	checkResult("Add Duplicate Taint", newNode, taint, result, false, err)
}

func TestTaintExists(t *testing.T) {
	testingTaints := []corev1.Taint{
		{
			Key:    "foo_1",
			Value:  "bar_1",
			Effect: corev1.TaintEffectNoExecute,
		},
		{
			Key:    "foo_2",
			Value:  "bar_2",
			Effect: corev1.TaintEffectNoSchedule,
		},
	}

	cases := []struct {
		name           string
		taintToFind    *corev1.Taint
		expectedResult bool
	}{
		{
			name:           "taint exists",
			taintToFind:    &corev1.Taint{Key: "foo_1", Value: "bar_1", Effect: corev1.TaintEffectNoExecute},
			expectedResult: true,
		},
		{
			name:           "different key",
			taintToFind:    &corev1.Taint{Key: "no_such_key", Value: "bar_1", Effect: corev1.TaintEffectNoExecute},
			expectedResult: false,
		},
		{
			name:           "different effect",
			taintToFind:    &corev1.Taint{Key: "foo_1", Value: "bar_1", Effect: corev1.TaintEffectNoSchedule},
			expectedResult: false,
		},
	}

	for _, c := range cases {
		result := TaintExists(testingTaints, c.taintToFind)

		if result != c.expectedResult {
			t.Errorf("[%s] unexpected results: %v", c.name, result)
			continue
		}
	}
}

func TestRemoveTaint(t *testing.T) {
	cases := []struct {
		name           string
		node           *corev1.Node
		taintToRemove  *corev1.Taint
		expectedTaints []corev1.Taint
		expectedResult bool
	}{
		{
			name: "remove taint unsuccessfully",
			node: &corev1.Node{
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "foo",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			taintToRemove: &corev1.Taint{
				Key:    "foo_1",
				Effect: corev1.TaintEffectNoSchedule,
			},
			expectedTaints: []corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedResult: false,
		},
		{
			name: "remove taint successfully",
			node: &corev1.Node{
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{
						{
							Key:    "foo",
							Effect: corev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			taintToRemove: &corev1.Taint{
				Key:    "foo",
				Effect: corev1.TaintEffectNoSchedule,
			},
			expectedTaints: []corev1.Taint{},
			expectedResult: true,
		},
		{
			name: "remove taint from node with no taint",
			node: &corev1.Node{
				Spec: corev1.NodeSpec{
					Taints: []corev1.Taint{},
				},
			},
			taintToRemove: &corev1.Taint{
				Key:    "foo",
				Effect: corev1.TaintEffectNoSchedule,
			},
			expectedTaints: []corev1.Taint{},
			expectedResult: false,
		},
	}

	for _, c := range cases {
		newNode, result, err := RemoveTaint(c.node, c.taintToRemove)
		if err != nil {
			t.Errorf("[%s] should not raise error but got: %v", c.name, err)
		}
		if result != c.expectedResult {
			t.Errorf("[%s] should return %t, but got: %t", c.name, c.expectedResult, result)
		}
		if !reflect.DeepEqual(newNode.Spec.Taints, c.expectedTaints) {
			t.Errorf("[%s] the new node object should have taints %v, but got: %v", c.name, c.expectedTaints, newNode.Spec.Taints)
		}
	}
}

func TestDeleteTaint(t *testing.T) {
	cases := []struct {
		name           string
		taints         []corev1.Taint
		taintToDelete  *corev1.Taint
		expectedTaints []corev1.Taint
		expectedResult bool
	}{
		{
			name: "delete taint with different name",
			taints: []corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			taintToDelete: &corev1.Taint{Key: "foo_1", Effect: corev1.TaintEffectNoSchedule},
			expectedTaints: []corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedResult: false,
		},
		{
			name: "delete taint with different effect",
			taints: []corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			taintToDelete: &corev1.Taint{Key: "foo", Effect: corev1.TaintEffectNoExecute},
			expectedTaints: []corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedResult: false,
		},
		{
			name: "delete taint successfully",
			taints: []corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			taintToDelete:  &corev1.Taint{Key: "foo", Effect: corev1.TaintEffectNoSchedule},
			expectedTaints: []corev1.Taint{},
			expectedResult: true,
		},
		{
			name:           "delete taint from empty taint array",
			taints:         []corev1.Taint{},
			taintToDelete:  &corev1.Taint{Key: "foo", Effect: corev1.TaintEffectNoSchedule},
			expectedTaints: []corev1.Taint{},
			expectedResult: false,
		},
	}

	for _, c := range cases {
		taints, result := DeleteTaint(c.taints, c.taintToDelete)
		if result != c.expectedResult {
			t.Errorf("[%s] should return %t, but got: %t", c.name, c.expectedResult, result)
		}
		if !reflect.DeepEqual(taints, c.expectedTaints) {
			t.Errorf("[%s] the result taints should be %v, but got: %v", c.name, c.expectedTaints, taints)
		}
	}
}

func TestDeleteTaintByKey(t *testing.T) {
	cases := []struct {
		name           string
		taints         []corev1.Taint
		taintKey       string
		expectedTaints []corev1.Taint
		expectedResult bool
	}{
		{
			name: "delete taint unsuccessfully",
			taints: []corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			taintKey: "foo_1",
			expectedTaints: []corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedResult: false,
		},
		{
			name: "delete taint successfully",
			taints: []corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			taintKey:       "foo",
			expectedTaints: []corev1.Taint{},
			expectedResult: true,
		},
		{
			name:           "delete taint from empty taint array",
			taints:         []corev1.Taint{},
			taintKey:       "foo",
			expectedTaints: []corev1.Taint{},
			expectedResult: false,
		},
	}

	for _, c := range cases {
		taints, result := DeleteTaintsByKey(c.taints, c.taintKey)
		if result != c.expectedResult {
			t.Errorf("[%s] should return %t, but got: %t", c.name, c.expectedResult, result)
		}
		if !reflect.DeepEqual(c.expectedTaints, taints) {
			t.Errorf("[%s] the result taints should be %v, but got: %v", c.name, c.expectedTaints, taints)
		}
	}
}

func TestCheckIfTaintsAlreadyExists(t *testing.T) {
	oldTaints := []corev1.Taint{
		{
			Key:    "foo_1",
			Value:  "bar",
			Effect: corev1.TaintEffectNoSchedule,
		},
		{
			Key:    "foo_2",
			Value:  "bar",
			Effect: corev1.TaintEffectNoSchedule,
		},
		{
			Key:    "foo_3",
			Value:  "bar",
			Effect: corev1.TaintEffectNoSchedule,
		},
	}

	cases := []struct {
		name           string
		taintsToCheck  []corev1.Taint
		expectedResult string
	}{
		{
			name:           "empty array",
			taintsToCheck:  []corev1.Taint{},
			expectedResult: "",
		},
		{
			name: "no match",
			taintsToCheck: []corev1.Taint{
				{
					Key:    "foo_1",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			expectedResult: "",
		},
		{
			name: "match one taint",
			taintsToCheck: []corev1.Taint{
				{
					Key:    "foo_2",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedResult: "foo_2",
		},
		{
			name: "match two taints",
			taintsToCheck: []corev1.Taint{
				{
					Key:    "foo_2",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "foo_3",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedResult: "foo_2,foo_3",
		},
	}

	for _, c := range cases {
		result := CheckIfTaintsAlreadyExists(oldTaints, c.taintsToCheck)
		if result != c.expectedResult {
			t.Errorf("[%s] should return '%s', but got: '%s'", c.name, c.expectedResult, result)
		}
	}
}

func TestReorganizeTaints(t *testing.T) {
	node := &corev1.Node{
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				{
					Key:    "foo",
					Value:  "bar",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	cases := []struct {
		name              string
		overwrite         bool
		taintsToAdd       []corev1.Taint
		taintsToDelete    []corev1.Taint
		expectedTaints    []corev1.Taint
		expectedOperation string
		expectedErr       bool
	}{
		{
			name:              "no changes with overwrite is true",
			overwrite:         true,
			taintsToAdd:       []corev1.Taint{},
			taintsToDelete:    []corev1.Taint{},
			expectedTaints:    node.Spec.Taints,
			expectedOperation: MODIFIED,
			expectedErr:       false,
		},
		{
			name:              "no changes with overwrite is false",
			overwrite:         false,
			taintsToAdd:       []corev1.Taint{},
			taintsToDelete:    []corev1.Taint{},
			expectedTaints:    node.Spec.Taints,
			expectedOperation: UNTAINTED,
			expectedErr:       false,
		},
		{
			name:      "add new taint",
			overwrite: false,
			taintsToAdd: []corev1.Taint{
				{
					Key:    "foo_1",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
			taintsToDelete:    []corev1.Taint{},
			expectedTaints:    append([]corev1.Taint{{Key: "foo_1", Effect: corev1.TaintEffectNoExecute}}, node.Spec.Taints...),
			expectedOperation: TAINTED,
			expectedErr:       false,
		},
		{
			name:        "delete taint with effect",
			overwrite:   false,
			taintsToAdd: []corev1.Taint{},
			taintsToDelete: []corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedTaints:    []corev1.Taint{},
			expectedOperation: UNTAINTED,
			expectedErr:       false,
		},
		{
			name:        "delete taint with no effect",
			overwrite:   false,
			taintsToAdd: []corev1.Taint{},
			taintsToDelete: []corev1.Taint{
				{
					Key: "foo",
				},
			},
			expectedTaints:    []corev1.Taint{},
			expectedOperation: UNTAINTED,
			expectedErr:       false,
		},
		{
			name:        "delete non-exist taint",
			overwrite:   false,
			taintsToAdd: []corev1.Taint{},
			taintsToDelete: []corev1.Taint{
				{
					Key:    "foo_1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedTaints:    node.Spec.Taints,
			expectedOperation: UNTAINTED,
			expectedErr:       true,
		},
		{
			name:      "add new taint and delete old one",
			overwrite: false,
			taintsToAdd: []corev1.Taint{
				{
					Key:    "foo_1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			taintsToDelete: []corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedTaints: []corev1.Taint{
				{
					Key:    "foo_1",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedOperation: MODIFIED,
			expectedErr:       false,
		},
	}

	for _, c := range cases {
		operation, taints, err := ReorganizeTaints(node, c.overwrite, c.taintsToAdd, c.taintsToDelete)
		if c.expectedErr && err == nil {
			t.Errorf("[%s] expect to see an error, but did not get one", c.name)
		} else if !c.expectedErr && err != nil {
			t.Errorf("[%s] expect not to see an error, but got one: %v", c.name, err)
		}

		if !reflect.DeepEqual(c.expectedTaints, taints) {
			t.Errorf("[%s] expect to see taint list %#v, but got: %#v", c.name, c.expectedTaints, taints)
		}

		if c.expectedOperation != operation {
			t.Errorf("[%s] expect to see operation %s, but got: %s", c.name, c.expectedOperation, operation)
		}
	}
}

func TestParseTaints(t *testing.T) {
	cases := []struct {
		name                   string
		spec                   []string
		expectedTaints         []corev1.Taint
		expectedTaintsToRemove []corev1.Taint
		expectedErr            bool
	}{
		{
			name:        "invalid spec format",
			spec:        []string{""},
			expectedErr: true,
		},
		{
			name:        "invalid spec format",
			spec:        []string{"foo=abc"},
			expectedErr: true,
		},
		{
			name:        "invalid spec format",
			spec:        []string{"foo=abc=xyz:NoSchedule"},
			expectedErr: true,
		},
		{
			name:        "invalid spec format",
			spec:        []string{"foo=abc:xyz:NoSchedule"},
			expectedErr: true,
		},
		{
			name:        "invalid spec format for adding taint",
			spec:        []string{"foo"},
			expectedErr: true,
		},
		{
			name:        "invalid spec effect for adding taint",
			spec:        []string{"foo=abc:invalid_effect"},
			expectedErr: true,
		},
		{
			name:        "invalid spec effect for deleting taint",
			spec:        []string{"foo:invalid_effect-"},
			expectedErr: true,
		},
		{
			name: "add new taints",
			spec: []string{"foo=abc:NoSchedule", "bar=abc:NoSchedule", "baz:NoSchedule", "qux:NoSchedule", "foobar=:NoSchedule"},
			expectedTaints: []corev1.Taint{
				{
					Key:    "foo",
					Value:  "abc",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "bar",
					Value:  "abc",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "baz",
					Value:  "",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "qux",
					Value:  "",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "foobar",
					Value:  "",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedErr: false,
		},
		{
			name: "delete taints",
			spec: []string{"foo:NoSchedule-", "bar:NoSchedule-", "qux=:NoSchedule-", "dedicated-"},
			expectedTaintsToRemove: []corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "bar",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "qux",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key: "dedicated",
				},
			},
			expectedErr: false,
		},
		{
			name: "add taints and delete taints",
			spec: []string{"foo=abc:NoSchedule", "bar=abc:NoSchedule", "baz:NoSchedule", "qux:NoSchedule", "foobar=:NoSchedule", "foo:NoSchedule-", "bar:NoSchedule-", "baz=:NoSchedule-"},
			expectedTaints: []corev1.Taint{
				{
					Key:    "foo",
					Value:  "abc",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "bar",
					Value:  "abc",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "baz",
					Value:  "",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "qux",
					Value:  "",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "foobar",
					Value:  "",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedTaintsToRemove: []corev1.Taint{
				{
					Key:    "foo",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "bar",
					Effect: corev1.TaintEffectNoSchedule,
				},
				{
					Key:    "baz",
					Value:  "",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
			expectedErr: false,
		},
	}

	for _, c := range cases {
		taints, taintsToRemove, err := ParseTaints(c.spec)
		if c.expectedErr && err == nil {
			t.Errorf("[%s] expected error for spec %s, but got nothing", c.name, c.spec)
		}
		if !c.expectedErr && err != nil {
			t.Errorf("[%s] expected no error for spec %s, but got: %v", c.name, c.spec, err)
		}
		if !reflect.DeepEqual(c.expectedTaints, taints) {
			t.Errorf("[%s] expected returen taints as %v, but got: %v", c.name, c.expectedTaints, taints)
		}
		if !reflect.DeepEqual(c.expectedTaintsToRemove, taintsToRemove) {
			t.Errorf("[%s] expected return taints to be removed as %v, but got: %v", c.name, c.expectedTaintsToRemove, taintsToRemove)
		}
	}
}
