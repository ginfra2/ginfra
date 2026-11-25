//go:build unit_test

/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Unit testing for the file configuration provider.
*/
package config

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	ctest "gitlab.com/ginfra/ginfra/common/test"
	"os"
	"path/filepath"
	"testing"
)

// ===============================================================================================================
// =

const TestConfigPath1 = "config1"
const TestConfigPath2 = "config2"

var (
	ftoken1 string
)

// tcpfGetStoreFilePath returns the test asset dir and the test target dir.
func tcpfGetStoreFilePaths() (string, string) {
	if err := base.LoadGinfra(); err != nil {
		panic(err)
	}
	return filepath.Join(base.GetGinfraHome(), "common", "config", "test"), base.GetGinfraTmpTestDir()
}

func tcpf1RegisterService(t *testing.T, p ConfigProvider, tt string) error {
	var (
		err error
		c   map[string]interface{}
	)

	sspc := make(map[string]interface{})
	sspc[s1Name] = s1Value

	if ftoken1, err = p.RegisterService(ITLocal, TCPSERVICE_NAME, TCPSERVICE_DEPLOYMENT, "BBBBBBBBBBB", "image",
		"http://localhost:5111", TCPSERVICE_SERVICECLASS, 8900, sspc); err != nil {
		return err
	}
	if _, err = p.RegisterInstance(ITLocal, TCPSERVICE_NAME, TCPSERVICE_DEPLOYMENT, token1, "aaa", "127.0.0.1",
		"localhost"); err != nil {
		return err
	}
	if err = p.DiscoveryAdd(TCPSERVICE_DEPLOYMENT, TCPSERVICE_SERVICECLASS, "http://localhost:8900"); err != nil {
		return err
	}

	if err = common.ReadJsonFileMap(filepath.Join(tt, TestConfigPath1, getSaveFileName(4)), "config", &c); err != nil {
		return err
	}

	ctest.TestCheckMapValue(t, false, c, TCPSERVICE_DEPLOYMENT, CR_DEPLOYMENTS, TCPSERVICE_DEPLOYMENT)
	ctest.TestCheckMapValue(t, false, c, "localhost", CR_DEPLOYMENTS, TCPSERVICE_DEPLOYMENT, CRD_DISCOVERY, CRD_PORTALS,
		CO_ANY_PORTAL)
	ctest.TestCheckMapValue(t, false, c, "http://localhost:8900", CR_DEPLOYMENTS, TCPSERVICE_DEPLOYMENT, CRD_DISCOVERY,
		CRD_SERVICES, TCPSERVICE_SERVICECLASS)

	tm := ctest.TestCheckMapValue(t, false, c, TCPSERVICE_NAME, CR_DEPLOYMENTS, TCPSERVICE_DEPLOYMENT, CRD_SERVICES, TCPSERVICE_NAME)

	tmi := ctest.TestCheckMapValue(t, false, tm, "aaa", CRD_SERVICES__MANAGEMENT, CRD_MSERVICE__INSTANCE, "aaa")
	ctest.TestCheckMapValue(t, false, tmi, "localhost", CRD_MSERVICE__INSTANCE__FQDN)
	is, _ := GetStringFromInstanceType(ITLocal)
	ctest.TestCheckMapValue(t, false, tmi, is, CRD_MSERVICE__INSTANCE__TYPE)
	ctest.TestCheckMapValue(t, false, tmi, "aaa", CRD_MSERVICE__INSTANCE__ID)
	ctest.TestCheckMapValue(t, false, tmi, "127.0.0.1", CRD_MSERVICE__INSTANCE__IP)

	ctest.TestCheckMapValue(t, false, tm, ftoken1, CRD_SERVICES__SERVICE, CRD_SERVICES__AUTH)
	ctest.TestCheckMapValue(t, false, tm, "BBBBBBBBBBB", CRD_SERVICES__SERVICE, CRD_SERVICES__SERVICE_AUTH)

	tmc := ctest.TestCheckMapValue(t, false, c, CRD_SSERVICE__CONFIG, CR_DEPLOYMENTS, TCPSERVICE_DEPLOYMENT, CRD_SERVICES,
		TCPSERVICE_NAME, CRD_SERVICES__SERVICE, CRD_SSERVICE__CONFIG)
	ctest.TestCheckMapValue(t, false, tmc, "http://localhost:5111", CRD_SSERVICE__CONFIG__CONFIG_PROVIDER)
	ctest.TestCheckMapValue(t, false, tmc, TCPSERVICE_SERVICECLASS, CRD_SSERVICE__CONFIG__SERVICE_CLASS)
	ctest.TestCheckMapValue(t, false, tmc, 8900, CRD_SSERVICE__CONFIG__PORT)
	ctest.TestCheckMapValue(t, false, tmc, "image", CRD_SSERVICE__CONFIG__IMAGE)

	return nil
}

