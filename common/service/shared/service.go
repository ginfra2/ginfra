/*
Package shared
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Shared infrastructure for services.
*/
package shared

import (
	"errors"
	"fmt"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"os"
	"path"
	"path/filepath"
)

// LoadSpecifics reads the contents of the "specifics.json" file located in the specified home directory.
// If the file does not exist, it falls back to reading the root directory ("/").
// It returns the contents of the file as a string. If an error occurs, it panics with a descriptive message.
func LoadSpecifics(home string) string {
	b, err := os.ReadFile(filepath.Join(home, base.SpecificsFileName))
	if err != nil && errors.Is(err, os.ErrNotExist) {
		b, err = os.ReadFile(filepath.Join("/", base.SpecificsFileName))
	}

	if err != nil {
		panic(fmt.Sprintf("Could not load specifics.  %v", err))
	}
	return string(b)
}

// ServiceConfig is a structure representing the configuration for a specific service instance.
type ServiceConfig struct {

	// Flavors
	InstanceType config.InstanceType

	// Values
	Auth             string // Service authorization id.
	Name             string
	DeploymentName   string
	ConfigUri        string
	ConfigUriPrimary string
	ConfigToken      string
	Id               string // Specific instance ID.
	Clear            bool   // Clear data stores and start fresh.

	// Not loaded by this package.
	ServiceHome string
}

func failLoadServiceConfig(err error) error {
	return base.NewGinfraErrorChildA("Failed load service configuration from file.  Make sure file is present and not corrupt.", err,
		base.LM_PATH, base.ConfigFileName)
}

// LoadServiceConfig loads the service configuration from a JSON file located at the given home directory.
// It returns a pointer to a ServiceConfig struct and an error if the configuration cannot be loaded or parsed.
func LoadServiceConfig(home string) (*ServiceConfig, error) {
	var (
		err error
		cfg ServiceConfig
		ok  bool
		val string
	)

	m := make(map[string]interface{})
	err = common.ReadJsonFileMap(path.Join(home, base.ConfigFileName), "service_config.json configuration file", &m)
	if err != nil {
		// to capture config from initcontainers.
		errr := common.ReadJsonFileMap(path.Join(base.ConfigMountDir, base.ConfigFileName), "service_config.json configuration file", &m)
		if errr != nil {
			return nil, failLoadServiceConfig(err)
		}
	}

	if _, ok = m[base.ConfigClear]; ok {
		cfg.Clear = true
	}

	if v, ok := m[base.ConfigAuth]; ok {
		if val, ok = v.(string); ok {
			cfg.Auth = val
		} else {
			return nil, failLoadServiceConfig(errors.New("bad type for " + base.ConfigAuth))
		}
	}

	if v, ok := m[base.ConfigUri]; ok {
		if val, ok = v.(string); ok {
			cfg.ConfigUri = val
		} else {
			return nil, failLoadServiceConfig(errors.New("bad type for " + base.ConfigUri))
		}
	}

	if v, ok := m[base.ConfigPortalUri]; ok {
		if val, ok = v.(string); ok {
			cfg.ConfigUriPrimary = val
		} else {
			return nil, failLoadServiceConfig(errors.New("bad type for " + base.ConfigPortalUri))
		}
	}

	if v, ok := m[base.ConfigAccessToken]; ok {
		if val, ok = v.(string); ok {
			cfg.ConfigToken = val
		} else {
			return nil, failLoadServiceConfig(errors.New("bad type for " + base.ConfigAccessToken))
		}
	}

	if v, ok := m[base.ConfigDeployment]; ok {
		if val, ok = v.(string); ok {
			cfg.DeploymentName = val
		} else {
			return nil, failLoadServiceConfig(errors.New("bad type for " + base.ConfigDeployment))
		}
	}

	if v, ok := m[base.ConfigName]; ok {
		if val, ok = v.(string); ok {
			cfg.Name = val
		} else {
			return nil, failLoadServiceConfig(errors.New("bad type for " + base.ConfigName))
		}
	}

	if v, ok := m[base.ConfigId]; ok {
		if val, ok = v.(string); ok {
			cfg.Id = val
		} else {
			return nil, failLoadServiceConfig(errors.New("bad type for " + base.ConfigId))
		}
	}

	if v, ok := m[base.ConfigType]; ok {
		if val, ok = v.(string); ok {
			if cfg.InstanceType, err = config.GetInstanceTypeFromString(val); err != nil {
				return nil, failLoadServiceConfig(base.NewGinfraErrorA("Bad instance type.",
					base.LM_TYPE, val))
			}
		} else {
			return nil, failLoadServiceConfig(errors.New("bad type for " + base.ConfigType))
		}
	}

	return &cfg, err
}
