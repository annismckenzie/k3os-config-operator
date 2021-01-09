package config

import (
	"os"

	"github.com/annismckenzie/k3os-config-operator/pkg/consts"
)

// EnableDevMode returns whether dev mode should be enabled.
func EnableDevMode() bool {
	return os.Getenv(consts.DevModeEnvName) == "true"
}