func tcpf1State(t *testing.T, p ConfigProvider) error {
	s, err := p.GetInitState(TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, "aaa")
	if err != nil {
		return err
	}
	assert.Equal(t, InitNew, int(s))

	if s, err = p.GetInitState(TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, "aaa"); err != nil {
		return err
	}
	assert.Equal(t, InitPending, int(s))

	if s, err = p.GetInitState(TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, "bbb"); err != nil {
		return err
	}
	assert.Equal(t, InitPending, int(s))

	if err = p.PostState(TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, "aaa", map[string]interface{}{"bork": "bork"}); err != nil {
		return err
	}

	if s, err = p.GetInitState(TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, "bbb"); err != nil {
		return err
	}
	assert.Equal(t, InitDone, int(s))

	var m map[string]interface{}
	if m, err = p.GetState(TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME); err != nil {
		return err
	}
	assert.Equal(t, "bork", m["bork"])

	return nil
}

func tcpf1Values(t *testing.T, p ConfigProvider) error {
	err := p.PutValue(TCPSERVICE_DEPLOYMENT, "name1", "value1")
	if err != nil {
		return err
	}

	cv := p.GetValue(TCPSERVICE_DEPLOYMENT, "", "name1")
	assert.Equal(t, "value1", cv.GetValueStringNoerr())
	return nil
}

func TestSavableFileConfigProviderRegistration(t *testing.T) {
	_, tt := tcpfGetStoreFilePaths()

	err := ctest.FreshTargetDir(filepath.Join(tt, TestConfigPath1))
	if err != nil {
		t.Error(err)
		return
	}

	ar := AUTH_ADMIN_TOKEN_DEFAULT
	p, _ := GetConfigProviderFileSavable(filepath.Join(tt, TestConfigPath1), 5, true)
	p.SetConfigAccessAuth(&ar)

	t.Run("register service", func(t *testing.T) {
		if err := tcpf1RegisterService(t, p, tt); err != nil {
			t.Error(err)
		}
	})

	t.Run("service state", func(t *testing.T) {
		if err := tcpf1State(t, p); err != nil {
			t.Error(err)
		}
	})

	t.Run("deployment values", func(t *testing.T) {
		if err := tcpf1Values(t, p); err != nil {
			t.Error(err)
		}
	})

}

func tcpf2Store(t *testing.T, p ConfigProvider, data []byte) error {

	sspc := make(map[string]interface{})
	sspc[s1Name] = s1Value
	a, err := p.RegisterService(ITLocal, TCPSERVICE_NAME, TCPSERVICE_DEPLOYMENT, "BBBBBBBBBBB", "image",
		"http://localhost:5111", TCPSERVICE_SERVICECLASS, 8900, sspc)
	if err != nil {
		return err
	}

	ar := AUTH_ADMIN_TOKEN_DEFAULT
	p.SetConfigAccessAuth(&ar)

	if err = p.StorePut("bork1", TCPSERVICE_DEPLOYMENT, "", data); err != nil {
		return err
	}
	if _, err = p.StoreGet("bork1", TCPSERVICE_DEPLOYMENT, "", true); err != nil {
		return err
	}
	if err = p.StorePut("bork2", TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, data); err != nil {
		return err
	}
	if _, err = p.StoreGet("bork2", TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, true); err != nil {
		return err
	}
	if err = p.StorePut("bork3", "", TCPSERVICE_NAME, data); err != nil {
		return err
	}
	if _, err = p.StoreGet("bork3", "", TCPSERVICE_NAME, true); err != nil {
		return err
	}
	if err = p.StorePut("bork4", "", "", data); err != nil {
		return err
	}
	if _, err = p.StoreGet("bork4", "", "", true); err != nil {
		return err
	}

	_, err = p.StoreGet("bork4", TCPSERVICE_DEPLOYMENT, "", true)
	assert.Error(t, err)
	_, err = p.StoreGet("bork3", TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, true)
	assert.Error(t, err)
	_, err = p.StoreGet("bork2", "", TCPSERVICE_NAME, true)
	assert.Error(t, err)
	_, err = p.StoreGet("bork1", "", "", true)
	assert.Error(t, err)

	p.SetConfigAccessAuth(&a)
	_, err = p.StoreGet("bork1", TCPSERVICE_DEPLOYMENT, "", true)
	assert.Error(t, err)
	_, err = p.StoreGet("bork2", TCPSERVICE_DEPLOYMENT, TCPSERVICE_NAME, true)
	assert.NoError(t, err)
	_, err = p.StoreGet("bork3", "", TCPSERVICE_NAME, true)
	assert.Error(t, err)
	_, err = p.StoreGet("bork4", "", "", true)
	assert.Error(t, err)

	return nil
}

