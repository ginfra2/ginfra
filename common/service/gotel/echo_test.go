//go:build functional_test

/*
Package gotel
Copyright (c) 2024 Erich Gatejen
[LICENSE]

OpenTelemetry.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package gotel

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	testutil "gitlab.com/ginfra/ginfra/common/test"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	common.SetupTest()
	_ = os.MkdirAll(filepath.Join(base.GetGinfraTmpTestDir(), "otel"), 0777)
}

func testOtelEchoFlush() {
	cfg := GetOtelConfig()
	_ = cfg.MeterProvider.ForceFlush(context.Background())
	_ = cfg.TracerProvider.ForceFlush(context.Background())
}

func TestOtelEcho(t *testing.T) {

	attr := []attribute.KeyValue{
		attribute.String("B1", "Binformation1"),
		attribute.String("B2", "Binformation2"),
	}

	s := echo.New()
	s.GET("/ok", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})
	s.GET("/happy", func(c echo.Context) error {
		cfg := GetOtelConfig()
		ct, _ := cfg.Meter.Int64Counter("happyhit", metric.WithDescription("Hit counts"))
		for i := 0; i < 100; i++ {
			ct.Add(context.Background(), 1, metric.WithAttributes(attr...))
		}
		return c.String(http.StatusOK, "OK")
	})

	t.Run("Setup OTEL Echo", func(t *testing.T) {
		if err := SetupOtel("test", "file:"+filepath.Join(base.GetGinfraTmpTestDir(), "otel", "otel2")); err != nil {
			t.Error(err)
			return
		}
		s.Use(GetEchoMiddleware("/path", "otel-service"))

		errCh := make(chan error)
		go func() {
			errCh <- s.Start("0.0.0.0:33433")
		}()

		time.Sleep(time.Millisecond * 75)
		select {
		case em := <-errCh:
			t.Error(em)
			return
		default:

			break
		}
	})

	t.Run("Trace OTEL Echo", func(t *testing.T) {

		resp, err := http.Get("http://0.0.0.0:33433/ok")
		_ = resp.Body.Close()
		testOtelEchoFlush()

		s, err := testutil.NewScanPackB(os.ReadFile(filepath.Join(base.GetGinfraTmpTestDir(), "otel", "otel2_trace.log")))
		if err != nil {
			t.Error(s)
			return
		}

		assert.NotEqual(t, -1, s.Seeks("\"Name\":\"/ok\"", "\"http.method\",\"Value\":{\"Type\":\"STRING\",\"Value\":\"GET\"}}",
			"\"Type\":\"INT64\",\"Value\":200"))
	})

	t.Run("Metrics OTEL Echo", func(t *testing.T) {

		resp, err := http.Get("http://0.0.0.0:33433/happy")
		_ = resp.Body.Close()
		testOtelEchoFlush()

		s, err := testutil.NewScanPackB(os.ReadFile(filepath.Join(base.GetGinfraTmpTestDir(), "otel", "otel2_metrics.log")))
		if err != nil {
			t.Error(s)
			return
		}

		assert.NotEqual(t, -1, s.Seeks("\"Value\":\"test\"", "\"happyhit\"",
			"{\"Key\":\"B1\",\"Value\":{\"Type\":\"STRING\",\"Value\":\"Binformation1\"}",
			"\"Value\":100}"))
	})

	t.Run("Shutdown OTEL echo", func(t *testing.T) {
		_ = s.Shutdown(context.Background())
		if err := ShutdownOtel(); err != nil {
			t.Error(err)
		}
	})
}

func TestOtelEchoCases(t *testing.T) {

	t.Run("echo status error code", func(t *testing.T) {
		c, s := getEchoStatus(501)
		assert.Equal(t, codes.Error, c)
		assert.Equal(t, "", s)
	})

	t.Run("echo status bad code", func(t *testing.T) {
		c, s := getEchoStatus(1)
		assert.Equal(t, codes.Error, c)
		assert.Contains(t, s, "Bad")
	})

	t.Run("otel not set up", func(t *testing.T) {
		_ = ShutdownOtel() // Make sure there is no lingering setup.
		assert.Panics(t, func() {
			_ = GetEchoMiddleware("x", "x")
		})
	})

}
