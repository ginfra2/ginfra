//go:build functional_test

/*
Package app
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Setup tests.
*/

package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/ginfra/ginfra/base"
	test "gitlab.com/ginfra/ginfra/common/test"
	"os"
	"testing"
)

func TestSetupTestNeedModFile(t *testing.T) {

	t.Run("No mod file", func(t *testing.T) {
		ghd := base.GetGinfraHome()

		defer func() {
			r := recover()
			require.NotNil(t, r, "The code did not panic")
			_ = os.Setenv(base.EnvGinfraHome, ghd)
			base.SetGinfraHome(ghd)

		}()

		base.SetGinfraHome("/")
		SetupTestNeedModFileTest()

	})

	t.Run("Bad path", func(t *testing.T) {
		if !test.SkipFlakyCheck() {
			assert.Panics(t, func() {
				SetupTestNeedModFile("../../.../../../../../../../.../../../../../../../../../../../.../../../../../../../../..")
			})

		}
	})
}
