package nodes

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// NodeConfig is the k3OS config.yaml file.
type NodeConfig struct {
	Hostname string `yaml:"hostname"`

	K3OS struct {
		Labels map[string]string `yaml:"labels"`
		Taints []string          `yaml:"taints"`
	} `yaml:"k3os"`
}

// ParseNodeConfig parses the data into a NodeConfig object.
func ParseNodeConfig(data []byte) (*NodeConfig, error) {
	nc := &NodeConfig{}
	if err := yaml.Unmarshal(data, nc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML node config data: %w", err)
	}
	return nc, nil
}
