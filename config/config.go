package config

import (
	"os"

	flags "github.com/jessevdk/go-flags"
)

// InitializeConfiguration initializes the configuration from CLI flags and the environment.
func InitializeConfiguration(parseOptions ...flags.Options) (*Configuration, error) {
	config := &Configuration{}
	var options flags.Options = flags.Default
	if len(parseOptions) > 0 {
		options = 0
		for _, option := range parseOptions {
			options |= option
		}
	}
	return config, parseWithOptions(config, options)
}

func parseWithOptions(c interface{}, parseOptions flags.Options) error {
	_, err := flags.NewParser(c, parseOptions).Parse()
	if e, ok := err.(*flags.Error); ok {
		// catch help message
		if e.Type == flags.ErrHelp {
			os.Exit(0)
			return nil
		}
	}
	return err
}

// Configuration contains all config values and is initialized via InitializeConfiguration.
type Configuration struct {
	MetricsAddr string `long:"metrics-addr" default:":8080" env:"METRICS_ADDR" description:"The address the metric endpoint binds to."`

	Namespace string `long:"namespace" required:"true" env:"NAMESPACE" description:"The namespace the operator is running in."`
	NodeName  string `long:"node-name" required:"true" env:"NODE_NAME" description:"The name of the node the operator is running on."`
	DevMode   bool   `long:"dev-mode"                  env:"DEV_MODE"  description:"Enable dev mode."`

	ManageNodeConfigFile   bool   `long:"manage-node-config-file"                        env:"ENABLE_NODECONFIG_FILE_MANAGEMENT" description:"Enable node config file management."`
	NodeConfigFileLocation string `long:"node-config-file-location"                      env:"NODECONFIG_FILE_LOCATION"          description:"Location of the node config file on disk."`
	NodeConfigSecretName   string `long:"node-config-secret-name"   default:"k3os-nodes" env:"NODECONFIG_SECRET_NAME"            description:"Name of the secret that contains the node configurations."`
}

// EnableDevMode returns whether dev mode should be enabled or not.
func (c *Configuration) EnableDevMode() bool {
	return c.DevMode
}

// EnableNodeConfigFileManagement returns whether the node config file management should be enabled or not.
func (c *Configuration) EnableNodeConfigFileManagement() bool {
	return c.ManageNodeConfigFile
}
