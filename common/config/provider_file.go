/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Configuration providers.  This is a file based provider that is convenient for local testing.
*/
package config

import (
	"encoding/json"
	"fmt"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// #####################################################################################################################
// # FILE CONFIG PROVIDER

type configProviderFile struct {
	config     *Config
	memoryOnly bool
	realauth   bool // Do real authentication.
	auth       *string

	savable   int
	savepath  string
	currentId int
}

func unmarshalGetConfigProviderFile(data []byte) (*Config, error) {
	c := NewConfig(nil, nil)
	err := json.Unmarshal(data, &c.Config)
	return c, err
}

func loadGetConfigProviderMultiFile(cpf *configProviderFile, location string) error {

	// Find latest file.
	var toLoad os.DirEntry
	var atLeastOneToLoad os.DirEntry
	sd := filepath.Join(location)
	fis, err := os.ReadDir(sd)
	if err != nil {
		return base.NewGinfraErrorChildA("Store directory is missing or corrupt.", err, base.LM_FILEPATH, sd)
	}
	for _, f := range fis {
		if !f.IsDir() && strings.HasPrefix(f.Name(), CONFIG_FILE_PREFIX) &&
			strings.HasSuffix(f.Name(), ".json") {
			v := strings.Split(f.Name(), ".")
			if len(v) < 3 {
				return base.NewGinfraErrorChildA("Config filename corrupt.", err, base.LM_FILEPATH,
					filepath.Join(sd, f.Name()))
			}
			id, err := strconv.Atoi(v[1])
			if err != nil {
				return base.NewGinfraErrorChildA("Config filename (after the '.') not a proper number.",
					err, base.LM_FILEPATH, filepath.Join(sd, f.Name()))
			}

			atLeastOneToLoad = f
			if id > cpf.currentId {
				cpf.currentId = id
				toLoad = f
			}
		}
	}

	if toLoad == nil && atLeastOneToLoad != nil {
		toLoad = atLeastOneToLoad
	}

	// Load it if it exists otherwise it is new and seed it.
	cpf.config = NewConfig(nil, nil)
	if toLoad == nil {
		cpf.config.Config[CR_DEPLOYMENTS] = make(map[string]interface{})
		cpf.config.Config[CR_RESERVATIONS] = make(map[string]interface{})
		cpf.config.Config[CR_AUTH] = AUTH_ADMIN_TOKEN_DEFAULT // Seed it as the default.

	} else {
		var data []byte
		data, err = os.ReadFile(filepath.Join(sd, toLoad.Name()))
		if err == nil {
			if err = json.Unmarshal(data, &cpf.config.Config); err == nil {
				// Update admin token if it is not set.
				if _, ok := cpf.config.Config[CR_AUTH]; !ok {
					cpf.config.Config[CR_AUTH] = GetNewAuthToken(false)
				}
			}
		}
	}
	return err
}

func loadGetConfigProviderFile(location string, memoryOnly bool, savable int, realauth bool) (ConfigProvider, error) {
	var (
		result configProviderFile
		err    error
		data   []byte
	)
	result.savable = savable
	result.memoryOnly = memoryOnly
	result.savepath = location
	result.realauth = realauth

	if memoryOnly {
		result.config = NewConfig(nil, nil)

	} else {

		if result.savable <= 1 {
			data, err = os.ReadFile(location)
			if err == nil {
				result.config, err = unmarshalGetConfigProviderFile(data)
			}

		} else {
			err = loadGetConfigProviderMultiFile(&result, location)
		}

	}

	return &result, err
}

func GetConfigProviderFile(location string) (ConfigProvider, error) {
	return loadGetConfigProviderFile(location, false, 0, false)
}

// GetConfigProviderFileSavable  If savable is zero, it will not be saved. If savable is set to one, it will be saved
// the originally loaded file.  location should point to the config file.  If it is more one, it will save to a new file
// while keeping only savable number around.  location should point to a directory for the config files.  realauth
// as true will cause it to create real auth tokens for new registrations.
func GetConfigProviderFileSavable(location string, savable int, realauth bool) (ConfigProvider, error) {
	return loadGetConfigProviderFile(location, false, savable, realauth)
}

func GetConfigProviderOnlyMemory() (ConfigProvider, error) {
	return loadGetConfigProviderFile("", true, 0, false)
}

func GetConfigProviderInMemory(location string) (ConfigProvider, error) {
	return loadGetConfigProviderFile(location, false, 0, false)
}

// configInMemoryOnly is intended for testing.
func (cpf *configProviderFile) configInMemoryOnly() {
	cpf.memoryOnly = true
}

/**
I'm on the fence on how to implement these.  Ideally there would be a single fetch
that handled all the subinterfaces.  However, that could leave the user with some
ambiguity on what they actually got.  I'm going to try this more brute method and see
how it wears.
*/

func (cpf *configProviderFile) IsConfigurable() bool {
	return true
}

func (cpf *configProviderFile) SetConfigAccessAuth(auth *string) {
	cpf.auth = auth
}

func (cpf *configProviderFile) authorize(rw, pathNotRequired bool, path ...string) error {

	if cpf.auth == nil {
		return nil
	}

	var (
		auths []string
		err   error
	)

	node := cpf.config.Config
	if a, ok := node[CR_AUTH]; ok {
		auths = append(auths, a.(string)) // TODO Consider type guarding here.
	}

	for idx, item := range path {
		if n, ok := node[item]; ok {
			if node, ok = n.(map[string]interface{}); ok {
				if a, ok := node[CR_AUTH]; ok {
					auths = append(auths, a.(string)) // TODO Consider type guarding here.
				}

			} else {
				// TERMINAL?
				if idx == len(path)-1 {
					break
				}
				err = base.NewGinfraErrorAFlags("Authorization failed.  Configuration path not found.  Path ended prematurely",
					base.ERRFLAG_NOT_FOUND, base.LM_CONFIG_PATH, common.Path2String(path...))
				break
			}

		} else {
			if pathNotRequired {
				break
			}
			err = base.NewGinfraErrorAFlags("Configuration path not found.", base.ERRFLAG_NOT_FOUND,
				base.LM_CONFIG_PATH, common.Path2String(path...))
			break
		}
	}
	if err != nil {
		return err
	}

	if rw {
		for _, i := range auths {
			if i == *cpf.auth && i[0:1] != CONFIG_RO_FLAG {
				return nil
			}
		}
	} else {
		for _, i := range auths {
			if i == *cpf.auth {
				return nil
			}
		}
	}

	return base.NewGinfraErrorFlags("Not authorized.", base.ERRFLAG_NOT_AUTHORIZED)
}

func (cpf *configProviderFile) FetchConfig(path ...string) *Config {
	if len(path) == 1 && strings.Index(path[0], ".") >= 0 {
		path = strings.Split(path[0], ".")
	}

	if err := cpf.authorize(false, false, path...); err != nil {
		return NewConfig(nil, err)
	}

	c := cpf.config.GetConfigPath(path...)
	return c
}

func getSaveFileName(id int) string {
	return fmt.Sprintf("%s%d%s", CONFIG_FILE_PREFIX, id, ".json")
}

func (cpf *configProviderFile) saveConfigStoreFile() {
	var err error
	if cpf.savable == 1 {
		d, _ := json.Marshal(cpf.config.Config)
		err = os.WriteFile(cpf.savepath, d, 0644)
		if err != nil {
			panic(base.NewGinfraErrorChildAFlags("Could not write configStoreFile.", err,
				base.ERRFLAG_SUBSYSTEM_SERIOUS, base.LM_FILEPATH, app.PathServiceDefPath))
		}

	} else if cpf.savable > 1 {
		cpf.currentId++

		sf := filepath.Join(cpf.savepath, getSaveFileName(cpf.currentId))
		d, _ := json.Marshal(cpf.config.Config)
		err = os.WriteFile(sf, d, 0644)
		if err != nil {
			panic(base.NewGinfraErrorChildAFlags("Could not write configStoreFile.", err,
				base.ERRFLAG_SUBSYSTEM_SERIOUS, base.LM_FILEPATH, app.PathServiceDefPath))
		} else {
			if cpf.currentId > 10 {
				// Lazy attempt to prune.
				_ = os.Remove(filepath.Join(cpf.savepath, getSaveFileName(cpf.currentId-10)))
			}
		}
	}
}

func (cpf *configProviderFile) PostConfig(config *Config, path ...string) error {
	var err error
	if err = cpf.authorize(true, true, path...); err == nil {
		if err = cpf.config.PutConfigPath(config, path...); err == nil {
			cpf.saveConfigStoreFile()
		}
	}
	return err
}

// RemoveConfig will remove a configuration point and all children.
func (cpf *configProviderFile) RemoveConfig(path ...string) error {
	var err error
	if err = cpf.authorize(true, false, path...); err == nil {
		if err := cpf.config.RemoveConfigPath(path...).Err; err == nil {
			cpf.saveConfigStoreFile()
		}
	}
	return err
}

func (cpf *configProviderFile) configCreateIfNotPresent(path ...string) *Config {
	c := cpf.config.GetConfigPath(path...)
	if c.Err != nil {
		c = NewConfig(nil, nil)
		c.Err = cpf.config.PutConfigPath(c, path...)
	}
	return c
}

func createNewDeploymentIfNotExist(cpf *configProviderFile, deployment string) {
	// Check if new deployment
	if c := cpf.config.GetConfigPath(CR_DEPLOYMENTS, deployment); c.Err != nil {
		config := NewConfig(nil, nil)
		config.Err = json.Unmarshal([]byte(DeploymentCreateJSON), &config.Config)
		err := cpf.PostConfig(config, CR_DEPLOYMENTS, deployment)
		if err != nil {
			panic(err) // Should never happen with file provider.
		}
	}
}

func (cpf *configProviderFile) RegisterService(itype InstanceType, name, deployment, auth, image, configProvider,
	serviceClass string, port int, sspc map[string]interface{}) (string, error) {

	var (
		its string
		err error
	)

	if err = cpf.authorize(true, false, CR_DEPLOYMENTS); err != nil {
		return "", err
	}

	if its, err = GetStringFromInstanceType(itype); err != nil {
		return "", err
	}

	confauth := CONFIG_NIL_AUTH
	if cpf.realauth {
		confauth = GetNewAuthToken(true)
	}

	if deployment == "" {
		deployment = CR_DEPLOYMENT_DEFAULT
	}

	_ = cpf.configCreateIfNotPresent(CR_DEPLOYMENTS)

	// Check if new deployment
	createNewDeploymentIfNotExist(cpf, deployment)
	srv := cpf.configCreateIfNotPresent(CR_DEPLOYMENTS, deployment, CRD_SERVICES)

	// Service
	srv = srv.GetConfigPath(name)
	if srv.Err == nil {
		return CONFIG_NIL_AUTH, base.NewGinfraErrorA("Service already registered.", base.LM_DEPLOYMENT_NAME,
			deployment)
	}
	//_ = cpf.configCreateIfNotPresent(CR_DEPLOYMENTS, deployment, CRD_SERVICES, name)

	d, err := common.SimpleReplacer(ServiceRegisterRequestJSON, map[string]interface{}{
		CR_AUTH: confauth, CR_SERVICE_AUTH: auth, CRD_SSERVICE__CONFIG__IMAGE: image,
		CRD_SSERVICE__CONFIG__PORT: strconv.Itoa(port), CRD_SSERVICE__CONFIG__CONFIG_PROVIDER: configProvider,
		CRD_SSERVICE__CONFIG__SERVICE_CLASS: serviceClass}, false)

	config := NewConfig(nil, nil)
	config.Err = json.Unmarshal([]byte(d), &config.Config)
	if config.Err != nil {
		panic(config.Err)
	}

	config.GetConfigPath(CRD_SERVICES__SERVICE, CRD_SSERVICE__CONFIG).PutValue(CRD_SSERVICE__CONFIG__SPECIFICS, sspc)

	err = cpf.config.PutConfigPath(config, CR_DEPLOYMENTS, deployment, CRD_SERVICES, name)
	if err != nil {
		panic(err) // Should never happen with file provider.
	}

	c := NewConfig(nil, nil)
	// Instance type
	c.PutValue(CL_INSTANCE_TYPE, its)
	err = cpf.config.PutConfigPath(c, CR_DEPLOYMENTS, deployment, CRD_SERVICES, name)
	if err != nil {
		panic(err) // Should never happen with file provider.
	}

	cpf.saveConfigStoreFile()

	return confauth, nil
}

func (cpf *configProviderFile) RegisterInstance(itype InstanceType, name, deployment, auth, id, ip, fqdn string) (string, error) {
	if err := cpf.authorize(true, false, CR_DEPLOYMENTS); err != nil {
		return "", err
	}

	var (
		its string
		err error
	)

	if deployment == "" {
		deployment = CR_DEPLOYMENT_DEFAULT
	}

	if its, err = GetStringFromInstanceType(itype); err != nil {
		return "", err
	}

	// Check deployment and service.
	c := cpf.FetchConfig(CR_DEPLOYMENTS)
	if c.Err != nil {
		panic(fmt.Sprintf("Corrupt file config store.  %v", c.Err))
	}
	c = c.GetConfig(deployment)
	if c.Err != nil {
		return CONFIG_NIL_AUTH, base.NewGinfraErrorA("Deployment not yet registered.", base.LM_DEPLOYMENT_NAME,
			deployment)
	}
	c = c.GetConfigPath(CRD_SERVICES, name)
	svc := c
	if c.Err != nil {
		return CONFIG_NIL_AUTH, base.NewGinfraErrorA("Service not yet registered.",
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, name)
	}

	// If it is the first it is ok for it to create the management path.
	c = c.GetConfigPath(CRD_SERVICES__MANAGEMENT)
	if c.Err != nil {
		mc := NewConfig(nil, nil)
		mc.PutValue(CR_AUTH, auth)
		_ = svc.PutConfigPath(mc, CRD_SERVICES__MANAGEMENT)
		mc = NewConfig(nil, nil)
		_ = svc.PutConfigPath(mc, CRD_SERVICES__MANAGEMENT, CRD_MSERVICE__INSTANCE)
	}

	// Create out new instance.
	mi := NewConfig(nil, nil)
	mi.PutValue(CRD_MSERVICE__INSTANCE__ID, id)
	mi.PutValue(CRD_MSERVICE__INSTANCE__TYPE, its)
	mi.PutValue(CRD_MSERVICE__INSTANCE__IP, ip)
	mi.PutValue(CRD_MSERVICE__INSTANCE__FQDN, fqdn)
	_ = svc.PutConfigPath(mi, CRD_SERVICES__MANAGEMENT, CRD_MSERVICE__INSTANCE, id)

	// Default discovery for instance.
	//err = cpf.discoveryRegisterInstance(name, deployment, fqdn)

	cpf.saveConfigStoreFile()
	return auth, err
}

func (cpf *configProviderFile) PortalAdd(deployment, serviceClass, location string) (string, error) {
	createNewDeploymentIfNotExist(cpf, deployment)

	c := cpf.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_DISCOVERY)
	if c.Err != nil {
		return "", base.NewGinfraErrorChildA("Could not add portal.  Is the deployment registered?", c.Err,
			base.LM_DEPLOYMENT_NAME, deployment)
	}

	c = NewConfig(nil, nil)
	c.PutValue(serviceClass, location)
	err := cpf.PostConfig(c, CR_DEPLOYMENTS, deployment, CRD_DISCOVERY, CRD_PORTALS)
	if err != nil {
		err = base.NewGinfraErrorChildAFlags("Could not add portal.  Config may be corrupt", err,
			base.ERRFLAG_SUBSYSTEM_SERIOUS, base.LM_DEPLOYMENT_NAME, deployment)
	}

	// Reserve the return for future use.
	return "", err
}

