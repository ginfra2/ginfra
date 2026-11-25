/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]

Configuration providers.  This is an API based provider that can us any service that implements the
ginterfaces/service_giconfig.yaml.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package config

import (
	"bytes"
	"context"
	"gitlab.com/ginfra/ginfra/api/service_giconfig/client"
	"gitlab.com/ginfra/ginfra/api/service_giconfig/models"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/ifx"
	"net/http"
)

// #####################################################################################################################
// # HTTP API CONFIG PROVIDER

type configProviderApi struct {
	Host, Port, Auth string
	Client           *client.ClientWithResponses
}

func GetConfigProviderApi(host, port, service, token string) (ConfigProvider, error) {

	prov := configProviderApi{
		Host: host,
		// Don't use localhost or loopback.  They could end up on an interface that other nodes can't talk to.
		Port: common.GetActualLocalhostIf(port),
		Auth: token,
	}

	// Ping the service to check health.
	_, err := ifx.PingService(prov.Host, prov.Port, service, "Configuration provider")
	if err != nil {
		return nil, base.NewGinfraErrorA("Configuration provider could not contact the configuration service",
			base.LM_CAUSE, err)
	}

	// Create a client.  The only way to get an error is through the options, which we don't use now.
	prov.Client, err = client.NewClientWithResponses("http://" + host + ":" + port)
	if err != nil {
		return nil, base.NewGinfraErrorA("Configuration provider could not create a client", base.LM_CAUSE,
			err)
	}

	return &prov, err
}

func (cpf *configProviderApi) IsConfigurable() bool {
	return true
}

func (cpf *configProviderApi) SetConfigAccessAuth(auth *string) {
	cpf.Auth = *auth
}

func (cpf *configProviderApi) FetchConfig(path ...string) *Config {

	var (
		config *Config
		params models.GiConfigQueryConfigParams
	)

	if len(path) < 1 {
		params.Qpath = ""
	} else if len(path) == 1 {
		params.Qpath = path[0]
	} else {
		params.Qpath = common.Path2String(path...)
	}
	params.Confauth = cpf.Auth

	if r, err := cpf.Client.GiConfigQueryConfigWithResponse(context.Background(), &params); err == nil {
		if r.StatusCode() != http.StatusOK {
			if r.StatusCode() == http.StatusNotFound {
				config = NewConfig(nil, base.NewGinfraErrorAFlags("Could not fetch config.  Path not found",
					base.ERRFLAG_NOT_FOUND, base.LM_CAUSE, string(r.Body), base.LM_CODE, r.StatusCode()))
			} else {
				config = NewConfig(nil, base.NewGinfraErrorA("Could not fetch config.", base.LM_CAUSE, string(r.Body),
					base.LM_CODE, r.StatusCode()))
			}
		} else {
			config = NewConfig(common.Base64ToJsonMap(*r.JSON200.Block)) // How in the world is this compiling?
		}

	} else {
		config = NewConfig(nil, err)
	}

	return config
}

func (cpf *configProviderApi) PostConfig(config *Config, path ...string) error {
	var items []models.ConfigSetItem

	cpath := common.Path2String(path...)
	params := models.GiConfigSetConfigParams{
		Confauth: cpf.Auth,
	}

	body := models.GiConfigSetConfigJSONRequestBody{
		Items: &items,
	}

	// Only supporting one item for now.
	block := common.JsonMapToBase64(config.Config)
	item := models.ConfigSetItem{
		Path:  &cpath,
		Block: &block,
	}
	items = append(items, item)

	r, err := cpf.Client.GiConfigSetConfigWithResponse(context.Background(), &params, body)
	if err == nil && r.StatusCode() != http.StatusOK {
		err = base.NewGinfraErrorA("Could not post config.", base.LM_CAUSE, string(r.Body), base.LM_CODE,
			r.StatusCode())
	}
	return err
}

// RemoveConfig will remove a configuration point and all children.
func (cpf *configProviderApi) RemoveConfig(path ...string) error {
	params := models.GiRemoveConfigParams{
		Confauth: cpf.Auth,
		Qpath:    common.Path2String(path...),
	}

	r, err := cpf.Client.GiRemoveConfigWithResponse(context.Background(), &params)
	if err == nil && r.StatusCode() != http.StatusOK {

		err = base.NewGinfraErrorA("Could not remove config.", base.LM_CAUSE, string(r.Body), base.LM_CODE,
			r.StatusCode())
	}
	return err
}

