//go:build unit_test

/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]

Unit testing for config package.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package config

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"path/filepath"
	"testing"
)

// ===============================================================================================================
// =

const TCPSERVICE_NAME = "userservice"
const TCPSERVICE_DEPLOYMENT = "deployment1"
const TCPSERVICE_SERVICECLASS = SERVICE_CLASS__USERINFO

var (
	s1Name  = "user_provider"
	s1Value = "file://test/userfile.json"
	token1  string
)

func TestConfigProvider(t *testing.T) {
	p, _ := GetConfigProviderOnlyMemory()
	var err error

	sspc := make(map[string]interface{})
	t.Run("Register server", func(t *testing.T) {
		sspc[s1Name] = s1Value
		token1, err = p.RegisterService(ITLocal, TCPSERVICE_NAME, TCPSERVICE_DEPLOYMENT, "BBBBBBBBBBB", "image",
			"http://localhost:5111", TCPSERVICE_SERVICECLASS, 8900, sspc)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = p.RegisterInstance(ITLocal, TCPSERVICE_NAME, TCPSERVICE_DEPLOYMENT, token1, "aaa", "127.0.0.1",
			"localhost")
		assert.NoError(t, err)
	})

	t.Run("Check registration", func(t *testing.T) {
		c := p.FetchConfig(CR_DEPLOYMENTS)
		cc := c.GetConfig(TCPSERVICE_DEPLOYMENT)
		if cc.Err != nil {
			t.Error(cc.Err)
			return
		}
		if cc = cc.GetConfigPath(TCPSERVICE_DEPLOYMENT, CRD_SERVICES, TCPSERVICE_NAME); cc.Err != nil {
			t.Error(cc.Err)
			return
		}
		cs := cc.GetConfig(CRD_SERVICES__SERVICE)
		if cs.Err != nil {
			t.Error(cc.Err)
			return
		}
		if cs.GetConfigPath(CRD_SSERVICE__CONFIG, CRD_SSERVICE__CONFIG__SPECIFICS).GetValue(s1Name).GetValueStringNoerr() !=
			s1Value {
			t.Error("Specifics not set.")
			return
		}
		if cs.GetConfigPath(CRD_SSERVICE__CONFIG).GetValue(CRD_SSERVICE__CONFIG__IMAGE).GetValueStringNoerr() !=
			"image" {
			t.Error("Image not set.")
			return
		}
		cs = cc.GetConfig(CRD_SERVICES__MANAGEMENT)
		if cs.Err != nil {
			t.Error(cs.Err)
			return
		}
		if cs.GetConfigPath(CRD_MSERVICE__INSTANCE, "aaa").GetValue(CRD_MSERVICE__INSTANCE__IP).GetValueStringNoerr() !=
			"127.0.0.1" {
			t.Error("ip not set.")
			return
		}
	})

	t.Run("Check helper -- get service config", func(t *testing.T) {
		svc := GetServiceConfigN(p, TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME)
		if svc.Err != nil {
			t.Error(svc.Err)
			return
		}
		assert.Equal(t, "http://localhost:5111", svc.GetValue(CRD_SSERVICE__CONFIG__CONFIG_PROVIDER).GetValueStringNoerr())
	})

	t.Run("Check helper -- get service instance", func(t *testing.T) {
		_, svc := GetServiceInstanceOnly(p, TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME)
		if svc.Err != nil {
			t.Error(svc.Err)
			return
		}
		assert.Equal(t, "local", svc.GetValue(CRD_MSERVICE__INSTANCE__TYPE).GetValueStringNoerr())
	})

	t.Run("Check helper -- set specifics item", func(t *testing.T) {
		cfg := SetServiceConfigSpecificsN(p, TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, "name2", "value2")
		if cfg.Err != nil {
			t.Error(cfg.Err)
			return
		}
	})

	t.Run("Check helper -- get specifics item", func(t *testing.T) {
		cfg := GetServiceConfigSpecificsN(p, TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, s1Name)
		if cfg.Err != nil {
			t.Error(cfg.Err)
			return
		}
		assert.Equal(t, s1Value, cfg.GetValue(s1Name).GetValueStringNoerr())

		cfg = GetServiceConfigSpecificsN(p, TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, "name2")
		if cfg.Err != nil {
			t.Error(cfg.Err)
			return
		}
		assert.Equal(t, "value2", cfg.GetValue("name2").GetValueStringNoerr())
	})

	scfg := GetServiceConfigN(p, TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME)
	t.Run("Check helper -- get service config", func(t *testing.T) {
		if scfg.Err != nil {
			t.Error(scfg.Err)
			return
		}
	})

	t.Run("Check helper -- put service config", func(t *testing.T) {
		scfg.PutValue("name3", "value3")
		err = PutServiceConfigN(p, TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, scfg)
		if err != nil {
			t.Error(err)
			return
		}

		scfg = GetServiceConfigN(p, TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME)
		if scfg.Err != nil {
			t.Error(scfg.Err)
			return
		}
		assert.Equal(t, "value3", scfg.GetValue("name3").GetValueStringNoerr())
	})

	t.Run("Discovery", func(t *testing.T) {
		err = p.DiscoveryAdd(TCPSERVICE_DEPLOYMENT, TCPSERVICE_SERVICECLASS, "http://localhost:8900")
		if err != nil {
			t.Error(err)
			return
		}

		uri, err := p.DiscoverService(TCPSERVICE_DEPLOYMENT, TCPSERVICE_SERVICECLASS)
		assert.NoError(t, err, "Discover "+TCPSERVICE_SERVICECLASS)
		assert.Equal(t, "http://localhost:8900", uri, "Discover uri "+TCPSERVICE_SERVICECLASS)

		uri, err = p.DiscoverService(TCPSERVICE_DEPLOYMENT, "monkey")
		assert.Error(t, err, "Don't discover.")
	})

}

