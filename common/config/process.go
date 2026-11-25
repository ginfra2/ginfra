/*
Package config
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Service and tool configuration manipulation.
*/
package config

import (
	"fmt"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"os"
	"strings"
)

// #####################################################################################################################
// # TYPES AND VALUES

type ProcessConfig struct {
	Deployment string
	Service    string
}

const ConfigPathPartSeparator = "."

const ConfigPathPrefix = "$"
const ConfigDiscoverServicePrefix = "="
const ConfigDiscoverPortalPrefix = "*"
const ConfigFetchServiceMngAuth = "+"
const ConfigEnvVariable = "^"
const ConfigServicePort = "&"

const ConfigPointPrefix = "%"
const ConfigPointDeployment = "DEPLOYMENT"
const ConfigPointService = "SERVICE"

// #####################################################################################################################
// # Func

func prepareConfigPath(provider ConfigProvider, cfg *ProcessConfig, target string) ([]string, error) {
	var err error
	path := strings.Split(target, ConfigPathPartSeparator)
	for i, v := range path {
		if path[i], err = processConfigDirective(provider, cfg, v); err != nil {
			return []string{}, err
		}
	}
	return path, err
}

func processConfigDirective(provider ConfigProvider, cfg *ProcessConfig, target string) (string, error) {
	var (
		r   = target
		err error
	)

	if strings.HasPrefix(target, ConfigPointPrefix) {
		switch strings.ToUpper(target[1:]) {
		case ConfigPointDeployment:
			r = cfg.Deployment
		case ConfigPointService:
			r = cfg.Service
		default:
			err = base.NewGinfraErrorA("Unknown configuration point directive.", base.LM_NAME, target)
		}
	}

	return r, err
}

func processConfigPartPath(provider ConfigProvider, cfg *ProcessConfig, target string) (string, error) {
	var (
		r    string
		path []string
		err  error
	)

	if path, err = prepareConfigPath(provider, cfg, target); err != nil {
		return "", err
	}

	if len(path) > 0 {
		c := provider.FetchConfig(path...)
		if c.Err == nil {
			r = c.GetValue(path[len(path)-1]).GetValueStringNoerr()

		} else {
			err = c.Err
		}
	}

	return r, err
}

func processConfigPartDiscoverService(provider ConfigProvider, cfg *ProcessConfig, target string) (string, error) {
	var (
		r    string
		err  error
		path []string
	)

	if path, err = prepareConfigPath(provider, cfg, target); err != nil {
		return "", err
	}

	if len(path) == 2 {
		r, err = provider.DiscoverService(path[0], path[1])

	} else {
		err = base.NewGinfraErrorA("Poorly formed Discover Service point.")
	}

	return r, err
}

func processConfigPartDiscoverPortal(provider ConfigProvider, cfg *ProcessConfig, target string) (string, error) {
	var (
		r    string
		err  error
		path []string
	)

	if path, err = prepareConfigPath(provider, cfg, target); err != nil {
		return "", err
	}

	if len(path) == 2 {
		r, err = provider.DiscoverPortal(path[0], path[1])

	} else {
		err = base.NewGinfraErrorA("Poorly formed Discover Portal point.")
	}

	return r, err
}

func processConfigPartFetchMngAuth(provider ConfigProvider, cfg *ProcessConfig, target string) (string, error) {
	var (
		r    string
		err  error
		path []string
	)

	if path, err = prepareConfigPath(provider, cfg, target); err != nil {
		return "", err
	}

	if len(path) == 2 {
		c := provider.FetchConfig(CR_DEPLOYMENTS, path[0], CRD_SERVICES, path[1], CRD_SERVICES__SERVICE, CRD_SERVICES__SERVICE_AUTH)
		err = c.Err
		if err == nil {
			r = c.GetValue(CRD_SERVICES__SERVICE_AUTH).GetValueStringNoerr()
		}

	} else {
		return "", base.NewGinfraErrorA("Poorly formed Fetch Management Auth Token point.")
	}

	return r, err
}

