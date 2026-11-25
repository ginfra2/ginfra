//go:build unit_test

/*
Package shared
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Configuration tests.
*/
package config

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	test "gitlab.com/ginfra/ginfra/common/test"
	"path/filepath"
	"testing"
)

func init() {
	common.SetupTest()
}

type testProcessConfigCase struct {
	Name     string
	Config   ProcessConfig
	Target   string
	ExpError string
	Scan     []string
}

var testProcessConfigCases = []testProcessConfigCase{
	{
		Name:   "ProcessConfigString Set 1",
		Config: ProcessConfig{Deployment: "default", Service: "server_otelcol"},
		Target: `configs:
			1:
				server_name: serverbork
				server_port: <<<$deployments.%deployment.services.%SERVICE.service.config.port>>>
			2:
				server_name: server3`,
		ExpError: "",
		Scan:     []string{"server_name", "server_port: 8900"},
	},
	{
		Name:   "ProcessConfigString Set 2 Unknown Directive",
		Config: ProcessConfig{},
		Target: `configs:
			1:
				server_port: <<<#deployments.%deployment.services.%SERVICE.service.config.port>>>
				server_name: serverbork`,
		ExpError: "Unknown directive",
		Scan:     []string{}, // Guard against failure to error
	},
	{
		Name:   "ProcessConfigString Set 3 Bad Path",
		Config: ProcessConfig{},
		Target: `configs:
			1:
				server_port: <<<$deployments.%deployment.%SERVICE.service.config.port>>>
				server_name: serverbork`,
		ExpError: "Config not found",
		Scan:     []string{},
	},
	{
		Name:   "ProcessConfigString Set 4 Bad Config Point Directive",
		Config: ProcessConfig{},
		Target: `configs:
			1:
				server_port: <<<$deployments.%deployment.services.%BORK.service.config.port>>>
				server_name: serverbork`,
		ExpError: "Unknown configuration point directive",
		Scan:     []string{},
	},
	{
		Name:   "ProcessConfigString Set 5 Service Discover",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
			1:
				server_url: <<<=%deployment.server_otelcol/collector>>>
				server_name: serverbork`,
		ExpError: "",
		Scan:     []string{"grpc"},
	},
	{
		Name:   "ProcessConfigString Set 6 Service Discover Bad Not Found",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
			1:
				server_url: <<<=%deployment.server_otelcol/askask>>>
				server_name: serverbork`,
		ExpError: "Service not discovered",
		Scan:     []string{""},
	},
	{
		Name:   "ProcessConfigString Set 7 Portal Discover",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
			1:
				server_url: <<<*%deployment.server_otelcol/collector>>>
				server_name: serverbork`,
		ExpError: "",
		Scan:     []string{"grpc:", "node1"},
	},
	{
		Name:   "ProcessConfigString Set 8 Portal Discover Bad Not Found",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
			1:
				server_url: <<<*%deployment.server_otelcol/askask>>>
				server_name: serverbork`,
		ExpError: "Portal not discovered",
		Scan:     []string{},
	},
	{
		Name:   "ProcessConfigString Set 9 Double error and bad config directive",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
			1:
				server_url: <<<$%deployment.%monkey.askask>>>
				server_name: %thing`,
		ExpError: "Unknown configuration point directive",
		Scan:     []string{},
	},
	{
		Name:   "ProcessConfigString Set 10 Fetch Management Auth",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
			1:
				server_url: <<<+%deployment.server_otelcol>>>
				server_name: serverbork`,
		ExpError: "",
		Scan:     []string{"U7698AEPCGI1ZDVFVDPZ"},
	},
	{
		Name:   "ProcessConfigString Set 11 Env variable -- not found",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
			1:
				server_url: <<<^BORKBORKBORKBORKBORKBORKBORKBORK>>>
				server_name: serverbork`,
		ExpError: "",
		Scan:     []string{"server_url"},
	},
	{
		Name:   "processConfigPartDiscoverService config path error",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
		1:
			server_url: <<<=%monkey>>>
			server_name: %thing`,
		ExpError: "Unknown configuration point directive",
		Scan:     []string{},
	},
	{
		Name:   "processConfigPartDiscoverPortal config path error",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
		1:
			server_url: <<<*%monkey>>>
			server_name: %thing`,
		ExpError: "Unknown configuration point directive",
		Scan:     []string{},
	},
	{
		Name:   "processConfigPartDiscoverPortal config path error",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
		1:
			server_url: <<<*%monkey>>>
			server_name: %thing`,
		ExpError: "Unknown configuration point directive",
		Scan:     []string{},
	},
	{
		Name:   "processConfigPartFetchMngAuth config path error",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
		1:
			server_url: <<<+%monkey>>>
			server_name: %thing`,
		ExpError: "Unknown configuration point directive",
		Scan:     []string{},
	},
	{
		Name:   "processConfigPartServicePort config path error",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
		1:
			server_url: <<<&%monkey>>>
			server_name: %thing`,
		ExpError: "Unknown configuration point directive",
		Scan:     []string{},
	},
	{
		Name:   "multiple errors",
		Config: ProcessConfig{Deployment: "default"},
		Target: `configs:
	1:
		server_url: <<<&%monkey>>>
		server_name: <<<+%monkey>>>`,
		ExpError: "2 errors occurred",
		Scan:     []string{},
	},
}

