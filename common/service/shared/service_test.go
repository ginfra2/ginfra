package shared

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func init() {
	common.SetupTest()
}

func TestLoadServiceConfig(t *testing.T) {
	cases := []struct {
		name         string
		configString string
		expectErr    bool
	}{

		{
			name: "validConfig",
			configString: `{
					"clear": "true",
					"auth": "sampleAuth",
					"config_uri": "sampleUri",
					"config_portal_uri": "sampleUriPrimary",
					"config_access_token": "sampleToken",
					"deployment": "sampleDeploymentName",
					"name": "sampleName",
					"id": "sampleId",
					"type": "host"
				}`,
			expectErr: false,
		},
		{
			name: "invalidAuthType",
			configString: `{
					"clear": "true",
					"auth": 100,
					"config_uri": "sampleUri",
					"config_portal_uri": "sampleUriPrimary",
					"config_access_token": "sampleToken",
					"deployment": "sampleDeploymentName",
					"name": "sampleName",
					"id": "sampleId",
					"type": "host"
				}`,
			expectErr: true,
		},
		{
			name: "invalidCuriType",
			configString: `{
					"clear": "true",
					"auth": "sampleAuth",
					"config_uri": 100,
					"config_portal_uri": "sampleUriPrimary",
					"config_access_token": "sampleToken",
					"deployment": "sampleDeploymentName",
					"name": "sampleName",
					"id": "sampleId",
					"type": "sampleInstanceType"
				}`,
			expectErr: true,
		},
		{
			name: "invalidCuripType",
			configString: `{
					"clear": "true",
					"auth": "sampleAuth",
					"config_uri": "sampleUri",
					"config_portal_uri": 100,
					"config_access_token": "sampleToken",
					"deployment": "sampleDeploymentName",
					"name": "sampleName",
					"id": "sampleId",
					"type": "host"
				}`,
			expectErr: true,
		},
		{
			name: "invalidCTokenType",
			configString: `{
					"clear": "true",
					"auth": "sampleAuth",
					"config_uri": "sampleUri",
					"config_portal_uri": "sampleUriPrimary",
					"config_access_token": 100,
					"deployment": "sampleDeploymentName",
					"name": "sampleName",
					"id": "sampleId",
					"type": "host"
				}`,
			expectErr: true,
		},
		{
			name: "invalidDeploymentType",
			configString: `{
					"clear": "true",
					"auth": "sampleAuth",
					"config_uri": "sampleUri",
					"config_portal_uri": "sampleUriPrimary",
					"config_access_token": "sampleToken",
					"deployment": 100,
					"name": "sampleName",
					"id": "sampleId",
					"type": "host"
				}`,
			expectErr: true,
		},
		{
			name: "invalidNameType",
			configString: `{
					"clear": "true",
					"auth": "sampleAuth",
					"config_uri": "sampleUri",
					"config_portal_uri": "sampleUriPrimary",
					"config_access_token": "sampleToken",
					"deployment": "sampleDeploymentName",
					"name": 100,
					"id": "sampleId",
					"type": "host"
				}`,
			expectErr: true,
		},
		{
			name: "invalidIdType",
			configString: `{
					"clear": "true",
					"auth": "sampleAuth",
					"config_uri": "sampleUri",
					"config_portal_uri": "sampleUriPrimary",
					"config_access_token": "sampleToken",
					"deployment": "sampleDeploymentName",
					"name": "sampleName",
					"id": 100,
					"type": "host"
				}`,
			expectErr: true,
		},
		{
			name: "invalidInstanceTType",
			configString: `{
				"clear": "true",
				"auth": "sampleAuth",
				"config_uri": "sampleUri",
				"config_portal_uri": "sampleUriPrimary",
				"config_access_token": "sampleToken",
				"deployment": "sampleDeploymentName",
				"name": "sampleName",
				"id": "sampleId",
				"type": "xxxx"
			}`,
			expectErr: true,
		},
	}

	fp := filepath.Join(base.GetGinfraTmpTestDir(), "service_config.json")

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := os.WriteFile(fp, []byte(tc.configString), os.ModePerm)
			if err != nil {
				t.Fatal(err)
			}

			_, err = LoadServiceConfig(base.GetGinfraTmpTestDir())
			if (err != nil) != tc.expectErr {
				t.Errorf("LoadServiceConfig() err = %v, expectErr = %v", err, tc.expectErr)
			}
		})
	}
}

func TestLoadSpecifics(t *testing.T) {
	home := base.GetGinfraTmpTestDir()

	t.Run("FileExistsInHome", func(t *testing.T) {
		err := os.WriteFile(path.Join(home, base.SpecificsFileName), []byte("{}"), 0666)
		assert.NoError(t, err)

		expected := "{}"
		got := LoadSpecifics(home)

		assert.Equal(t, expected, got)
	})

	t.Run("FileDoesNotExistAnywhere", func(t *testing.T) {
		assert.Panics(t, func() {
			LoadSpecifics(base.GetGinfraHome())
		}, "Expected panic when specifics file does not exist anywhere.")
	})
}