func (cpf *configProviderFile) PortalRemove(deployment, serviceClass string) error {
	c := cpf.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_DISCOVERY, CRD_PORTALS)
	if c.Err != nil {
		return base.NewGinfraErrorChildA("Portals do not exist.  Is the deployment registered?", c.Err,
			base.LM_DEPLOYMENT_NAME, deployment)
	}

	// Eat any error.
	_ = c.RemoveConfigPath(serviceClass)
	return nil
}

func (cpf *configProviderFile) DeregisterService(deployment, name string) error {
	if err := cpf.authorize(true, false, CR_DEPLOYMENTS); err != nil {
		return err
	}

	sc := cpf.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_SERVICES, name)
	if sc.Err != nil {
		return base.NewGinfraErrorChildA("Cannot deregister service that isn't registered.", sc.Err,
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, name)
	}

	sclass := sc.GetConfigPath(CRD_SERVICES__SERVICE, CRD_SSERVICE__CONFIG).GetValue(CRD_SSERVICE__CONFIG__SERVICE_CLASS).
		GetValueStringNoerr() + "/"

	d := cpf.FetchConfig(CR_DEPLOYMENTS, deployment, CRD_DISCOVERY, CRD_SERVICES)
	for n, _ := range d.Config {
		if n == name || n == sclass || strings.HasPrefix(n, sclass) {
			delete(d.Config, name)
		}
	}

	if err := cpf.RemoveConfig(CR_DEPLOYMENTS, deployment, CRD_SERVICES, name); err != nil {
		return base.NewGinfraErrorChildA("Could not deregister service.",
			err, base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, name)
	}

	cpf.saveConfigStoreFile()
	return nil
}