func TestConfigProviderCases(t *testing.T) {

	t.Run("poorly formed provider url", func(t *testing.T) {
		assert.PanicsWithValue(t, "Poorly formed configuration provider uri: :ksdfsd:dsfsdf : parse \":ksdfsd:dsfsdf\": missing protocol scheme.  Note that uri's do not support windows drive letter notation.",
			func() { _, _ = GetConfigProvider(":ksdfsd:dsfsdf", "X") })
	})

	t.Run("poorly formed error proxy", func(t *testing.T) {
		_, err := GetConfigProvider("error:", "X")
		assert.ErrorContains(t, err, "No real provider uri")
		_, err = GetConfigProvider("error:asdfasfd", "X")
		assert.ErrorContains(t, err, "at least one directive")
		_, err = GetConfigProvider("error/sdfsdf/asdfasds:bork://", "X")
		assert.ErrorContains(t, err, "Unknown")
		_, err = GetConfigProvider("error/FetchConfig/:memory://", "X")
		assert.ErrorContains(t, err, "error directive")
	})

	t.Run("helpers", func(t *testing.T) {

		auth := AUTH_ADMIN_TOKEN_DEFAULT
		cp, err := GetConfigProviderFileSavable(filepath.Join(base.GetGinfraHome(), "common", "config", "test", "config.json"),
			0, true)
		if err != nil {
			t.Error(err)
			return
		}
		cp.SetConfigAccessAuth(&auth)

		// This should be the only time this package test will do this, but if not, it will panic here or somewhere else.
		base.RegisterService("giuser")
		c := GetServiceConfig(cp, "default")
		assert.NoError(t, c.Err)
		assert.Equal(t, "gservices/giuser", c.GetValue("image").GetValueStringNoerr())

		c = GetServiceInstance(cp, "default", "giuser", "709f9fd7cf56")
		assert.NoError(t, c.Err)
		assert.Equal(t, "host", c.GetValue("type").GetValueStringNoerr())

		cv := GetServiceAuthN(cp, "default", "giuser")
		assert.NoError(t, cv.GetValueError())
		assert.Equal(t, "MOMPT5A0GSLJIDPIENF5", cv.GetValueStringNoerr())

		// Not found
		cmp, err := GetConfigProviderOnlyMemory()
		assert.NoError(t, err)
		_, c = GetServiceInstanceOnly(cmp, "default", "giuser")
		assert.ErrorContains(t, c.Err, "No instances")
	})
}

