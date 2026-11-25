/*
Package base
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Values and helper functions to find common values use by the other packages.
*/
package base

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// CONFIGURATION POINTS
// There are dependant scripts that are sensitive to any changes to these.  BE CAREFUL.
const (
	ConfigAuthLength      = 20
	ConfigFileName        = "service_config.json"
	ConfigMountDir        = "/config"
	DebuggerFlagFileName  = "DEBUGGER"
	NoServiceFlagFileName = "NOSERVICE"

	ConfigAuth        = "auth" // Authorization to manage service.
	ConfigUri         = "config_uri"
	ConfigPortalUri   = "config_portal_uri"
	ConfigAccessToken = "config_access_token" // Access token to config service.
	ConfigDeployment  = "deployment"
	ConfigName        = "name"
	ConfigType        = "type"
	ConfigPort        = "port"
	ConfigWebPort     = "web_port"
	ConfigSpecifics   = "specifics"
	ConfigId          = "id" // Specific instance Id.

	ConfigClear = "clear"
)

// SPECIFICS

const SpecificsFileName = "specifics.json"

// GENERAL

const ShortIdLength = 12
const DefaultWebPort = "8080"

// SERVICE
const (
	ConfigServiceName = "configservice"
)

// # ----------------------------------------------------------------------------------------------------------------
// # This is meant for the services, but other things can use it.

var (
	executableDirPath string
	registeredService string
)

// SetupValues initializes values typically used by services and should be called once before any other methods below
// it.  It returns an error if an error occurs.
func SetupValues() error {

	var (
		err      error = nil
		execPath string
	)

	// Get root directory
	if execPath, err = os.Executable(); err == nil {
		executableDirPath = filepath.Dir(execPath)
	}
	return err
}

// GetExecutableDirPath  Not sure that I want to keep this.  Be careful using it outside the services.  Also note that
// while your IDE may say it is not used, it could very well be used in the services that consume this module.
func GetExecutableDirPath() string {
	return executableDirPath
}

// RegisterService registers the running instance as a service by its name. Only one service can be registered per
// instance; otherwise, it panics.
func RegisterService(name string) {
	if registeredService != "" {
		panic("Only one service can be registered per instance.")
	}
	registeredService = name
}

// GetRegisterService retrieves the name of the currently registered service, if any. Returns an empty string if none
// is registered.
func GetRegisterService() string {
	return registeredService
}

// #####################################################################################################################
// # LOADERS AND CLIENTS FOR GENERALLY IMMUTABLE VALUES

const EnvGinfraHome = "GINFRA_HOME"
const EnvGinfraTestHome = "GINFRA_TEST_HOME"
const PathTmpDir = "tmp"
const PathTmpTestDir = "test"
const PathCacheDir = "cache"
const PathGinterfacesDir = "ginterfaces"
const PathServicesDir = "services"
const PathTemplatesDir = "_templates"

var (
	ginfraHomeDir = ""
)

// # ----------------------------------------------------------------------------------------------------------------
// # This is meant for the Ginfra executable, but others might have a use for it.

// LoadGinfra initializes the ginfra executable.  It requires the GINFRA_HOME environment variable is set.  It expects
// it to be pointing to the location of the checked out ginfra repo.  However, unless you are running the provided
// tests, it would very likely be ok to point elsewhere.  Returns an error if an error occurs.
// IMPORTANT: This function must remain either reentrant or locked, otherwise mass tests will fail.
func LoadGinfra() error {
	var err error = nil

	if ginfraHomeDir == "" {
		// Normally using environment variables in code is a bad idea, but
		// this one has to be set for anything else to work.
		ginfraHomeDir = os.Getenv(EnvGinfraHome)
		if ginfraHomeDir == "" {
			err = errors.New(fmt.Sprintf("Required environment variable %s is not set.", EnvGinfraHome))
		}
	}

	return err
}

// GetGinfraHome gets the GINFRA_HOME path.  If you plan on using this or any below that use this, LoadGinfra() must
// be called first to initialize the value.  Anything under the application (/app) will have already done this during
// application startup.
func GetGinfraHome() string {
	return ginfraHomeDir
}

// GetGinfraTmpDir returns the path to a temporary file directory in the GINFRA_HOME path.
func GetGinfraTmpDir() string {
	return filepath.Join(GetGinfraHome(), PathTmpDir)
}

// GetGinfraRandomTmpDir generates a new random temporary directory path under the GINFRA_HOME temporary file directory.
func GetGinfraRandomTmpDir() string {
	return filepath.Join(GetGinfraHome(), PathTmpDir,
		PathTmpDir+"_"+GenerateCodeword(10))
}

// GetGinfraTmpCacheDir returns the path to the temporary cache directory under the GINFRA_HOME temporary file directory.
func GetGinfraTmpCacheDir() string {
	return filepath.Join(GetGinfraHome(), PathTmpDir, PathCacheDir)
}

// GetGinfraTmpTestDir returns the path to a temporary test directory within the GINFRA_HOME temporary directory.  This
// is used by the provided tests.  Most other uses should use GetGinfraRandomTmpDir instead.
func GetGinfraTmpTestDir() string {
	return filepath.Join(GetGinfraTmpDir(), PathTmpTestDir)
}

// GetGinfraRandomTmpTestDir generates a new random temporary directory path under the GINFRA_HOME temporary test file directory.
func GetGinfraRandomTmpTestDir() string {
	return filepath.Join(GetGinfraTmpTestDir(), PathTmpDir+"_"+GenerateCodeword(10))
}

// GetGinfraRandomTmpTestDirAndMake generates a new random temporary directory path under the GINFRA_HOME temporary test file directory.
// It will make the directory on the filesystem.
func GetGinfraRandomTmpTestDirAndMake() string {
	d := GetGinfraRandomTmpTestDir()
	if err := os.MkdirAll(d, 0777); err != nil {
		panic(err)
	}
	return d
}

// GetGinfraGinterfacesDir constructs and returns the path to the ginterfaces directory within GINFRA_HOME.
func GetGinfraGinterfacesDir() string {
	return filepath.Join(GetGinfraHome(), PathGinterfacesDir)
}

// GetGinfraTemplatesDir returns the path to the templates directory within the GINFRA_HOME directory structure.
func GetGinfraTemplatesDir() string {
	return filepath.Join(GetGinfraHome(), PathServicesDir, PathTemplatesDir)
}

// SetGinfraHome is provided for testing only.
func SetGinfraHome(p string) {
	ginfraHomeDir = p
}

// StringPointer the AI generates tests like this for some reason.
func StringPointer(v string) *string {
	return &v
}
