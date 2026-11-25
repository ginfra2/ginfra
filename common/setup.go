/*
Package common
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Setup for the app and for testing.
*/
package common

import (
	"gitlab.com/ginfra/ginfra/base"
	"os"
)

var (
	setupComplete     = false
	testSetupComplete = false
)

// Setup initializes the global environment by invoking required dependencies and ensures it is done only once.
// It panics if there is an error during the initialization process.
func Setup() {
	// TODO: Warning! I don't know what will happen if tests are parallelized but I don't want to prematurely gate this.

	var err error = nil

	if !setupComplete {
		err = base.LoadGinfra()
	}

	if err != nil {
		panic(err)
	}
	setupComplete = true
}

// SetupTest initializes the test setup by calling the Setup function and creating a test directory.
// If the test setup has already been completed, it does nothing.  You will likely have to add this to your
// package func init() if you plan on using any of the cooked values in values.go.
func SetupTest() {

	Setup()

	var err error = nil
	if !testSetupComplete {
		err = os.MkdirAll(base.GetGinfraTmpTestDir(), os.ModePerm)
	}

	if err != nil {
		panic(err)
	}
	testSetupComplete = true
}