func TestProcessConfigString(t *testing.T) {
	p, err := GetConfigProviderInMemory(filepath.Join(base.GetGinfraHome(), "common", "config", "test", "1.json"))
	if err != nil {
		t.Error(err)
		return
	}

	var sc *test.ScanPack
	for _, c := range testProcessConfigCases {

		t.Run(c.Name, func(t *testing.T) {
			sc, err = test.NewScanPack(ProcessConfigString(p, c.Config, c.Target))
			if err == nil {
				if c.ExpError != "" {
					t.Error("Expected error but did not get one.")
				}
				assert.Greater(t, sc.Seeks(c.Scan...), 0)
			} else {
				if c.ExpError == "" {
					t.Errorf("Did not expect error but got one: %s", err.Error())
				} else {
					assert.ErrorContains(t, err, c.ExpError)
				}
			}
		})

	}
}

var testProcessConfigString21 = `receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:
    send_batch_size: 10000
    timeout: 10s

exporters:
  otlp/uptrace:
    endpoint: <<<=%deployment.uptrace/uptracegrpc_endpoint>>>
    tls:
      insecure: true
    headers:
      uptrace-dsn: '<<<=%deployment.uptrace/uptrace_dsn>>>'

extensions:
  health_check:
  pprof:
  zpages:

service:
  extensions: [health_check, pprof, zpages]
  pipelines:
    traces:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/uptrace]
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/uptrace]
    logs:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/uptrace]
`

func TestProcessConfigString2(t *testing.T) {
	p, err := GetConfigProviderInMemory(filepath.Join(base.GetGinfraHome(), "common", "config", "test", "2.json"))
	if err != nil {
		t.Error(err)
		return
	}

	var sc *test.ScanPack

	t.Run("User case 1", func(t *testing.T) {
		sc, err = test.NewScanPack(ProcessConfigString(p, ProcessConfig{Deployment: "default"}, testProcessConfigString21))
		if err == nil {
			assert.Greater(t, sc.Seeks("grpc=14317"), 0)
		} else {
			t.Error(err)
		}
	})

}

func TestProcessConfigCases(t *testing.T) {

	cmp, err := GetConfigProviderInMemory(filepath.Join(base.GetGinfraHome(), "common", "config", "test", "1.json"))
	if err != nil {
		panic("Test is broken")
	}

	t.Run("processConfigPartDiscoverService poorly formed", func(t *testing.T) {
		_, err = processConfigPartDiscoverService(cmp, &ProcessConfig{Deployment: "default"}, "xxxx.yyy.zzz")
		assert.ErrorContains(t, err, "Poorly formed Discover Service point")
	})
	t.Run("processConfigPartDiscoverPortal poorly formed", func(t *testing.T) {
		_, err = processConfigPartDiscoverPortal(cmp, &ProcessConfig{Deployment: "default"}, "xxxx.yyy.zzz")
		assert.ErrorContains(t, err, "Poorly formed Discover Portal point")
	})
	t.Run("processConfigPartFetchMngAuth poorly formed", func(t *testing.T) {
		_, err = processConfigPartFetchMngAuth(cmp, &ProcessConfig{Deployment: "default"}, "xxxx.yyy.zzz")
		assert.ErrorContains(t, err, "Poorly formed Fetch Management Auth Token point")
	})
}
