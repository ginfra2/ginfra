/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Configuration providers.  There is a file based provider that is convenient for local testing.  There is an API
based provider that can us any service that implements the ginterfaces/service_giconfig.yaml.  Of course you can
always use ours--service_giconfig.
*/
package config

import (
	"fmt"
	"gitlab.com/ginfra/ginfra/base"
	"net/url"
	"strings"
)

// #####################################################################################################################
// # CONFIGURATION PROVIDER

// Operation These must match API tokens for the same operations, as you would find in service_giconfig.yaml.
type Operation string

const (
	OpPut    Operation = "put"
	OpGet              = "get"
	OpRemove           = "remove"
)

// InitState represents the initialization state of a service during its lifecycle.
type InitState int

const (
	InitNew = iota
	InitPending
	InitDone
)

// ConfigProvider is the common interface for configuration providers.  Use GetConfigProvider to get an instance of this.
type ConfigProvider interface {

	// IsConfigurable tells you if you can use this provider for configuration.  Strange, I know, but if the service
	// is a configuration provider, they do not participate in the same configuration as everything else.  As such they
	// should neither load, save nor modify configuration, nor should they register with or user another configuration
	// provider.
	IsConfigurable() bool

	// SetConfigAccessAuth will set the current authorization token.  If set to nil, it will not do authorization.  The
	// default is nil.
	SetConfigAccessAuth(auth *string)

	// FetchConfig will fetch the configuration at the end of the hierarchical path.
	FetchConfig(path ...string) *Config

	// PostConfig will place the config at the end of the path.
	PostConfig(config *Config, path ...string) error

	// RemoveConfig will remove a configuration point and all children.
	RemoveConfig(path ...string) error

	// RegisterService registers a service using the provided parameters.
	// Input:
	// - itype: You can use any InstanceType in registering, but those directly supported will be those noted as Host types.
	// - name: Service name.
	// - deployment: Deployment name.
	// - auth: The authorization token that the service may require to accept control requests.
	// - image: Container image name or id, if applicable.
	// - configProvider: Uri for the configuration provider.
	// - serviceClass: Primary service class that will automatically be made available for discovery.
	// - port: Port for the primary service class.  It may be overridden by service specifics.
	// - sspc: Service specifics is information given to the service to help with its customization.  They may be altered,
	//         but it is up to the service how they are used.  There is no specific set of these, since each service will have their own.
	// Output:
	// - confauth: The configuration provider access token given to the service that can be used to access its configuration.
	//             It is typically read only for the service.
	// - err: Any error encountered during registration.
	RegisterService(itype InstanceType, name, deployment, auth, image, configProvider, serviceClass string, port int,
		sspc map[string]interface{}) (string, error)

	// RegisterInstance should be done after RegisterService and the service is started.  auth will be the conig auth
	// token that the service can use to read the information.  itype is the sort of specific instance
	// was started.   'local' would be a process on the host where 'host' would mean a container in the hosts docker
	// engine.  Ginfra app defines that.   Ip and fqdn are as they say--the IP address and the Fully Qualified
	// Domain Name for the instance.  It returns an undefined screen (reserved for future use) and an error.
	RegisterInstance(itype InstanceType, name, deployment, auth, id, ip, fqdn string) (string, error)

	// DeregisterService removes a service registration for the specified name and deployment.
	DeregisterService(name, deployment string) error

	// DiscoverService will get a registered entry point for a service.  If none are registered it will return an error
	// with the ERRFLAG_NOT_FOUND flag.  If no flag is present an actual problem occured.
	DiscoverService(deployment, serviceClass string) (string, error)

	// DiscoveryAdd adds a new discovery entry for the specified deployment and service class.
	// The URI parameter specifies the location of the service discovery endpoint.   If the uri is an empty string, it
	// will remove any previously added discovery for the serviceClass.
	// Returns an error if the operation fails.
	DiscoveryAdd(deployment, serviceClass, uri string) error

	// PortalAdd registers a portal for a deployment.  The service class can ether be named, being one that had been
	// registered by services, or '*' meaning for all services.  The location is a resolvable bind name for the portal.
	// Portals are used to enter the deployment from outside, which means they are generally pointless for 'local' and
	// 'host' instance types and should be to service class '*' and the location being the name of the host machine.
	// It returns a string containing the registration token and an error if any.
	PortalAdd(deployment, serviceClass, location string) (string, error)

	// PortalRemove PortalAll will remove a portal for a service class.
	PortalRemove(deployment, serviceClass string) error

	// DiscoverPortal GetPortal for a serviceclass in a deployment.  The serviceclass may be "*" in which case it will get the top
	// level portal for the whole deployment.  The portal is some sort of usable address.
	DiscoverPortal(deployment, serviceClass string) (string, error)

	/***************************************************************************************
	 *  State is primarily used by the services for cooperation and is optional.
	 */

	// GetInitState will get the current initialization state.  Only one instance of a service will ever get InitNew,
	// indicating it is allowed to initialize the service.  Until that service posts the state, all other instances
	// will get an InitPending.  id uniquely identifies the requesting instance.
	GetInitState(deployment, name, id string) (InitState, error)

	// GetState will get the state for the service.
	GetState(deployment, name string) (map[string]interface{}, error)

	// PostState will post the state for the service.  If initialization is pending, it will only allow the pending id
	// to do it.
	PostState(deployment, name, id string, state map[string]interface{}) error

	/***************************************************************************************
	 *  Store allows maintaining common configuration files for services and whatnot.
	 *  There is nothing to stop you from using it as a file store, but I highly discourage it, as
	 *  it is intended only to secure a relatively small number of config files.
	 */

	// StorePut stores the given configuration under the specified name, deployment, and service.  Service and/or
	// deployment may be blank, but if used then the associated authorization is needed to put them. It returns
	// nil if the operation is successful, otherwise it returns an error.
	StorePut(name, deployment, service string, data []byte) error

	// StoreGet gets the given configuration under the specified name, deployment, service.  Service and/or
	// deployment may be blank, but it could determine specifically while file to get.  It will process the file
	// against the current configuration using config.ProcessConfigString, unless raw is set to true.
	StoreGet(name, deployment, service string, raw bool) ([]byte, error)

	// ClearStore clears the store of the ConfigProvider.  DO NOT implement as a service!
	ClearStore() error

	// PutValue sets or updates a configuration value identified by `name` in the specified `deployment`. Returns an
	// error if the operation fails.
	PutValue(deployment, name, value string) error

	// GetValue retrieves a configuration value identified by its name within the specified deployment, returning a
	// ConfigValue (which may contain an error).  Service is optional and may be blank, but should be used by services so
	// authentication works.
	GetValue(deployment, service, name string) *ConfigValue
}

