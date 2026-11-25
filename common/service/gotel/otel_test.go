//go:build unit_test

/*
Package gotel
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

OpenTelemetry.
*/
package gotel

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	testutil "gitlab.com/ginfra/ginfra/common/test"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	common.SetupTest()
	_ = os.MkdirAll(filepath.Join(base.GetGinfraTmpTestDir(), "otel"), 0777)
}

func testOtelFlush() {
	cfg := GetOtelConfig()
	_ = cfg.MeterProvider.ForceFlush(context.Background())
	_ = cfg.TracerProvider.ForceFlush(context.Background())
	_ = cfg.LoggerProvider.ForceFlush(context.Background())
}

func TestOtel(t *testing.T) {

	attr := []attribute.KeyValue{
		attribute.String("A1", "information1"),
		attribute.String("A2", "information2"),
	}

	t.Run("Setup OTEL", func(t *testing.T) {
		if err := SetupOtel("test", "file:"+filepath.Join(base.GetGinfraTmpTestDir(), "otel", "otel")); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("Trace OTEL", func(t *testing.T) {
		cfg := GetOtelConfig()

		ctx, span := cfg.Tracer.Start(context.Background(), "TestSpan", trace.WithAttributes(attr...))

		for i := 0; i < 100; i++ {
			_, ns := cfg.Tracer.Start(ctx, fmt.Sprintf("Trace #%d", i))
			ns.End()
		}

		span.End()
		testOtelFlush()

		s, err := testutil.NewScanPackB(os.ReadFile(filepath.Join(base.GetGinfraTmpTestDir(), "otel", "otel_trace.log")))
		if err != nil {
			t.Error(s)
			return
		}

		assert.NotEqual(t, -1, s.Seeks("\"Trace #0", "\"Trace #99", "\"Name\":\"TestSpan\"",
			"\"Key\":\"A2\",\"Value\":{\"Type\":\"STRING\",\"Value\":\"information2\""))
	})

	t.Run("Metrics OTEL", func(t *testing.T) {
		cfg := GetOtelConfig()

		c, err := cfg.Meter.Int64Counter("hit", metric.WithDescription("Hit counts"))
		if err != nil {
			t.Error(err)
			return
		}

		for i := 0; i < 100; i++ {
			c.Add(context.Background(), 1, metric.WithAttributes(attr...))
		}

		testOtelFlush()

		s, err := testutil.NewScanPackB(os.ReadFile(filepath.Join(base.GetGinfraTmpTestDir(), "otel", "otel_metrics.log")))
		if err != nil {
			t.Error(s)
			return
		}

		assert.NotEqual(t, -1, s.Seeks("\"Name\":\"hit\",\"Description\":\"Hit counts\"",
			"\"Type\":\"STRING\",\"Value\":\"information1\"", "\"Value\":100"))

	})

	t.Run("Logger OTEL", func(t *testing.T) {
		cfg := GetOtelConfig()

		cfg.Logger.ErrorContext(context.Background(), "Log entry 1", "A1", "information1", "A2", "information2")
		cfg.Logger.InfoContext(context.Background(), "Log entry 2", "B1", "information1", "B2", "information2")
		cfg.Logger.InfoContext(context.Background(), "Log entry 3", "C1", "information1", "C2", "information2")

		testOtelFlush()

		s, err := testutil.NewScanPackB(os.ReadFile(filepath.Join(base.GetGinfraTmpTestDir(), "otel", "otel_logs.log")))
		if err != nil {
			t.Error(s)
			return
		}

		assert.NotEqual(t, -1, s.Seeks("Log entry 1", "Log entry 2", "Log entry 3", "C2"))

	})

	t.Run("Shutdown OTEL", func(t *testing.T) {
		if err := ShutdownOtel(); err != nil {
			t.Error(err)
		}
	})
}

func TestOtelCases(t *testing.T) {

	t.Run("bad path", func(t *testing.T) {
		_, err := setupOtelTraceExporterFile("/dev/null/../../../../../../../../../../../../../../../../../../../../../../../../../../../../../../../")
		assert.ErrorContains(t, err, "export file")
	})
	t.Run("bad path", func(t *testing.T) {
		_, err := setupOtelMetricsExporterFile("/dev/null/../../../../../../../../../../../../../../../../../../../../../../../../../../../../../../../")
		assert.ErrorContains(t, err, "export file")
	})
	t.Run("unknown URL", func(t *testing.T) {
		_, err := setupOtelTraceExporter("bork:/asdfasdf")
		assert.ErrorContains(t, err, "Unknown")
	})
	t.Run("unknown URL", func(t *testing.T) {
		_, err := setupOtelMetricsExporter("bork:/asdfasdf")
		assert.ErrorContains(t, err, "Unknown")
	})
	t.Run("unknown URL", func(t *testing.T) {
		err := SetupOtel("xxx", "bork:/asdfasdf")
		assert.ErrorContains(t, err, "Could not get")
	})

}

func TestOtelGrpc(t *testing.T) {
	t.Run("ok URL trace", func(t *testing.T) {
		_, err := setupOtelTraceExporter("grpc://bork")
		assert.NoError(t, err)
	})
	t.Run("bad URL trace", func(t *testing.T) {
		_, err := setupOtelTraceExporterGrcp("asdfasdf\nasdfasdf")
		assert.ErrorContains(t, err, "bad collector url")
	})
	t.Run("ok URL metrics", func(t *testing.T) {
		_, err := setupOtelMetricsExporter("grpc://bork")
		assert.NoError(t, err)
	})
	t.Run("ok bad metrics", func(t *testing.T) {
		_, err := setupOtelTraceExporterGrcp("asdfasdf\nasdfasdf")
		assert.ErrorContains(t, err, "bad collector url")
	})
	t.Run("ok URL log", func(t *testing.T) {
		_, err := setupOtelLogsExporter("grpc://bork")
		assert.NoError(t, err)
	})
	t.Run("ok bad log", func(t *testing.T) {
		_, err := setupOtelLogsExporterGrcp("asdfasdf\nasdfasdf")
		assert.ErrorContains(t, err, "bad collector url")
	})
}
