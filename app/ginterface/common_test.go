/*
Package ginterfaces
Copyright (c) 2023, 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Common code tests.
*/
package ginterfaces

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"os"
	"path/filepath"
	"testing"
)

func TestGinterfaceCommonCases(t *testing.T) {

	t.Run("checkApiDir", func(t *testing.T) {
		p := base.GetGinfraTmpTestDir()
		_ = os.MkdirAll(filepath.Join(p, "api"), 0644)
		assert.NoError(t, checkApiDir(p))
		_, err := os.Stat(p)
		assert.NoError(t, err)
	})

	t.Run("checkApiDir file", func(t *testing.T) {
		p := filepath.Join(base.GetGinfraHome(), "app", "ginterface", "test", "stubtest")
		assert.ErrorContains(t, checkApiDir(p), "File exists")
	})

}