func TestSavableFileConfigProviderStore(t *testing.T) {
	ta, tt := tcpfGetStoreFilePaths()

	err := ctest.FreshTargetDir(filepath.Join(tt, TestConfigPath2))
	if err != nil {
		t.Error(err)
		return
	}

	p, _ := GetConfigProviderFileSavable(filepath.Join(tt, TestConfigPath2), 5, true)

	d, err := os.ReadFile(filepath.Join(ta, "source_store.txt"))
	if err != nil {
		t.Error(err)
		return
	}

	if err = tcpf2Store(t, p, d); err != nil {
		t.Error(err, fmt.Sprintf("TA: %s", ta))
		return
	}

}

func TestCookedConfig(t *testing.T) {
	auth := AUTH_ADMIN_TOKEN_DEFAULT
	ta, _ := tcpfGetStoreFilePaths()

	cp, err := GetConfigProviderFileSavable(filepath.Join(ta, "config.json"), 0, true)
	if err != nil {
		t.Error(err)
		return
	}
	cp.SetConfigAccessAuth(&auth)

	g, err := cp.DiscoverPortal("default", "giuser")
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, "http://node1:8900/", g)
}

func TestLoadGetConfigProviderMultiFile(t *testing.T) {
	type testCase struct {
		Setup       func() (*configProviderFile, string)
		CleanUp     func(location string)
		ExpectedErr string
		Name        string
	}

	testCases := []testCase{
		{
			Name: "Store directory does not exist",
			Setup: func() (*configProviderFile, string) {
				location := filepath.Join(base.GetGinfraTmpTestDir(), "cpf", "bad1")
				return &configProviderFile{currentId: 0}, location
			},
			ExpectedErr: "Store directory is missing or corrupt.",
		},
		{
			Name: "Bad seed file name",
			Setup: func() (*configProviderFile, string) {
				location := filepath.Join(base.GetGinfraTmpTestDir(), "cpf", "1")
				if err := os.MkdirAll(location, 0755); err != nil {
					panic(err)
				}
				_ = os.WriteFile(filepath.Join(location, "config.json"), []byte(`{}`), 0644)
				return &configProviderFile{currentId: 0}, location
			},
			ExpectedErr: "Config filename corrupt",
		},
		{
			Name: "Bad seed file name number",
			Setup: func() (*configProviderFile, string) {
				location := filepath.Join(base.GetGinfraTmpTestDir(), "cpf", "2")
				if err := os.MkdirAll(location, 0755); err != nil {
					panic(err)
				}
				_ = os.WriteFile(filepath.Join(location, "config.A.json"), []byte(`{}`), 0644)
				return &configProviderFile{currentId: 0}, location
			},
			ExpectedErr: "not a proper number",
		},
		{
			Name: "Load good", // This one should always be last.
			Setup: func() (*configProviderFile, string) {
				location := filepath.Join(base.GetGinfraTmpTestDir(), "cpf", "config3")
				if err := common.CopyFiles(filepath.Join(base.GetGinfraHome(), "common", "config", "test", "config3"),
					location, false, false, false); err != nil {
					panic(nil)
				}
				return &configProviderFile{currentId: 0}, location
			},
			ExpectedErr: "",
		},
	}

	var (
		cpf      *configProviderFile
		location string
	)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			cpf, location = tc.Setup()
			err := loadGetConfigProviderMultiFile(cpf, location)
			if err == nil {
				if tc.ExpectedErr != "" {
					t.Errorf("ExpectedErr error but did not get one: %s", tc.ExpectedErr)
				}
			} else {
				if tc.ExpectedErr != "" {
					assert.ErrorContains(t, err, tc.ExpectedErr)
				} else {
					t.Error(err)
				}
			}
			if tc.CleanUp != nil {
				tc.CleanUp(location)
			}
		})
	}

	// The last one will be a good load.
}