func (cpf *configProviderFile) DiscoveryAdd(deployment, serviceClass, uri string) error {
	sc := NewConfig(nil, nil)
	sc.PutValue(serviceClass, uri)
	err := cpf.PostConfig(sc, CR_DEPLOYMENTS, deployment, CRD_DISCOVERY, CRD_SERVICES)
	if err != nil {
		return base.NewGinfraErrorChildAFlags("Could add discovery data.  The config may be corrupt.",
			err, base.ERRFLAG_SUBSYSTEM_SERIOUS, base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_CLASS, serviceClass)
	}
	return nil
}

func (cpf *configProviderFile) DiscoverService(deployment, serviceClass string) (string, error) {
	// Everyone is authorized for now.
	ds := cpf.config.GetConfigPath(CR_DEPLOYMENTS, deployment, CRD_DISCOVERY, CRD_SERVICES)
	if ds.Err != nil {
		return "", base.NewGinfraErrorChildA("Service not discovered.  Deployment never registered.", ds.Err,
			base.LM_DEPLOYMENT_NAME, deployment)
	}

	var (
		ok  bool
		uri string
	)

	if strings.HasPrefix(serviceClass, "*/") {
		serviceClass = serviceClass[2:]
		for n, v := range ds.Config {
			nn := strings.Split(n, "/")
			if len(nn) > 1 && nn[1] == serviceClass {
				if uri, ok = v.(string); !ok {
					panic("Unsupported value in discovery service for deployment: " + deployment)
				}
				break
			}
		}

	} else {
		if v, ok := ds.Config[serviceClass]; ok {
			if uri, ok = v.(string); !ok {
				panic("Unsupported value in discovery service for deployment: " + deployment)
			}
		}
	}

	if uri == "" {
		return "", base.NewGinfraErrorAFlags("Service not discovered.  No service class running it.",
			base.ERRFLAG_NOT_FOUND, base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, serviceClass)
	}

	return uri, nil
}