func processConfigPartEnvVar(target string) string {
	return os.Getenv(target)

}

func processConfigPartServicePort(provider ConfigProvider, cfg *ProcessConfig, target string) (string, error) {
	var (
		r    string
		err  error
		path []string
	)

	if path, err = prepareConfigPath(provider, cfg, target); err != nil {
		return "", err
	}

	if len(path) == 2 {
		c := provider.FetchConfig(CR_DEPLOYMENTS, path[0], CRD_SERVICES, path[1], CRD_SERVICES__SERVICE, CRD_SSERVICE__CONFIG, CRD_SSERVICE__CONFIG__PORT)
		err = c.Err
		if err == nil {
			r = c.GetValue(CRD_SSERVICE__CONFIG__PORT).GetValueStringNoerr()
		}

	} else {
		return "", base.NewGinfraErrorA("Poorly formed Fetch Service Port point.")
	}

	return r, err
}

func processConfigPart(provider ConfigProvider, cfg *ProcessConfig, target string) (string, error) {
	var (
		r   string
		err error
	)

	// TODO this has gotten large enough to want a better parser.
	if strings.HasPrefix(target, ConfigPathPrefix) {
		r, err = processConfigPartPath(provider, cfg, target[1:])
	} else if strings.HasPrefix(target, ConfigDiscoverServicePrefix) {
		r, err = processConfigPartDiscoverService(provider, cfg, target[1:])
	} else if strings.HasPrefix(target, ConfigDiscoverPortalPrefix) {
		r, err = processConfigPartDiscoverPortal(provider, cfg, target[1:])
	} else if strings.HasPrefix(target, ConfigFetchServiceMngAuth) {
		r, err = processConfigPartFetchMngAuth(provider, cfg, target[1:])
	} else if strings.HasPrefix(target, ConfigEnvVariable) {
		r = processConfigPartEnvVar(target[1:])
	} else if strings.HasPrefix(target, ConfigServicePort) {
		r, err = processConfigPartServicePort(provider, cfg, target[1:])
	} else {
		err = base.NewGinfraErrorAFlags("Unknown directive.", base.ERRFLAG_RECOVERABLE, base.LM_TARGET, target)
	}

	return r, err
}

// ProcessConfigString replaces placeholders in the target string using configuration data from the given ConfigProvider.
// The function handles substitutions using a custom logic and reports any errors encountered during processing.
// Returns the processed string and any errors that occurred during the operation.
func ProcessConfigString(provider ConfigProvider, config ProcessConfig, target string) (string, error) {

	var errs []error

	s, err := common.SimpleReplacerHandler(target, func(in string) string {
		ss, e := processConfigPart(provider, &config, in)
		if e != nil {
			errs = append(errs, e)
		}
		return ss
	}, false)

	if err != nil {
		errs = append(errs, err)
	}

	// Let's see how this looks.
	if len(errs) > 0 {
		err = fmt.Errorf("%d errors occurred during processing: %v", len(errs), errs)
	}

	return s, err
}

// ProcessVarConfigString processes a configuration string, replacing variables with their resolved values from a map or provider.
// It supports error handling, including recovery for recoverable errors, and returns the processed string or an error if any.
func ProcessVarConfigString(provider ConfigProvider, config ProcessConfig, v map[string]interface{}, target string) (string,
	error) {

	var errs []error

	s, err := common.SimpleReplacerHandler(target, func(in string) string {
		ss, e := processConfigPart(provider, &config, in)
		if e != nil {
			if (base.GetErrorFlags(e) & base.ERRFLAG_RECOVERABLE) > 0 {
				if val, ok := v[in]; ok {
					ss = fmt.Sprintf("%v", val)
				}
			} else {
				errs = append(errs, e)
			}
		}
		return ss
	}, false)

	if err != nil {
		errs = append(errs, err)
	}

	// Let's see how this looks.
	if len(errs) > 0 {
		err = fmt.Errorf("errors occurred during processing: %v", errs)
	}

	return s, err
}