// GetConfigProvider differentiates the providers based on uri.   A file: uri points to a file from where the config
// is stored.  file://test/config.json would use the test/config.json in the current working directory.  Whereas
// triple slashed such as file:///root/config/config.json would load /root/config/config.json starting at the
// filesystem root.  A http: (and eventually https: ) uri points to an API service.  The base should only be given
// such as http://localhost:8881.
func GetConfigProvider(uri string, token string) (ConfigProvider, error) {
	var (
		provider ConfigProvider
	)

	// Special case for error proxy provider.
	if strings.HasPrefix(uri, "error") {
		return getProxyErrorConfigProvider(uri, token)
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		panic(fmt.Sprintf("Poorly formed configuration provider uri: %s : %v.  Note that uri's do not support windows drive letter notation.", uri, err))
	}

	switch strings.ToLower(parsedURL.Scheme) {
	case "file":
		provider, err = GetConfigProviderFile(parsedURL.Host + parsedURL.Path)

	case "memory":
		provider, err = GetConfigProviderOnlyMemory()

	case "http", "https":
		provider, err = GetConfigProviderApi(parsedURL.Hostname(), parsedURL.Port(), base.ConfigServiceName, token)

	case "config":
		provider, err = GetConfigProviderConfigProvider()

	default:
		err = base.NewGinfraError(fmt.Sprintf(
			"ERROR: Unknown ConfigProvider type: %v", strings.ToLower(parsedURL.Scheme)))
	}

	return provider, err
}

