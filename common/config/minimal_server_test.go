//go:build unit_test

/*
Package shared
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Configuration tests.
*/

package config

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/api/service_giconfig/models"
	test "gitlab.com/ginfra/ginfra/common/test"
	"testing"
)

func TestMinimalServerCases(t *testing.T) {

	t.Run("provider panic", func(t *testing.T) {
		if test.SkipFlakyCheck() {
			return
		}
		assert.Panics(t, func() {
			var ms MinimalConfigService
			ms.homepath = "/dev/null/../../../../../../../../../../../../../../../../../../../../../../../../../../../../../../../"
			_ = getConfigStoreProvider(&ms)
		})
	})
	t.Run("stop panic", func(t *testing.T) {
		assert.Panics(t, func() {
			var ms MinimalConfigService
			_ = StopService(&ms)
		})
	})
	t.Run("return error", func(t *testing.T) {
		err := errors.New("XXX")
		err2 := processError(nil, err, "xxx")
		assert.Equal(t, err, err2)
	})
	t.Run("bad base64 leads to panic", func(t *testing.T) {
		b := "2haef29223fh9ihe9wv8ch2Q@)(#@R*FUWJWEF)IU@JF"
		m := models.ConfigSetItem{Block: &b}
		assert.Panics(t, func() { _ = giConfigSetConfigHandleItem(nil, nil, &m) })
	})
}
