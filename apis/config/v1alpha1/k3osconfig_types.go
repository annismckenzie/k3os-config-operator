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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K3OSConfigKind contains the Kind of the K3OSConfig CR.
const K3OSConfigKind = "K3OSConfig"

// K3OSConfigListKind contains the Kind of a list of K3OSConfig CRs.
const K3OSConfigListKind = "K3OSConfigList"

// K3OSConfigSpec defines the desired state of K3OSConfig.
type K3OSConfigSpec struct {
	// SyncNodeLabels enables syncing node labels set in the K3OS config.yaml.
	// K3OS by default only sets labels on nodes on first boot.
	SyncNodeLabels bool `json:"syncNodeLabels,omitempty"`

	// SyncNodeTaints enables syncing node taints set in the K3OS config.yaml.
	// K3OS by default only sets taints on nodes on first boot.
	SyncNodeTaints bool `json:"syncNodeTaints,omitempty"`
}

// K3OSConfigStatus defines the observed state of K3OSConfig.
type K3OSConfigStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// K3OSConfig is the Schema for the k3osconfigs API.
type K3OSConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   K3OSConfigSpec   `json:"spec,omitempty"`
	Status K3OSConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// K3OSConfigList contains a list of K3OSConfig.
type K3OSConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K3OSConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&K3OSConfig{}, &K3OSConfigList{})
}