func (cpf *configProviderFile) DiscoverPortal(deployment, serviceClass string) (string, error) {
	if serviceClass == "" {
		serviceClass = CO_ANY_PORTAL
	}

	// Everyone is authorized for now.
	ds := cpf.config.GetConfigPath(CR_DEPLOYMENTS, deployment, CRD_DISCOVERY, CRD_PORTALS)
	if ds.Err != nil {
		return "", base.NewGinfraErrorChildA("Portal not discovered.  Deployment never registered.", ds.Err,
			base.LM_DEPLOYMENT_NAME, deployment)
	}

	var (
		ok bool
		ui interface{}
		ph string
		r  string
	)
	if ui, ok = ds.Config[serviceClass]; !ok {
		// Fall back to any
		if ui, ok = ds.Config[CO_ANY_PORTAL]; !ok {
			return "", base.NewGinfraErrorA("Portal not discovered.  Default does not exist",
				base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, serviceClass)
		}
	}
	if ph, ok = ui.(string); !ok || len(ph) < 1 {
		return "", base.NewGinfraErrorA("Portal not discovered.  Corrupt portal configuration.",
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, serviceClass)
	}

	// EARLY IMPLEMENTATION.  I'm not sold completely on how to handle this.
	uri, err := cpf.DiscoverService(deployment, serviceClass)
	if err != nil {
		return "", base.NewGinfraErrorA("Portal not discovered.  No service associated with service class.",
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, serviceClass)
	}

	suri, err := url.Parse(uri)
	if err != nil {
		return "", base.NewGinfraErrorChildAFlags("Portal not discovered.  Service class config is corrupt.",
			err, base.ERRFLAG_SUBSYSTEM_SERIOUS, base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, serviceClass)
	}

	var phost, pport string
	sph := strings.Split(ph, ":")
	phost = sph[0]
	if len(sph) > 1 {
		pport = sph[1]
	}

	// TODO this is a mess.  Clean it up.
	if strings.HasPrefix(suri.Path, "/") {
		suri.Path = suri.Path[1:]
	}
	if len(pport) > 0 {

		//  HAX for now.
		if (suri.Scheme == "http" && pport == "80") || (suri.Scheme == "https" && pport == "443") {
			r = fmt.Sprintf("%s://%s/%s", suri.Scheme, phost, suri.Path)
		} else {
			r = fmt.Sprintf("%s://%s:%s/%s", suri.Scheme, phost, pport, suri.Path)
		}

	} else {
		sport := suri.Port()
		if len(sport) > 0 {
			r = fmt.Sprintf("%s://%s:%s/%s", suri.Scheme, phost, sport, suri.Path)
		} else {
			r = fmt.Sprintf("%s://%s/%s", suri.Scheme, phost, suri.Path)
		}

	}

	r = common.CleanUrl(r)

	if len(suri.RawPath) > 0 {
		r = r + "?" + suri.RawPath
	}

	return r, nil
}

