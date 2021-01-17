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
	"fmt"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// K3OSConfigFileKind contains the Kind of the K3OSConfigFile CR.
const K3OSConfigFileKind = "K3OSConfigFile"

// K3OSConfigFileSectionK3OS contains the spec of the `k3os` section of
// the K3OS YAML config file.
type K3OSConfigFileSectionK3OS struct {
	Labels map[string]string `json:"labels" yaml:"labels"`
	Taints []string          `json:"taints" yaml:"taints"`
}

// K3OSConfigFileSpec defines the desired state of K3OSConfigFile.
// Use `ParseConfigYAML()` to parse a k3OS config.yaml file.
type K3OSConfigFileSpec struct {
	Hostname string `json:"hostname" yaml:"hostname"`

	K3OS K3OSConfigFileSectionK3OS `json:"k3os" yaml:"k3os"`

	// Data contains the raw contents of the config file.
	Data []byte `json:"-" yaml:"-"`
}

// Validate checks the contents of the spec for errors and returns them.
func (s *K3OSConfigFileSpec) Validate() error {
	return nil
}

// ParseConfigYAML parses the data of a k3OS config.yaml into a K3OSConfigFileSpec object.
func ParseConfigYAML(data []byte) (*K3OSConfigFileSpec, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("cannot parse data into YAML node config because data is empty")
	}
	c := &K3OSConfigFileSpec{Data: data}
	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML node config data: %w", err)
	}

	// validate K3OSConfigFileSpec before returning
	if err := c.Validate(); err != nil {
		return c, err
	}

	return c, nil
}

// K3OSConfigFileStatus defines the observed state of K3OSConfigFile.
type K3OSConfigFileStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// K3OSConfigFile is the Schema for the k3osconfigfiles API.
type K3OSConfigFile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   K3OSConfigFileSpec   `json:"spec,omitempty"`
	Status K3OSConfigFileStatus `json:"status,omitempty"`
}

// MarshalYAML returns the YAML representation of the config file.
func (s *K3OSConfigFile) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(s.Spec)
}

// +kubebuilder:object:root=true

// K3OSConfigFileList contains a list of K3OSConfigFile.
type K3OSConfigFileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K3OSConfigFile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&K3OSConfigFile{}, &K3OSConfigFileList{})
}