func (cpf *configProviderApi) RegisterService(itype InstanceType, name, deployment, auth, image, configProvider,
	serviceClass string, port int, sspc map[string]interface{}) (string, error) {
	var (
		its string
		err error
	)

	params := models.GiConfigRegisterServiceParams{
		Confauth: cpf.Auth,
	}

	if its, err = GetStringFromInstanceType(itype); err != nil {
		return "", err
	}
	spc := common.JsonMapToBase64(sspc)

	p32 := int32(port)
	body := models.GiConfigRegisterServiceJSONRequestBody{
		Itype:          &its,
		Name:           &name,
		Deployment:     &deployment,
		Auth:           &auth,
		Image:          &image,
		Configprovider: &configProvider,
		Port:           &p32,
		Serviceclass:   &serviceClass,
		Specifics:      &spc,
	}

	r, err := cpf.Client.GiConfigRegisterServiceWithResponse(context.Background(), &params, body)
	if err != nil {
		return CONFIG_NIL_AUTH, base.NewGinfraErrorChildA("Could not connect to configuration provider service.", err,
			base.LM_SERVICE_NAME, name, base.LM_TARGET, configProvider)
	}
	if r.StatusCode() != 200 {
		return CONFIG_NIL_AUTH, base.NewGinfraErrorChildA("Could not get register service.", err,
			base.LM_SERVICE_NAME, name, base.LM_CODE, r.StatusCode(), base.LM_CAUSE, string(r.Body))
	}
	return *r.JSON200.Confauth, nil
}

func (cpf *configProviderApi) RegisterInstance(itype InstanceType, name, deployment, auth, id, ip,
	fqdn string) (string, error) {
	var (
		its string
		err error
	)

	if its, err = GetStringFromInstanceType(itype); err != nil {
		return "", err
	}

	params := models.GiConfigRegisterInstanceParams{
		Confauth: cpf.Auth,
	}
	body := models.GiConfigRegisterInstanceJSONRequestBody{
		Name:       &name,
		Deployment: &deployment,
		Sauth:      &auth,
		Itype:      &its,
		Id:         &id,
		Ip:         &ip,
		Fqdn:       &fqdn,
	}

	// -- Default deployment --------------------------------------------------------------

	r, err := cpf.Client.GiConfigRegisterInstanceWithResponse(context.Background(), &params, body)
	if err != nil || r.StatusCode() != 200 {
		return CONFIG_NIL_AUTH, base.NewGinfraErrorChildA("Could not register service instance.", err,
			base.LM_SERVICE_NAME, name, base.LM_CODE, r.StatusCode(), base.LM_CAUSE, string(r.Body))
	}
	return *r.JSON200.Confauth, nil
}

func (cpf *configProviderApi) PortalAdd(deployment, serviceClass, location string) (string, error) {
	body := models.GiConfigPortalJSONRequestBody{
		Sclass: &serviceClass,
		Portal: &location,
	}
	params := models.GiConfigPortalParams{
		Confauth: cpf.Auth,
	}

	r, err := cpf.Client.GiConfigPortalWithResponse(context.Background(), models.GiConfigPortalParamsOperation(OpPut),
		deployment, &params, body)
	if err != nil || r.StatusCode() != 200 {
		err = base.NewGinfraErrorChildA("Could not add portal.", err,
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_CLASS, serviceClass, base.LM_CAUSE, string(r.Body))
	}

	// Reserve the return for future use.
	return "", err
}

func (cpf *configProviderApi) PortalRemove(deployment, serviceClass string) error {
	body := models.GiConfigPortalJSONRequestBody{
		Sclass: &serviceClass,
	}
	params := models.GiConfigPortalParams{
		Confauth: cpf.Auth,
	}

	r, err := cpf.Client.GiConfigPortalWithResponse(context.Background(), models.GiConfigPortalParamsOperation(OpRemove),
		deployment, &params, body)
	if err != nil || r.StatusCode() != 200 {
		err = base.NewGinfraErrorChildA("Could not remove portal.", err,
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_CLASS, serviceClass, base.LM_CAUSE, string(r.Body))
	}

	// Reserve the return for future use.
	return err
}

func (cpf *configProviderApi) DeregisterService(name, deployment string) error {
	params := models.GiConfigDeregisterServiceParams{
		Confauth: cpf.Auth,
	}

	r, err := cpf.Client.GiConfigDeregisterServiceWithResponse(context.Background(), deployment, name, &params)
	if err != nil || r.StatusCode() != 200 {
		return base.NewGinfraErrorChildA("Could not get deregister service deployment.", err,
			base.LM_SERVICE_NAME, name, base.LM_CODE, r.StatusCode(), base.LM_CAUSE, string(r.Body))
	}
	return nil
}

