//go:build unit_test

/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

This is a provider proxy test.
*/
package config

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"testing"
)

func init() {
	common.SetupTest()
}

func TestConfigProviderErrorable(t *testing.T) {

	cpe, err := GetConfigProvider("error/FetchConfig=FetchConfigError/:file://"+base.GetGinfraHome()+"/test/local/service_dirs/ok_specifics/config.json",
		"AAAAAAAAAAAA")
	if err != nil {
		t.Error(err)
		return
	}
	cpes := cpe.(*configProviderErrorable)

	t.Run("FetchConfig", func(t *testing.T) {
		c := cpe.FetchConfig("blah")
		assert.Contains(t, c.Err.Error(), "FetchConfigError")
	})

	t.Run("PostConfig", func(t *testing.T) {
		cpes.SetError("PostConfig", "PostConfigError")
		c := Config{}
		err = cpe.PostConfig(&c, "blah")
		assert.Contains(t, err.Error(), "PostConfigError")
	})

	t.Run("RemoveConfig", func(t *testing.T) {
		cpes.SetError("RemoveConfig", "RemoveConfigError")
		err = cpe.RemoveConfig("blah")
		assert.Contains(t, err.Error(), "RemoveConfigError")
	})

	t.Run("RegisterService", func(t *testing.T) {
		cpes.SetError("RegisterService", "RegisterServiceError")
		_, err = cpe.RegisterService(ITLocal, "sdf", "sdf", "sdf", "dsf", "Sdf",
			"sdf", 1, map[string]interface{}{})
		assert.Contains(t, err.Error(), "RegisterServiceError")
	})

	t.Run("RegisterInstance", func(t *testing.T) {
		cpes.SetError("RegisterInstance", "RegisterInstanceError")
		_, err = cpe.RegisterInstance(ITLocal, "sdf", "sdf", "sdf", "dsf", "Sdf", "sdf")
		assert.Contains(t, err.Error(), "RegisterInstanceError")
	})

	t.Run("DeregisterService", func(t *testing.T) {
		cpes.SetError("DeregisterService", "DeregisterServiceError")
		err = cpe.DeregisterService("sdf", "sdf")
		assert.Contains(t, err.Error(), "DeregisterServiceError")
	})

	t.Run("PortalAdd", func(t *testing.T) {
		cpes.SetError("PortalAdd", "PortalAddError")
		_, err = cpe.PortalAdd("blah", "blah", "blah")
		assert.Contains(t, err.Error(), "PortalAddError")
	})

	t.Run("PortalRemove", func(t *testing.T) {
		cpes.SetError("PortalRemove", "PortalRemoveError")
		err = cpe.PortalRemove("blah", "blah")
		assert.Contains(t, err.Error(), "PortalRemoveError")
	})

	t.Run("DiscoveryAdd", func(t *testing.T) {
		cpes.SetError("DiscoveryAdd", "DiscoveryAddError")
		err = cpe.DiscoveryAdd("blah", "blah", "blah")
		assert.Contains(t, err.Error(), "DiscoveryAddError")
	})

	t.Run("DiscoverService", func(t *testing.T) {
		cpes.SetError("DiscoverService", "DiscoverServiceError")
		_, err = cpe.DiscoverService("blah", "blah")
		assert.Contains(t, err.Error(), "DiscoverServiceError")
	})

	t.Run("DiscoverPortal", func(t *testing.T) {
		cpes.SetError("DiscoverPortal", "DiscoverPortalError")
		_, err = cpe.DiscoverPortal("blah", "blah")
		assert.Contains(t, err.Error(), "DiscoverPortalError")
	})

	t.Run("GetInitState", func(t *testing.T) {
		cpes.SetError("GetInitState", "GetInitStateError")
		_, err = cpe.GetInitState("blah", "blah", "blah")
		assert.Contains(t, err.Error(), "GetInitStateError")
	})

	t.Run("GetState", func(t *testing.T) {
		cpes.SetError("GetState", "GetStateError")
		_, err = cpe.GetState("blah", "blah")
		assert.Contains(t, err.Error(), "GetStateError")
	})

	t.Run("PostState", func(t *testing.T) {
		cpes.SetError("PostState", "PostStateError")
		err = cpe.PostState("blah", "blah", "blah", map[string]interface{}{})
		assert.Contains(t, err.Error(), "PostStateError")
	})

	t.Run("StorePut", func(t *testing.T) {
		cpes.SetError("StorePut", "StorePutError")
		err = cpe.StorePut("blah", "blah", "blah", []byte{})
		assert.Contains(t, err.Error(), "StorePutError")
	})

	t.Run("StoreGet", func(t *testing.T) {
		cpes.SetError("StoreGet", "StoreGetError")
		_, err = cpe.StoreGet("blah", "blah", "blah", true)
		assert.Contains(t, err.Error(), "StoreGetError")
	})

	t.Run("ClearStore", func(t *testing.T) {
		cpes.SetError("ClearStore", "ClearStoreError")
		err = cpe.ClearStore()
		assert.Contains(t, err.Error(), "ClearStoreError")
	})

	t.Run("PutValue", func(t *testing.T) {
		cpes.SetError("PutValue", "PutValueError")
		err = cpe.PutValue("blah", "blah", "blah")
		assert.Contains(t, err.Error(), "PutValueError")
	})

	t.Run("GetValue", func(t *testing.T) {
		cpes.SetError("GetValue", "GetValueError")
		cfg := cpe.GetValue("blah", "blah", "blah")
		assert.Contains(t, cfg.err.Error(), "GetValueError")
	})
}

