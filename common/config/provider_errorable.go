/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

This is a provider proxy that allows you to force errors for the methods.  Obviously, this is for testing.
*/
package config

import (
	"errors"
	"gitlab.com/ginfra/ginfra/base"
)

// #####################################################################################################################
// # FILE CONFIG PROVIDER

type configProviderErrorable struct {
	errorMethods map[string]string
	cp           ConfigProvider
}

func (cpe *configProviderErrorable) SetError(method, err string) {
	cpe.errorMethods[method] = err
}

func (cpe *configProviderErrorable) IsConfigurable() bool {
	return true
}

func (cpe *configProviderErrorable) SetConfigAccessAuth(auth *string) {
	cpe.cp.SetConfigAccessAuth(auth)
}

func (cpe *configProviderErrorable) FetchConfig(path ...string) *Config {
	if e, ok := cpe.errorMethods["FetchConfig"]; ok {
		return NewConfig(nil, errors.New(e))
	}
	return cpe.cp.FetchConfig(path...)
}

func (cpe *configProviderErrorable) PostConfig(config *Config, path ...string) error {
	if e, ok := cpe.errorMethods["PostConfig"]; ok {
		return base.NewGinfraError(e)
	}
	return cpe.cp.PostConfig(config, path...)
}

// RemoveConfig will remove a configuration point and all children.
func (cpe *configProviderErrorable) RemoveConfig(path ...string) error {
	if e, ok := cpe.errorMethods["RemoveConfig"]; ok {
		return base.NewGinfraError(e)
	}
	return cpe.cp.RemoveConfig(path...)
}

func (cpe *configProviderErrorable) RegisterService(itype InstanceType, name, deployment, auth, image, configProvider,
	serviceClass string, port int, sspc map[string]interface{}) (string, error) {
	if e, ok := cpe.errorMethods["RegisterService"]; ok {
		return "", base.NewGinfraError(e)
	}
	return cpe.cp.RegisterService(itype, name, deployment, auth, image, configProvider, serviceClass, port, sspc)
}

func (cpe *configProviderErrorable) RegisterInstance(itype InstanceType, name, deployment, auth, id, ip, fqdn string) (string, error) {
	if e, ok := cpe.errorMethods["RegisterInstance"]; ok {
		return "", base.NewGinfraError(e)
	}
	return cpe.cp.RegisterInstance(itype, name, deployment, auth, id, ip, fqdn)
}

func (cpe *configProviderErrorable) PortalAdd(deployment, serviceClass, location string) (string, error) {
	if e, ok := cpe.errorMethods["PortalAdd"]; ok {
		return "", base.NewGinfraError(e)
	}
	return cpe.cp.PortalAdd(deployment, serviceClass, location)
}

func (cpe *configProviderErrorable) PortalRemove(deployment, serviceClass string) error {
	if e, ok := cpe.errorMethods["PortalRemove"]; ok {
		return base.NewGinfraError(e)
	}
	return cpe.cp.PortalRemove(deployment, serviceClass)
}

func (cpe *configProviderErrorable) DeregisterService(deployment, name string) error {
	if e, ok := cpe.errorMethods["DeregisterService"]; ok {
		return base.NewGinfraError(e)
	}
	return cpe.cp.DeregisterService(deployment, name)
}

func (cpe *configProviderErrorable) DiscoveryAdd(deployment, serviceClass, uri string) error {
	if e, ok := cpe.errorMethods["DiscoveryAdd"]; ok {
		return base.NewGinfraError(e)
	}
	return cpe.cp.DiscoveryAdd(deployment, serviceClass, uri)
}

func (cpe *configProviderErrorable) DiscoverService(deployment, serviceClass string) (string, error) {
	if e, ok := cpe.errorMethods["DiscoverService"]; ok {
		return "", base.NewGinfraError(e)
	}
	return cpe.cp.DiscoverService(deployment, serviceClass)
}

func (cpe *configProviderErrorable) DiscoverPortal(deployment, serviceClass string) (string, error) {
	if e, ok := cpe.errorMethods["DiscoverPortal"]; ok {
		return "", base.NewGinfraError(e)
	}
	return cpe.cp.DiscoverPortal(deployment, serviceClass)
}

func (cpe *configProviderErrorable) GetInitState(deployment, name, id string) (InitState, error) {
	if e, ok := cpe.errorMethods["GetInitState"]; ok {
		return InitPending, base.NewGinfraError(e)
	}
	return cpe.cp.GetInitState(deployment, name, id)
}

func (cpe *configProviderErrorable) GetState(deployment, name string) (map[string]interface{}, error) {
	if e, ok := cpe.errorMethods["GetState"]; ok {
		return nil, base.NewGinfraError(e)
	}
	return cpe.cp.GetState(deployment, name)
}

func (cpe *configProviderErrorable) PostState(deployment, name, id string, state map[string]interface{}) error {
	if e, ok := cpe.errorMethods["PostState"]; ok {
		return base.NewGinfraError(e)
	}
	return cpe.cp.PostState(deployment, name, id, state)
}

func (cpe *configProviderErrorable) StorePut(name, deployment, service string, data []byte) error {
	if e, ok := cpe.errorMethods["StorePut"]; ok {
		return base.NewGinfraError(e)
	}
	return cpe.cp.StorePut(deployment, name, service, data)
}

func (cpe *configProviderErrorable) StoreGet(name, deployment, service string, raw bool) ([]byte, error) {
	if e, ok := cpe.errorMethods["StoreGet"]; ok {
		return []byte{}, base.NewGinfraError(e)
	}
	return cpe.cp.StoreGet(name, deployment, service, raw)
}

// ClearStore clears the entire store, removing all files.
func (cpe *configProviderErrorable) ClearStore() error {
	if e, ok := cpe.errorMethods["ClearStore"]; ok {
		return base.NewGinfraError(e)
	}
	return cpe.cp.ClearStore()
}

func (cpe *configProviderErrorable) PutValue(deployment, name, value string) error {
	if e, ok := cpe.errorMethods["PutValue"]; ok {
		return base.NewGinfraError(e)
	}
	return cpe.cp.PutValue(deployment, name, value)
}

func (cpe *configProviderErrorable) GetValue(deployment, service, name string) *ConfigValue {
	if e, ok := cpe.errorMethods["GetValue"]; ok {
		return &ConfigValue{err: base.NewGinfraError(e)}
	}
	return cpe.cp.GetValue(deployment, service, name)
}
