//go:build local_test

/*
Package config
Copyright (c) 2024 Erich Gatejen
[LICENSE]

Local testing for config api.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	test "gitlab.com/ginfra/ginfra/common/test"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// ===============================================================================================================
// =

const MyConfigPort = 9500

const cTDeployment = "mydeploy"
const cTService = "myservice"
const cTType = ITHost
const ctPort int32 = 8905
const ctIP = "192.168.100.200"
const ctFQDN = "myservice.1.test.ginfra.tech"

const ctAltServiceClass = "alt.service.class"

var (
	ctTopAuth = AUTH_ADMIN_TOKEN_DEFAULT
	ctNewAuth string
)

func TestApiConfigProvider(t *testing.T) {

	err := RequestConfigService(MyConfigPort)
	if err != nil {
		t.Error("Could not get config service", err)
		return
	}
	defer func() {
		_ = ReleaseConfigService(MyConfigPort)
	}()

	cp, err := GetConfigProviderApi("localhost", strconv.Itoa(MyConfigPort), base.ConfigServiceName, AUTH_ADMIN_TOKEN_DEFAULT)

	if err != nil {
		t.Error("Could not get config provider", err)
		return
	}
	cpurl := fmt.Sprintf("http://localhost:%d", MyConfigPort)

	cp.SetConfigAccessAuth(&ctTopAuth)

	t.Run("Is configurable.  (API should be)", func(t *testing.T) {
		is := cp.IsConfigurable()
		assert.Equal(t, true, is)
	})

	sspc := make(map[string]interface{})

	t.Run("Register Service", func(t *testing.T) {
		ctNewAuth, err = cp.RegisterService(cTType, cTService, cTDeployment, AUTH_ADMIN_TOKEN_DEFAULT, "image", cpurl, cTService,
			int(ctPort), sspc)
		if err != nil {
			t.Error("Register failed", err)
			return
		}
	})

	t.Run("Register Instance", func(t *testing.T) {
		_, err = cp.RegisterInstance(cTType, cTService, cTDeployment, AUTH_ADMIN_TOKEN_DEFAULT, ctIP, ctIP, ctFQDN)
		if err != nil {
			t.Error("Register instance failed", err)
			return
		}
	})

	t.Run("Cannot Re-register Service", func(t *testing.T) {
		_, err = cp.RegisterService(cTType, cTService, cTDeployment, AUTH_ADMIN_TOKEN_DEFAULT, "image", cpurl, cTService,
			int(ctPort), sspc)
		if err == nil {
			t.Error("Allowed to re-register service.")
		}
	})

	t.Run("Discovery Add", func(t *testing.T) {
		err = cp.DiscoveryAdd(cTDeployment, cTService, fmt.Sprintf("http://%s:%d", ctFQDN, ctPort))
		if err != nil {
			t.Error("Discovery Add failed", err)
			return
		}
	})

	t.Run("Discover service", func(t *testing.T) {
		var ip string
		ip, err = cp.DiscoverService(cTDeployment, cTService)
		assert.Equal(t, fmt.Sprintf("http://%s:%d", ctFQDN, ctPort), ip, "Discovered.")

		ip, err = cp.DiscoverService("sfsdfsdfs", cTService)
		assert.NotEqual(t, fmt.Sprintf("http://%s:%d", ctFQDN, ctPort), ip, "Not discovered.  Bad deployment")

		ip, err = cp.DiscoverService(cTDeployment, "afsdf")
		assert.NotEqual(t, fmt.Sprintf("http://%s:%d", ctFQDN, ctPort), ip, "Not discovered.  Bad service class")
	})

	t.Run("Portal", func(t *testing.T) {
		p, err := cp.DiscoverPortal(cTDeployment, cTService)
		assert.NoError(t, err, "discover portal")
		assert.Equal(t, "http://localhost:8905/", p)
	})

	t.Run("Alternate Portal", func(t *testing.T) {

		_, err = cp.PortalAdd(cTDeployment, ctAltServiceClass, "xxx")
		if err != nil {
			t.Error("Could not add portal.", err)
			return
		}

		_, err := cp.DiscoverPortal(cTDeployment, ctAltServiceClass)
		assert.Error(t, err, "discover portal with service class not running")

		_, err = cp.DiscoverPortal(cTDeployment, "asdfasdf")
		assert.Error(t, err, "Discovered a portal that doesn't exist.")

		err = cp.PortalRemove(cTDeployment, ctAltServiceClass)
		if err != nil {
			t.Error("Could not remove portal.", err)
			return
		}
	})

	t.Run("Fetch Config", func(t *testing.T) {
		cp.SetConfigAccessAuth(&ctNewAuth)

		c := GetServiceConfigN(cp, cTDeployment, cTService)
		if c.Err != nil {
			t.Error("Could not fetch service config.", err)
			return
		}

		assert.Equal(t, cpurl, c.GetValue(CRD_SSERVICE__CONFIG__CONFIG_PROVIDER).GetValueStringNoerr(),
			"fetch config provider")
		assert.Equal(t, "image", c.GetValue(CRD_SSERVICE__CONFIG__IMAGE).GetValueStringNoerr())
		assert.Equal(t, cTService, c.GetValue(CRD_SSERVICE__CONFIG__SERVICE_CLASS).GetValueStringNoerr(),
			"fetch service class")

		assert.Error(t, c.GetValue("sfsdfs").GetValueError())
	})

	t.Run("Post/remove config", func(t *testing.T) {
		cp.SetConfigAccessAuth(&ctTopAuth)

		var c Config
		c.PutValue("newname", "newvalue")

		err = cp.PostConfig(&c, CR_DEPLOYMENTS, cTDeployment, CRD_SERVICES, cTService, CRD_SERVICES__SERVICE,
			CRD_SSERVICE__CONFIG, "newconfig")
		if err != nil {
			t.Error("Could not post config.", err)
			return
		}

		cr := cp.FetchConfig(CR_DEPLOYMENTS, cTDeployment, CRD_SERVICES, cTService, CRD_SERVICES__SERVICE,
			CRD_SSERVICE__CONFIG, "newconfig")
		if cr.Err != nil {
			t.Error("Could not fetch posted config.", cr.Err)
			return
		}
		assert.Equal(t, "newvalue", cr.GetValue("newname").GetValueStringNoerr())

		err = cp.RemoveConfig(CR_DEPLOYMENTS, cTDeployment, CRD_SERVICES, cTService, CRD_SERVICES__SERVICE,
			CRD_SSERVICE__CONFIG, "newconfig")
		if err != nil {
			t.Error("Could not remove posted config.", err)
			return
		}

		cr = cp.FetchConfig(CR_DEPLOYMENTS, cTDeployment, CRD_SERVICES, cTService, CRD_SERVICES__SERVICE,
			CRD_SSERVICE__CONFIG, "newconfig")
		assert.Error(t, cr.Err)

	})

	cp.SetConfigAccessAuth(&ctTopAuth)

	t.Run("State", func(t *testing.T) {
		s, err := cp.GetInitState(cTDeployment, cTService, ctIP)
		if err != nil {
			t.Error("Could not get init state.", err)
			return
		}
		assert.Equal(t, InitState(InitNew), s)

		s, err = cp.GetInitState(cTDeployment, cTService, "zzzzz")
		assert.Equal(t, InitState(InitPending), s)

		m := map[string]interface{}{"bork": "bork"}
		err = cp.PostState(cTDeployment, cTService, "zzzzz", m)
		assert.Error(t, err, "Not the initializing id")

		err = cp.PostState(cTDeployment, cTService, ctIP, m)
		assert.NoError(t, err, "The initializing id")

		s, err = cp.GetInitState(cTDeployment, cTService, "zzzzz")
		assert.Equal(t, InitState(InitDone), s)

		m, err = cp.GetState(cTDeployment, cTService)
		if err != nil {
			t.Error("Could not get state.", err)
			return
		}
		assert.Equal(t, "bork", m["bork"])

	})

	const ctFileName = "gggg"

	t.Run("Store", func(t *testing.T) {
		var data []byte

		if err = base.LoadGinfra(); err != nil {
			panic(err)
		}
		src := filepath.Join(base.GetGinfraHome(), "common", "config", "test", "api_source_store.txt")
		if data, err = os.ReadFile(src); err != nil {
			t.Error("Could not load test file.", err)
			return
		}

		err = cp.StorePut(ctFileName, cTDeployment, cTService, data)
		if err != nil {
			t.Error("Could not store.", err)
			return
		}

		_, err = cp.StoreGet("essdfsdfsd", cTDeployment, cTService, false)
		assert.Error(t, err)

		data, err = cp.StoreGet(ctFileName, cTDeployment, cTService, true)
		if err != nil {
			t.Error("Could not get stored file.", err)
			return
		}

		sp, err := test.NewScanPack(string(data), nil)
		if err == nil {
			assert.Less(t, 0, sp.Seek("api storable"))
		} else {
			t.Error("Could not scan.")
		}

		// This is a NOOP.
		//ClearStore() error
	})

	t.Run("Deregister Service", func(t *testing.T) {
		err = cp.DeregisterService(cTService, cTDeployment)
		if err != nil {
			t.Error("Could not deregister service.", err)
			return
		}

		err = cp.DeregisterService(cTService, cTDeployment)
		assert.Error(t, err, "Could deregister an already deregistered service.")
	})

}

func TestApiConfigProviderCases(t *testing.T) {

	err := RequestConfigService(MyConfigPort)
	if err != nil {
		t.Error("Could not get config service", err)
		return
	}
	defer func() {
		_ = ReleaseConfigService(MyConfigPort)
	}()

	t.Run("Bad config service.", func(t *testing.T) {
		_, err := GetConfigProviderApi("localhost", "77777", "xxx", AUTH_ADMIN_TOKEN_DEFAULT)
		assert.ErrorContains(t, err, "could not contact the configuration service")
	})

	cp, err := GetConfigProviderApi("localhost", strconv.Itoa(MyConfigPort), base.ConfigServiceName, AUTH_ADMIN_TOKEN_DEFAULT)
	if err != nil {
		panic(err)
	}

	aa := AUTH_ADMIN_TOKEN_DEFAULT
	ba := AUTH_USER_TOKEN_DEFAULT
	cp.SetConfigAccessAuth(&ba)

	t.Run("Post failed.", func(t *testing.T) {
		err = cp.PostConfig(NewConfig(map[string]interface{}{"a": "b"}, nil), "aaa")
		assert.ErrorContains(t, err, "Could not post config")
	})
	t.Run("Remove failed.", func(t *testing.T) {
		err = cp.RemoveConfig("aaa")
		assert.ErrorContains(t, err, "Could not remove config")
	})
	t.Run("Bad discovery add.", func(t *testing.T) {
		err = cp.DiscoveryAdd("b", "b", "b")
		assert.ErrorContains(t, err, "Could not add discovery")
	})

	cp.SetConfigAccessAuth(&aa)
	t.Run("Register bad instance type.", func(t *testing.T) {
		_, err = cp.RegisterService(11, "a", "a", "a", "a", "a", "a",
			9999, map[string]interface{}{})
		assert.ErrorContains(t, err, "Bad instance type")
	})
	t.Run("Instance bad instance type.", func(t *testing.T) {
		_, err = cp.RegisterInstance(11, "a", "a", "a", "a", "a", "a")
		assert.ErrorContains(t, err, "Bad instance type")
	})
	t.Run("Bad instance registration.", func(t *testing.T) {
		_, err = cp.RegisterInstance(ITLocal, "a", "a", "a", "a", "a", "a")
		assert.ErrorContains(t, err, "Could not register service instance")
	})
	t.Run("Bad get init state.", func(t *testing.T) {
		_, err = cp.GetInitState("a", "a", "a")
		assert.ErrorContains(t, err, "Could not get init state")
	})
	t.Run("Bad get state.", func(t *testing.T) {
		_, err = cp.GetState("a", "a")
		assert.ErrorContains(t, err, "Could not get state")
	})
	t.Run("Bad post state.", func(t *testing.T) {
		err = cp.PostState("a", "a", "a", map[string]interface{}{})
		assert.ErrorContains(t, err, "Could not post state")
	})

}
