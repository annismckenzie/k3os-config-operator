package nodes

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Config is the k3OS config.yaml file.
type Config struct {
	Hostname string `yaml:"hostname"`

	K3OS struct {
		Labels map[string]string `yaml:"labels"`
		Taints []string          `yaml:"taints"`
	} `yaml:"k3os"`
}

// ParseConfig parses the data into a Config object.
func ParseConfig(data []byte) (*Config, error) {
	nc := &Config{}
	if err := yaml.Unmarshal(data, nc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML node config data: %w", err)
	}
	return nc, nil
}
