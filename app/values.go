/*
Package app
Copyright (c) 2023, 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Various values.
*/
package app

import (
	"gitlab.com/ginfra/ginfra/base"
	"path/filepath"
)

// #####################################################################################################################
// # STATIC CONFIG

// NAMES

const HostNetworkNamePrefix = "ginfranet_"

const TestExpectedServiceName = "test"

// PATHS

const PathGinterfaces = "ginterfaces"
const PathGinterfacesCommon = "gi_common.yaml"
const PathStubFile = "stub.yaml"
const PathApiDir = "api"
const PathModelsDir = "/models"
const PathServerDir = "/server"
const PathClientDir = "/client"
const PathDebuggingLog = "debugging.log"
const PathServiceDefPath = ".service"
const PathBuildPath = "build"
const PathServiceExecutable = "service.exe"
const PathHostingFilePath = "Hosting.json"
const PathMetaFilePath = "meta"
const PathHostnameFile = "hostname"
const PathSpecificsFile = "specifics.json"

const ArgSeparator = "^"

const JsonIndent = "    "

// ENVIRONMENTS

// SYSTEM

const ExitErrorCode = 9

// SERVICE

const ServiceTypeLocal = "local"
const ServiceTypeHost = "host"
const ServiceConnectRetries = 4
const ServiceConnectRetryTime = 200

// Supported classes

// GetDebuggingLogPath returns the absolute path to the debugging log file based on the configured GINFRA_HOME directory.
func GetDebuggingLogPath() string {
	return filepath.Join(base.GetGinfraHome(), PathDebuggingLog)
}
