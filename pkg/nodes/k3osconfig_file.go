package nodes

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	configv1alpha1 "github.com/annismckenzie/k3os-config-operator/apis/config/v1alpha1"
	"github.com/annismckenzie/k3os-config-operator/config"
	"github.com/annismckenzie/k3os-config-operator/pkg/errors"
)

// K3OSConfigFileUpdater handles updating the k3OS config file on disk.
type K3OSConfigFileUpdater interface {
	Update(*configv1alpha1.K3OSConfigFileSpec) error
}

// NewK3OSConfigFileUpdater returns an initialized K3OSConfigFileUpdater.
func NewK3OSConfigFileUpdater(configuration *config.Configuration) K3OSConfigFileUpdater {
	return &k3OSConfigFileUpdater{configuration: configuration}
}

type k3OSConfigFileUpdater struct {
	configuration *config.Configuration
}

func (u *k3OSConfigFileUpdater) enabled() bool {
	return u.configuration.EnableNodeConfigFileManagement()
}

// Update handles updating the k3OS config file on disk.
// It can be called anytime and will return errors.ErrSkipUpdate if
// the feature isn't enabled or the config file is already up to date.
func (u *k3OSConfigFileUpdater) Update(configFileSpec *configv1alpha1.K3OSConfigFileSpec) (err error) {
	if !u.enabled() {
		return errors.ErrSkipUpdate
	}

	if err = configFileSpec.Validate(); err != nil {
		return fmt.Errorf("the provided config file is invalid: %w", err)
	}

	// open the config file for read-write
	var configFile *os.File
	if configFile, err = os.OpenFile(u.configuration.NodeConfigFileLocation, os.O_RDWR, 0); err != nil {
		return fmt.Errorf("failed to open node config file: %w", err)
	}
	defer func() {
		// leaves the original error in place but still tries to close the file
		if errClose := configFile.Close(); err == nil {
			err = errClose
		}
	}()

	var configFileBytes []byte
	if configFileBytes, err = ioutil.ReadAll(configFile); err != nil {
		return
	}

	if bytes.Equal(configFileBytes, configFileSpec.Data) {
		return errors.ErrSkipUpdate
	}

	// truncate the file
	if err = configFile.Truncate(0); err != nil {
		return
	}
	// write the new configuration data
	_, err = configFile.WriteAt(configFileSpec.Data, 0)
	return err
}