func TestConfigProviderErrorablePassthrough(t *testing.T) {

	d := "sdf"
	s := "fds"

	cpe, err := GetConfigProvider("error/thing=thang/:memory://", AUTH_ADMIN_TOKEN_DEFAULT)

	t.Run("FetchConfig", func(t *testing.T) {
		c := cpe.FetchConfig("blah")
		assert.Contains(t, c.Err.Error(), "Config not found")
	})

	t.Run("PostConfig", func(t *testing.T) {
		c := Config{}
		err = cpe.PostConfig(&c, "blah")
		assert.NoError(t, err)
	})

	t.Run("RemoveConfig", func(t *testing.T) {
		err = cpe.RemoveConfig("blah")
		assert.NoError(t, err)
	})

	t.Run("RegisterService", func(t *testing.T) {
		_, err = cpe.RegisterService(ITLocal, s, d, "sdf", "dsf", "Sdf",
			"sdf", 1, map[string]interface{}{})
		assert.NoError(t, err)
	})

	t.Run("RegisterInstance", func(t *testing.T) {
		_, err = cpe.RegisterInstance(ITLocal, s, d, "sdf", "dsf", "Sdf", "sdf")
		assert.NoError(t, err)
	})

	t.Run("PortalAdd", func(t *testing.T) {
		_, err = cpe.PortalAdd(d, "blah", "blah")
		assert.NoError(t, err)
	})

	t.Run("DiscoveryAdd", func(t *testing.T) {
		err = cpe.DiscoveryAdd(d, "blah", "blah")
		assert.NoError(t, err)
	})

	t.Run("DiscoverService", func(t *testing.T) {
		_, err = cpe.DiscoverService(d, "blah")
		assert.NoError(t, err)
	})

	t.Run("DiscoverPortal", func(t *testing.T) {
		_, err = cpe.DiscoverPortal(d, "blah")
		assert.NoError(t, err)
	})

	t.Run("PortalRemove", func(t *testing.T) {
		err = cpe.PortalRemove(d, "blah")
		assert.NoError(t, err)
	})

	t.Run("GetInitState", func(t *testing.T) {
		_, err = cpe.GetInitState(d, s, "blah")
		assert.NoError(t, err)
	})

	t.Run("PostState", func(t *testing.T) {
		err = cpe.PostState(d, s, "blah", map[string]interface{}{})
		assert.NoError(t, err)
	})

	t.Run("GetState", func(t *testing.T) {
		_, err = cpe.GetState(d, s)
		assert.NoError(t, err)
	})

	t.Run("StorePut", func(t *testing.T) {
		err = cpe.StorePut("blah", "blah", "blah", []byte{})
		assert.Contains(t, err.Error(), "Memory only cannot use the store")
	})

	t.Run("StoreGet", func(t *testing.T) {
		_, err = cpe.StoreGet("blah", "blah", "blah", true)
		assert.Contains(t, err.Error(), "Memory only cannot use the store")
	})

	t.Run("PutValue", func(t *testing.T) {
		err = cpe.PutValue("blah", "blah", "blah")
		assert.NoError(t, err)
	})

	t.Run("GetVallue", func(t *testing.T) {
		cfg := cpe.GetValue("blah", "", "blah")
		assert.NoError(t, cfg.GetValueError())
	})

	t.Run("ClearStore", func(t *testing.T) {
		err = cpe.ClearStore()
		assert.Contains(t, err.Error(), "Memory only cannot use the store")
	})

	t.Run("DeregisterService", func(t *testing.T) {
		err = cpe.DeregisterService(s, d)
		assert.NoError(t, err)
		cpe.SetConfigAccessAuth(&d)
	})

}