func TestConfigProviderConfigProvider(t *testing.T) {
	t.Run("non-configurable functions allowed", func(t *testing.T) {
		c, err := GetConfigProviderConfigProvider()
		assert.NoError(t, err)
		assert.False(t, c.IsConfigurable())
		assert.NoError(t, c.ClearStore())

		v := "sfsdfs"
		c.SetConfigAccessAuth(&v)

		c, err = GetConfigProvider("config://", "XXXX")
		assert.NoError(t, err)
		assert.False(t, c.IsConfigurable())
		assert.Panics(t, func() { c.FetchConfig("blah") }, "FetchConfig did not panic.")
	})

	t.Run("panic when attempting configurable functions.", func(t *testing.T) {
		cpf := &configProviderConfigProvider{
			port: 8080,
		}
		assert.Panics(t, func() { cpf.FetchConfig("blah") }, "FetchConfig did not panic.")
		assert.Panics(t, func() { _ = cpf.PostConfig(&Config{}, "") }, "PostConfig did not panic.")
		assert.Panics(t, func() { _ = cpf.RemoveConfig("blah") }, "RemoveConfig did not panic.")
		assert.Panics(t, func() {
			_, _ = cpf.RegisterService(ITLocal, "blah", "blah", "blah",
				"blah", "blah", "blah", 8080, map[string]interface{}{})
		},
			"RegisterService did not panic.")
		assert.Panics(t, func() {
			_, _ = cpf.RegisterInstance(ITLocal, "name", "deployment", "auth",
				"id", "ip", "fqdn")
		}, "RegisterInstance did not panic.")
		assert.Panics(t, func() { _, _ = cpf.PortalAdd("", "", "") },
			"PortalAdd did not panic.")
		assert.Panics(t, func() { _ = cpf.PortalRemove("", "") }, "PortalRemove did not panic.")
		assert.Panics(t, func() { _ = cpf.DeregisterService("", "") }, "DeregisterService did not panic.")
		assert.Panics(t, func() { _ = cpf.DiscoveryAdd("", "", "") }, "DiscoveryAdd did not panic.")
		assert.Panics(t, func() { _, _ = cpf.DiscoverService("", "") },
			"DiscoverService did not panic.")
		assert.Panics(t, func() { _, _ = cpf.DiscoverPortal("", "") }, "DiscoverPortal did not panic.")
		assert.Panics(t, func() { _, _ = cpf.GetInitState("", "", "") }, "GetInitState did not panic.")
		assert.Panics(t, func() { _, _ = cpf.GetState("", "") }, "GetState did not panic.")
		assert.Panics(t, func() { _ = cpf.PostState("", "", "", map[string]interface{}{}) },
			"PostState did not panic.")
		assert.Panics(t, func() { _ = cpf.StorePut("", "", "", []byte{}) }, "StorePut did not panic.")
		assert.Panics(t, func() { _, _ = cpf.StoreGet("", "", "", true) }, "StoreGet did not panic.")
		assert.Panics(t, func() { _ = cpf.PutValue("", "", "") }, "PutValue did not panic.")
		assert.Panics(t, func() { _ = cpf.GetValue("", "", "") }, "GetValue did not panic.")
	})

}

func TestConfigProvideMock(t *testing.T) {

	t.Run("Call mocks.", func(t *testing.T) {
		cpf := MockConfigProvider{}
		assert.True(t, cpf.IsConfigurable())
		cpf.SetConfigAccessAuth(nil) // For coverage
		assert.Nil(t, cpf.FetchConfig(""))
		assert.Nil(t, cpf.PostConfig(nil))
		assert.Nil(t, cpf.RemoveConfig(""))
		i, e := cpf.RegisterService(ITLocal, "", "", "", "", "", "", 0, nil)
		assert.Equal(t, "", i)
		assert.Nil(t, e)
		i, e = cpf.RegisterInstance(ITLocal, "", "", "", "", "", "")
		assert.Equal(t, "", i)
		assert.Nil(t, e)
		assert.Nil(t, cpf.DeregisterService("", ""))
		i, e = cpf.DiscoverService("", "")
		assert.Equal(t, "", i)
		assert.Nil(t, e)
		assert.Nil(t, cpf.DiscoveryAdd("", "", ""))
		i, e = cpf.PortalAdd("", "", "")
		assert.Equal(t, "", i)
		assert.Nil(t, e)
		assert.Nil(t, cpf.PortalRemove("", ""))
		i, e = cpf.DiscoverPortal("", "")
		assert.Equal(t, "", i)
		assert.Nil(t, e)
		is, er := cpf.GetInitState("", "", "")
		assert.Equal(t, InitState(InitNew), is)
		assert.Nil(t, er)
		ifx, err := cpf.GetState("", "")
		assert.Nil(t, ifx)
		assert.Nil(t, err)
		assert.Nil(t, cpf.PostState("", "", "", nil))
		assert.Nil(t, cpf.StorePut("", "", "", nil))
		b, errr := cpf.StoreGet("", "", "", false)
		assert.Nil(t, b)
		assert.Nil(t, errr)
		assert.Nil(t, cpf.ClearStore())
		assert.Nil(t, cpf.PutValue("", "", ""))
		assert.Nil(t, cpf.GetValue("", "", ""))
	})

}
