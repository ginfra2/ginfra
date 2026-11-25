//go:build functional_test

/*
Package common
Copyright (c) 2024 Erich Gatejen
[LICENSE]

Values and helper functions tests.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package base

import (
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestValuesSetup(t *testing.T) {
	err := SetupValues()
	if err != nil {
		t.Error(err)
		return
	}

	ep, err := os.Executable()
	if err != nil {
		panic(err)
	}

	assert.NotEqual(t, ep, GetExecutableDirPath())
}

func TestValuesSetupEnv(t *testing.T) {
	e := os.Environ()
	os.Clearenv()
	err := LoadGinfra()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Required environment variable")

	// Restore env
	for _, i := range e {
		nv := strings.SplitN(i, "=", 2)
		if len(nv) > 1 {
			_ = os.Setenv(nv[0], nv[1])
		} else {
			_ = os.Setenv(nv[0], "")
		}
	}
	err = LoadGinfra()
}

func registerPanic() (p bool) {
	p = false
	defer func() {
		if r := recover(); r != nil {
			p = true
		}
	}()

	RegisterService("bbb")

	return false
}

func TestValuesRegisterService(t *testing.T) {
	RegisterService("aaa")
	assert.Equal(t, "aaa", GetRegisterService())

	assert.Equal(t, true, registerPanic(), "Reregistration should have panicked.")
}

func TestValuesDir(t *testing.T) {
	assert.True(t, strings.HasSuffix(GetGinfraTmpCacheDir(), PathCacheDir))
}