func getProxyErrorConfigProvider(uri string, token string) (ConfigProvider, error) {

	cb := strings.Index(uri, ":")
	if len(uri) <= cb+1 {
		return nil, base.NewGinfraError("Broken error provider specification.  No real provider uri.")
	}

	d := strings.Split(uri[:cb], "/")
	if len(d) < 2 {
		return nil, base.NewGinfraError("Bad error provider specification.  You need at least one directive")
	}

	provider, err := GetConfigProvider(uri[cb+1:], token)
	if err != nil {
		return nil, err
	}

	ep := configProviderErrorable{
		errorMethods: make(map[string]string),
		cp:           provider,
	}

	for _, i := range d[1:] {
		if i != "" {
			nv := strings.Split(i, "=")
			if len(nv) < 2 {
				return nil, base.NewGinfraErrorA("Bad name=value for error directive.", base.LM_CAUSE, d)
			}
			ep.errorMethods[nv[0]] = nv[1]
		}

	}

	return &ep, nil
}

// #####################################################################################################################
// # SPECIFIC CONFIGS HELPERS

func GetServiceConfig(provider ConfigProvider, deployment string) *Config {
	return GetServiceConfigN(provider, deployment, base.GetRegisterService())
}

func GetServiceConfigN(provider ConfigProvider, deployment, service string) *Config {
	return provider.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_SERVICES, service, CRD_SERVICES__SERVICE,
		CRD_SSERVICE__CONFIG)
}

func GetServiceConfigSpecificsN(provider ConfigProvider, deployment, service, item string) *Config {
	return provider.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_SERVICES, service, CRD_SERVICES__SERVICE,
		CRD_SSERVICE__CONFIG, CRD_SSERVICE__CONFIG__SPECIFICS, item)
}

func SetServiceConfigSpecificsN(provider ConfigProvider, deployment, service, name, value string) *Config {
	cfg := provider.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_SERVICES, service, CRD_SERVICES__SERVICE,
		CRD_SSERVICE__CONFIG, CRD_SSERVICE__CONFIG__SPECIFICS)
	if cfg.Err != nil {
		return cfg
	}
	cfg.PutValue(name, value)
	cfg.Err = provider.PostConfig(cfg, CR_DEPLOYMENTS, deployment, CRD_SERVICES, service, CRD_SERVICES__SERVICE,
		CRD_SSERVICE__CONFIG, CRD_SSERVICE__CONFIG__SPECIFICS)
	return cfg
}

func PutServiceConfigN(provider ConfigProvider, deployment, service string, config *Config) error {
	return provider.PostConfig(config, CR_DEPLOYMENTS, deployment, CRD_SERVICES, service, CRD_SERVICES__SERVICE,
		CRD_SSERVICE__CONFIG)
}

//func GetServiceInstance(provider ConfigProvider, deployment string) *Config {
//	return GetServiceInstanceN(provider, deployment, common.GetRegisterService())
//}

func GetServiceInstance(provider ConfigProvider, deployment, service, id string) *Config {
	// Use default for everything for now.
	return provider.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_SERVICES, service, CRD_SERVICES__MANAGEMENT,
		CRD_MSERVICE__INSTANCE, id)
}

func GetServiceInstanceOnly(provider ConfigProvider, deployment, service string) (string, *Config) {
	// Use default for everything for now.
	c := provider.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_SERVICES, service, CRD_SERVICES__MANAGEMENT,
		CRD_MSERVICE__INSTANCE)
	for n, _ := range c.Config {
		return n, provider.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_SERVICES, service, CRD_SERVICES__MANAGEMENT,
			CRD_MSERVICE__INSTANCE, n)
	}
	return "", NewConfig(nil, base.NewGinfraError("No instances registered."))
}

func GetServiceAuthN(provider ConfigProvider, deployment, service string) *ConfigValue {
	// Use default for everything for now.
	return provider.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_SERVICES, service, CRD_SERVICES__SERVICE).
		GetValue(CRD_SERVICES__SERVICE_AUTH)
}

// #####################################################################################################################
// # CONFIGURATION PROVIDER CONFIG PROVIDER
// # The unique case where the service is the configuration provider and doesn't need any configuration

type configProviderConfigProvider struct {
	port int
}

func GetConfigProviderConfigProvider() (ConfigProvider, error) {
	var prov configProviderConfigProvider
	return &prov, nil
}

func (cpf *configProviderConfigProvider) IsConfigurable() bool {
	return false
}

func (cpf *configProviderConfigProvider) SetConfigAccessAuth(auth *string) {
}