func (cpf *configProviderFile) GetInitState(deployment, name, id string) (InitState, error) {
	i := cpf.config.GetConfigPath(CR_DEPLOYMENTS, deployment, CRD_SERVICES, name)
	if i.Err != nil {
		return 0, base.NewGinfraErrorA("Service not registered.", base.LM_DEPLOYMENT_NAME, deployment,
			base.LM_SERVICE_NAME, name)
	}

	init := i.GetConfigPath(CRD_SERVICES_INITID)
	if init.Err != nil {
		// If not set, set it with this id.  Return that the init is new.
		err := cpf.PostConfig(NewConfig(map[string]interface{}{CRD_SERVICES_INITID: id}, nil), CR_DEPLOYMENTS, deployment,
			CRD_SERVICES, name)
		return InitNew, err
	}

	if init = i.GetConfigPath(CRD_SERVICES__SERVICE, CRD_SSERVICE__STATE); init.Err != nil {
		// Id set but there is no state data.  That means it is pending waiting on the original request to post the state.
		return InitPending, nil
	}

	// Otherwise it is done.
	return InitDone, nil
}

func (cpf *configProviderFile) GetState(deployment, name string) (map[string]interface{}, error) {
	i := cpf.config.GetConfigPath(CR_DEPLOYMENTS, deployment, CRD_SERVICES, name)
	if i.Err != nil {
		return nil, base.NewGinfraErrorA("Service not registered.", base.LM_DEPLOYMENT_NAME, deployment,
			base.LM_SERVICE_NAME, name)
	}

	i = i.GetConfigPath(CRD_SERVICES__SERVICE, CRD_SSERVICE__STATE)
	if i.Err != nil {
		return nil, base.NewGinfraErrorA("State has need been initialized for this service not registered.",
			base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, name)
	}

	return i.Config, nil
}

