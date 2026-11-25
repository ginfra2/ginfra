//go:build functional_test

/*
Package log
Copyright (c) 2024 Erich Gatejen
[LICENSE]

Logging facility test.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package log

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	test "gitlab.com/ginfra/ginfra/common/test"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	common.SetupTest()
}

func TestLog(t *testing.T) {

	t.Run("Do logging", func(t *testing.T) {
		l := GetLogger(base.GetGinfraTmpTestDir(), true)
		if l == nil {
			t.Error("Nil logger")
			return
		}
		l.Debug("This is a debugging log entry.  aaaaaaaa bbbbbbb")
	})

	t.Run("Verify logging", func(t *testing.T) {
		s, err := os.ReadFile(filepath.Join(base.GetGinfraTmpTestDir(), BaseFileName))
		scan, err := test.NewScanPack(string(s), err)
		if err != nil {
			t.Error(err)
			return
		}

		assert.Greater(t, scan.Seek("debugging log entry"), 0)

	})
}

func TestGLoggerWriter(t *testing.T) {

	t.Run("Write", func(t *testing.T) {
		tests := []struct {
			name   string
			input  string
			hasErr bool
			output int
		}{
			{
				name:   "Valid JSON input",
				input:  `{"key":"value"}`,
				hasErr: false,
				output: len(`{"key":"value"}`),
			},
			{
				name:   "Invalid JSON input",
				input:  `invalid_json`,
				hasErr: true,
				output: 0,
			},
			{
				name:   "Empty JSON input",
				input:  `{}`,
				hasErr: false,
				output: len(`{}`),
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				mockSlog := slog.New(slog.NewTextHandler(io.Discard, nil))
				writer := NewGLoggerWriter(mockSlog)

				bytesWritten, err := writer.Write([]byte(tt.input))

				if tt.hasErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.output, bytesWritten)
				}
			})
		}
	})

}

func TestAddSlogCore(t *testing.T) {
	t.Run("AddSlogCore", func(t *testing.T) {
		tests := []struct {
			name       string
			logger     *slog.Logger
			expectCore bool
		}{
			{
				name:       "Valid slog.Logger",
				logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
				expectCore: true,
			},
			{
				name:       "Nil slog.Logger",
				logger:     nil,
				expectCore: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				glogger := NewGLogger(false)
				glogger.AddSlogCore(tt.logger)

				if tt.expectCore {
					assert.NotEmpty(t, glogger.cores)
				} else {
					assert.Empty(t, glogger.cores)
				}
			})
		}
	})
}