func (cpf *configProviderConfigProvider) FetchConfig(path ...string) *Config {
	// Why in the world would I do this?  Because it makes a clean coverate report.
	return func() *Config { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) PostConfig(config *Config, path ...string) error {
	return func() error { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) RemoveConfig(path ...string) error {
	return func() error { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) RegisterService(itype InstanceType, name, deployment, auth, image, configProvider,
	serviceClass string, port int, sspc map[string]interface{}) (string, error) {
	return func() (string, error) { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) RegisterInstance(itype InstanceType, name, deployment, auth, id, ip,
	fqdn string) (string, error) {
	return func() (string, error) { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) PortalAdd(deployment, serviceClass, location string) (string, error) {
	return func() (string, error) { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) PortalRemove(deployment, serviceClass string) error {
	return func() error { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) DeregisterService(name, deployment string) error {
	return func() error { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) DiscoveryAdd(deployment, serviceClass, uri string) error {
	return func() error { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) DiscoverService(deployment, serviceClass string) (string, error) {
	return func() (string, error) { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) DiscoverPortal(deployment, serviceClass string) (string, error) {
	return func() (string, error) { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) GetInitState(deployment, name, id string) (InitState, error) {
	return func() (InitState, error) { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) GetState(deployment, name string) (map[string]interface{}, error) {
	return func() (map[string]interface{}, error) { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) PostState(deployment, name, id string, state map[string]interface{}) error {
	return func() error { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) StorePut(name, deployment, service string, data []byte) error {
	return func() error { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) StoreGet(name, deployment, service string, raw bool) ([]byte, error) {
	return func() ([]byte, error) { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) ClearStore() error {
	return nil
}

func (cpf *configProviderConfigProvider) PutValue(deployment, name, value string) error {
	return func() error { panic("configProviderConfigProvider cannot be configured.") }()
}

func (cpf *configProviderConfigProvider) GetValue(deployment, service, name string) *ConfigValue {
	panic("configProviderConfigProvider cannot be configured.")
}

// #####################################################################################################################
// # MOCK CONFIG PROVIDER
// # For testing

type MockConfigProvider struct{}

func (m *MockConfigProvider) IsConfigurable() bool                         { return true }
func (m *MockConfigProvider) SetConfigAccessAuth(auth *string)             {}
func (m *MockConfigProvider) FetchConfig(path ...string) *Config           { return nil }
func (m *MockConfigProvider) PostConfig(cfg *Config, path ...string) error { return nil }
func (m *MockConfigProvider) RemoveConfig(path ...string) error            { return nil }
func (m *MockConfigProvider) RegisterService(_ InstanceType, _, _, _, _, _, _ string, _ int, _ map[string]interface{}) (string, error) {
	return "", nil
}
func (m *MockConfigProvider) RegisterInstance(_ InstanceType, _, _, _, _, _, _ string) (string, error) {
	return "", nil
}
func (m *MockConfigProvider) DeregisterService(_, _ string) error                      { return nil }
func (m *MockConfigProvider) DiscoverService(_, _ string) (string, error)              { return "", nil }
func (m *MockConfigProvider) DiscoveryAdd(_, _, _ string) error                        { return nil }
func (m *MockConfigProvider) PortalAdd(_, _, _ string) (string, error)                 { return "", nil }
func (m *MockConfigProvider) PortalRemove(_, _ string) error                           { return nil }
func (m *MockConfigProvider) DiscoverPortal(_, _ string) (string, error)               { return "", nil }
func (m *MockConfigProvider) GetInitState(_, _, _ string) (InitState, error)           { return InitNew, nil }
func (m *MockConfigProvider) GetState(_, _ string) (map[string]interface{}, error)     { return nil, nil }
func (m *MockConfigProvider) PostState(_, _, _ string, _ map[string]interface{}) error { return nil }
func (m *MockConfigProvider) StorePut(_, _, _ string, _ []byte) error                  { return nil }
func (m *MockConfigProvider) StoreGet(_, _, _ string, _ bool) ([]byte, error)          { return nil, nil }
func (m *MockConfigProvider) ClearStore() error                                        { return nil }
func (m *MockConfigProvider) PutValue(_, _, _ string) error                            { return nil }
func (m *MockConfigProvider) GetValue(_, _, _ string) *ConfigValue                     { return nil }