func (cpf *configProviderApi) DiscoveryAdd(deployment, serviceClass, uri string) error {
	params := models.GiConfigDiscoveryAddParams{
		Confauth:   cpf.Auth,
		Deployment: deployment,
		Sclass:     serviceClass,
		Uri:        uri,
	}

	r, err := cpf.Client.GiConfigDiscoveryAddWithResponse(context.Background(), &params)
	if err != nil {
		return base.NewGinfraErrorChildA("Could not add discovery service class due to a problem.", err,
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_CLASS, serviceClass)
	}

	if r.StatusCode() != 200 {
		return base.NewGinfraErrorA("Could not add discovery service class because of a remote error.",
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_CLASS, serviceClass, base.LM_CODE, r.StatusCode(),
			base.LM_CAUSE, string(r.Body))
	}

	return nil
}

func (cpf *configProviderApi) DiscoverService(deployment, serviceClass string) (string, error) {
	params := models.GiDiscoverServiceParams{
		Confauth: cpf.Auth,
	}

	r, err := cpf.Client.GiDiscoverServiceWithResponse(context.Background(), deployment, serviceClass, &params)
	if err != nil || r.StatusCode() != 200 {
		if r.StatusCode() == http.StatusNotFound {
			return "", base.NewGinfraErrorChildAFlags("Could not discover service class because none is running.",
				err, base.ERRFLAG_NOT_FOUND, base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_CLASS, serviceClass,
				base.LM_CODE, r.StatusCode(), base.LM_CAUSE, string(r.Body))
		} else {
			return "", base.NewGinfraErrorChildA("Could not discover service class due to a problem.", err,
				base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_CLASS, serviceClass, base.LM_CODE, r.StatusCode(),
				base.LM_CAUSE, string(r.Body))
		}
	}
	return *r.JSON200.Uri, nil
}

func (cpf *configProviderApi) DiscoverPortal(deployment, serviceClass string) (string, error) {
	params := models.GiDiscoverPortalParams{
		Confauth: cpf.Auth,
	}

	r, err := cpf.Client.GiDiscoverPortalWithResponse(context.Background(), deployment, serviceClass, &params)
	if err != nil || r.StatusCode() != 200 {
		return "", base.NewGinfraErrorChildA("Could not discover portal for service class.", err,
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_CLASS, serviceClass, base.LM_CODE, r.StatusCode(),
			base.LM_CAUSE, string(r.Body))
	}
	return *r.JSON200.Uri, nil
}

func (cpf *configProviderApi) GetInitState(deployment, name, id string) (InitState, error) {
	params := models.GiConfigGetInitParams{
		Confauth: cpf.Auth,
	}

	r, err := cpf.Client.GiConfigGetInitWithResponse(context.Background(), deployment, name, id, &params)
	if err != nil || r.StatusCode() != 200 {
		return 0, base.NewGinfraErrorChildA("Could not get init state for service.", err,
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, name, base.LM_CODE, r.StatusCode(),
			base.LM_CAUSE, string(r.Body))
	}
	return InitState(*r.JSON200.State), nil
}

func (cpf *configProviderApi) GetState(deployment, name string) (map[string]interface{}, error) {
	var (
		state map[string]interface{}
	)
	params := models.GiConfigGetStateParams{
		Confauth: cpf.Auth,
	}

	if r, err := cpf.Client.GiConfigGetStateWithResponse(context.Background(), deployment, name, &params); err == nil {
		if r.StatusCode() == http.StatusOK {
			state, err = common.Base64ToJsonMap(*r.JSON200.Block) // How in the world is this compiling?

		} else {
			return nil, base.NewGinfraErrorA("Could not get state of service.", base.LM_DEPLOYMENT_NAME,
				deployment, base.LM_SERVICE_NAME, name, base.LM_CAUSE, string(r.Body), base.LM_CODE, r.StatusCode())
		}
	} else {
		return nil, base.NewGinfraErrorChildA("Could not get state from configuration service.", err,
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, name)
	}

	return state, nil
}

func (cpf *configProviderApi) PostState(deployment, name, id string, state map[string]interface{}) error {
	var body models.GiConfigPostStateJSONRequestBody
	params := models.GiConfigPostStateParams{
		Confauth: cpf.Auth,
	}

	// Only supporting one item for now.
	block := common.JsonMapToBase64(state)
	body.Block = &block

	r, err := cpf.Client.GiConfigPostStateWithResponse(context.Background(), deployment, name, id, &params, body)
	if err == nil {
		if r.StatusCode() != http.StatusOK {
			err = base.NewGinfraErrorA("Could not post state.", base.LM_DEPLOYMENT_NAME, deployment,
				base.LM_SERVICE_NAME, name, base.LM_CAUSE, string(r.Body), base.LM_CODE, r.StatusCode())
		}
	} else {
		err = base.NewGinfraErrorChildA("Could not post state to configuration service.", err,
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, name)
	}
	return err
}

