/*
Package app
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Setup for the app and/or testing.
*/
package app

import (
	"gitlab.com/ginfra/ginfra/base"
	"os"
	"path/filepath"
)

// SetupTestNeedModFileTest prepares a temporary test directory with a `go.mod` file by calling SetupTestNeedModFile.  This
// is intended for testing only.
func SetupTestNeedModFileTest() {
	SetupTestNeedModFile(base.GetGinfraTmpTestDir())
}

// SetupTestNeedModFile ensures the specified path exists and contains a `go.mod` file, creating it if necessary.  This
// is intended for testing only.
func SetupTestNeedModFile(path string) {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		panic(err)
	}

	var data []byte
	_, err = os.Stat(filepath.Join(path, "go.mod"))
	if os.IsNotExist(err) {
		if data, err = os.ReadFile(filepath.Join(base.GetGinfraHome(), "go.mod")); err == nil {
			err = os.WriteFile(filepath.Join(path, "go.mod"), data, 0644)
		}
	}
	if err != nil {
		panic(err)
	}
}