func (cpf *configProviderFile) PostState(deployment, name, id string, state map[string]interface{}) error {
	i := cpf.config.GetConfigPath(CR_DEPLOYMENTS, deployment, CRD_SERVICES, name)
	if i.Err != nil {
		return base.NewGinfraErrorA("Service not registered.", base.LM_DEPLOYMENT_NAME, deployment,
			base.LM_SERVICE_NAME, name)
	}

	s := i.GetConfigPath(CRD_SERVICES__SERVICE, CRD_SSERVICE__STATE)
	if s.Err != nil {
		// Haven't init state yet so make sure this call is allowed.
		sid := i.GetValue(CRD_SERVICES_INITID)
		if sid.value != id {
			return base.NewGinfraErrorA("Only the initializing instance of a service my initialize the state.",
				base.LM_DEPLOYMENT_NAME, deployment, base.LM_SERVICE_NAME, name)
		}
	}

	// Post it
	i = i.GetConfigPath(CRD_SERVICES__SERVICE)
	i.PutValue(CRD_SSERVICE__STATE, state)

	return nil
}

func checkStorePut(cpf *configProviderFile, deployment, service string) (string, error) {
	var (
		err error
		dp  string
	)

	if cpf.memoryOnly {
		return "", base.NewGinfraError("Memory only cannot use the store.")
	}

	// It must be able to access deployments.
	if err = cpf.authorize(false, false, []string{CR_DEPLOYMENTS}...); err != nil {
		return "", err
	}

	if deployment == "" {

		if service == "" {
			dp = filepath.Join(cpf.savepath, CONFIG_FILE_STORE, CONFIG_FILE_ALL)
		} else {
			dp = filepath.Join(cpf.savepath, CONFIG_FILE_STORE, CONFIG_FILE_SERVICE, service)
		}

	} else {

		if service == "" {
			dp = filepath.Join(cpf.savepath, CONFIG_FILE_STORE, deployment)

		} else {
			dp = filepath.Join(cpf.savepath, CONFIG_FILE_STORE, deployment, service)
		}
	}

	return dp, nil
}