func (cpf *configProviderApi) StorePut(name, deployment, service string, data []byte) error {
	params := models.GiConfigStorePutParams{
		Confauth:   cpf.Auth,
		Name:       name,
		Deployment: &deployment,
		Service:    &service,
	}

	r, err := cpf.Client.GiConfigStorePutWithBodyWithResponse(context.Background(), &params,
		"application/octet-stream", bytes.NewReader(data))

	if err == nil {
		if r.StatusCode() != http.StatusOK {
			err = base.NewGinfraErrorA("Could not put file to store due to error.", base.LM_NAME, name,
				base.LM_CAUSE, string(r.Body), base.LM_CODE, r.StatusCode())
		}
	} else {
		err = base.NewGinfraErrorChildA("Could not put file to store.", err, base.LM_NAME, name)
	}
	return err
}

func (cpf *configProviderApi) StoreGet(name, deployment, service string, raw bool) ([]byte, error) {
	params := models.GiConfigStoreGetParams{
		Confauth:   cpf.Auth,
		Name:       name,
		Deployment: &deployment,
		Service:    &service,
		Raw:        raw,
	}

	r, err := cpf.Client.GiConfigStoreGetWithResponse(context.Background(), &params)

	if err == nil {
		if r.StatusCode() != http.StatusOK {
			return []byte{}, base.NewGinfraErrorA("Could not get file from store due to error.", base.LM_NAME, name,
				base.LM_CAUSE, string(r.Body), base.LM_CODE, r.StatusCode())

		}

	} else {
		return []byte{}, base.NewGinfraErrorChildA("Could not get file from store.", err, base.LM_NAME, name)
	}

	return r.Body, nil

}

// ClearStore DO NOT implement this as a service.  Consequently, the client is a NOOP.
func (cpf *configProviderApi) ClearStore() error {
	return nil
}

func (cpf *configProviderApi) PutValue(deployment, name, value string) error {
	params := models.GiConfigPutValueParams{
		Confauth:   cpf.Auth,
		Deployment: deployment,
		Name:       name,
	}

	r, err := cpf.Client.GiConfigPutValueWithResponse(context.Background(), &params, value)
	if err == nil {
		if r.StatusCode() != http.StatusOK {
			err = base.NewGinfraErrorA("Could not put value.", base.LM_DEPLOYMENT_NAME, deployment,
				base.LM_NAME, name, base.LM_CAUSE, string(r.Body), base.LM_CODE, r.StatusCode())
		}
	} else {
		err = base.NewGinfraErrorChildA("Could not put value.", err, base.LM_DEPLOYMENT_NAME, deployment)
	}

	return err
}

func (cpf *configProviderApi) GetValue(deployment, service, name string) *ConfigValue {
	var (
		value string
		err   error
		r     *client.GiConfigGetValueResponse
	)

	params := models.GiConfigGetValueParams{
		Confauth:   cpf.Auth,
		Deployment: deployment,
		Service:    &service,
		Name:       name,
	}
	r, err = cpf.Client.GiConfigGetValueWithResponse(context.Background(), &params)
	if err == nil {
		if r.StatusCode() == http.StatusOK {
			value = *r.JSON200.Block // How in the world is this compiling?

		} else if r.StatusCode() == http.StatusNotFound {
			err = base.NewGinfraErrorAFlags("Value not found.", base.ERRFLAG_NOT_FOUND, base.LM_DEPLOYMENT_NAME,
				deployment, base.LM_SERVICE_NAME, name, base.LM_CODE, r.StatusCode())
		} else if r.StatusCode() == http.StatusUnauthorized {
			err = base.NewGinfraErrorAFlags("Not authorized to get value.", base.ERRFLAG_NOT_AUTHORIZED, base.LM_DEPLOYMENT_NAME,
				deployment, base.LM_SERVICE_NAME, name, base.LM_CODE, r.StatusCode())
		} else {
			err = base.NewGinfraErrorAFlags("Could not get value from configuration service.",
				base.ERRFLAG_RECOVERABLE, base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, name, base.LM_CAUSE,
				string(r.Body), base.LM_CODE, r.StatusCode())
		}
	} else {
		err = base.NewGinfraErrorChildAFlags("Could not get value from configuration service.", err,
			base.ERRFLAG_RECOVERABLE, base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, name)
	}

	return &ConfigValue{
		name:  name,
		value: value,
		err:   err,
	}

}