func checkStoreGet(cpf *configProviderFile, deployment, service string) (string, error) {
	var (
		err error
		dp  string
		ap  []string
	)

	if cpf.memoryOnly {
		return "", base.NewGinfraError("Memory only cannot use the store.")
	}

	if deployment == "" {

		ap = []string{CR_DEPLOYMENTS}
		if service == "" {
			dp = filepath.Join(cpf.savepath, CONFIG_FILE_STORE, CONFIG_FILE_ALL)
		} else {
			dp = filepath.Join(cpf.savepath, CONFIG_FILE_STORE, CONFIG_FILE_SERVICE, service)
		}

	} else {

		if service == "" {
			dp = filepath.Join(cpf.savepath, CONFIG_FILE_STORE, deployment)
			ap = []string{CR_DEPLOYMENTS, deployment}

		} else {
			dp = filepath.Join(cpf.savepath, CONFIG_FILE_STORE, deployment, service)
			ap = []string{CR_DEPLOYMENTS, deployment, CRD_SERVICES, service, CRD_SERVICES__SERVICE, CRD_SSERVICE__CONFIG}
		}
	}

	// Authorize based on what it is trying to get.
	if err = cpf.authorize(false, false, ap...); err != nil {
		return "", err
	}

	return dp, nil
}

func (cpf *configProviderFile) StorePut(name, deployment, service string, data []byte) error {

	dp, err := checkStorePut(cpf, deployment, service)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(dp, fs.ModePerm); err != nil {
		return base.NewGinfraErrorChildAFlags("Store directory could not be created.", err,
			base.ERRFLAG_SUBSYSTEM_SERIOUS, base.LM_PATH, dp)
	}

	if err = os.WriteFile(filepath.Join(dp, name), data, 0644); err != nil {
		return base.NewGinfraErrorChildA("Could not put file in store.", err, base.LM_DEPLOYMENT_NAME,
			deployment, base.LM_SERVICE_NAME, service, base.LM_NAME, name)
	}

	return nil
}

func (cpf *configProviderFile) StoreGet(name, deployment, service string, raw bool) ([]byte, error) {
	var (
		data []byte
	)

	dp, err := checkStoreGet(cpf, deployment, service)
	if err != nil {
		return data, err
	}

	if data, err = os.ReadFile(filepath.Join(dp, name)); err != nil {
		return data, base.NewGinfraErrorChildA("Could not get file from store.", err, base.LM_DEPLOYMENT_NAME,
			deployment, base.LM_SERVICE_NAME, service, base.LM_NAME, name)
	}

	if !raw {
		s, err := ProcessConfigString(cpf, ProcessConfig{Deployment: deployment, Service: service}, string(data))
		if err != nil {
			return data, base.NewGinfraErrorChildA("Could not process the file from store.", err, base.LM_DEPLOYMENT_NAME,
				deployment, base.LM_SERVICE_NAME, service, base.LM_NAME, name)
		}
		data = []byte(s)
	}

	return data, nil
}

// ClearStore clears the entire store, removing all files.
func (cpf *configProviderFile) ClearStore() error {
	if cpf.memoryOnly {
		return base.NewGinfraError("Memory only cannot use the store.")
	}

	err := os.RemoveAll(filepath.Join(cpf.savepath, CONFIG_FILE_STORE))
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(cpf.savepath, CONFIG_FILE_STORE), fs.ModePerm)
}

func (cpf *configProviderFile) PutValue(deployment, name, value string) error {
	var err error

	cfg := NewConfig(nil, nil)
	cfg.PutValue(name, value)
	path := []string{CR_DEPLOYMENTS, deployment, CRD_SERVICES__VALUES}

	if err = cpf.authorize(true, true, path...); err == nil {
		if err = cpf.config.PutConfigPath(cfg, path...); err == nil {
			cpf.saveConfigStoreFile()
		}
	}
	return err
}

func (cpf *configProviderFile) GetValue(deployment, service, name string) *ConfigValue {
	cfg := cpf.FetchConfig([]string{CR_DEPLOYMENTS, deployment, CRD_SERVICES__VALUES, name}...)
	return cfg.GetValue(name)
}
