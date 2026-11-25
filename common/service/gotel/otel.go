/*
Package gotel
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

OpenTelemetry trace and metrics exports.  Once setup they can be accessed by anything.  See the echo middleware as an
example.

Logs are on the TODO list.
*/
package gotel

import (
	"context"
	"gitlab.com/ginfra/ginfra/base"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelstdlog "go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	otelstdmetric "go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	otelstdtrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const OtelVersion = "0.51.0"

type OtelConfig struct {
	Tracer         oteltrace.Tracer
	TracerProvider *sdktrace.TracerProvider
	Propagators    propagation.TextMapPropagator

	Meter         otelmetric.Meter
	MeterProvider *sdkmetric.MeterProvider

	Logger         *slog.Logger
	LoggerProvider *sdklog.LoggerProvider
}

var (
	otelCfg *OtelConfig
)

func GetOtelConfig() *OtelConfig {
	return otelCfg
}

// ==========================================================================================================
// = TRACE EXPORTER

func setupOtelTraceExporterGrcp(loc string) (*otlptrace.Exporter, error) {

	u, err := url.Parse(loc)
	if err != nil {
		return nil, base.NewGinfraErrorChild("OpenTelemetry: bad collector url.", err)
	}

	conn, err := grpc.NewClient(u.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// This is probably not possible
		return nil, base.NewGinfraErrorChild("OpenTelemetry: failed to create gRPC connection to collector.", err)
	}

	// Set up a trace exporter
	te, err := otlptracegrpc.New(context.Background(), otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		// This is also probably not possible.  Only failing options can cause it and we don't have any.
		return nil, base.NewGinfraErrorChild("OpenTelemetry: failed to create trace exporter.", err)
	}
	return te, nil
}

func setupOtelTraceExporterFile(path string) (*otelstdtrace.Exporter, error) {

	f, err := os.OpenFile(path+OtelTraceFileSuffix, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return nil, base.NewGinfraErrorChildA("Could not open otel trace export file", err, base.LM_FILEPATH,
			filepath.Join(path, OtelTraceFileSuffix))
	}

	var e *otelstdtrace.Exporter
	e, err = otelstdtrace.New(otelstdtrace.WithWriter(f))

	return e, nil
}

func setupOtelTraceExporter(collector string) (sdktrace.SpanExporter, error) {
	if strings.HasPrefix(collector, "grpc:") {
		return setupOtelTraceExporterGrcp(collector)
	} else if strings.HasPrefix(collector, "file:") {
		return setupOtelTraceExporterFile(collector[5:])
	}
	return nil, base.NewGinfraErrorA("Unknown collector.", base.LM_URL, collector)
}

// ==========================================================================================================
// = METRICS EXPORTER

func setupOtelMetricsExporterGrcp(loc string) (*otlpmetricgrpc.Exporter, error) {

	u, err := url.Parse(loc)
	if err != nil {
		return nil, base.NewGinfraErrorChild("OpenTelemetry: bad collector url.", err)
	}

	conn, err := grpc.NewClient(u.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, base.NewGinfraErrorChild("OpenTelemetry: failed to create gRPC connection to collector.", err)
	}

	// Set up a metrics exporter
	te, err := otlpmetricgrpc.New(context.Background(), otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, base.NewGinfraErrorChild("OpenTelemetry: failed to create trace exporter.", err)
	}
	return te, nil
}

func setupOtelMetricsExporterFile(path string) (metric.Exporter, error) {

	f, err := os.OpenFile(path+OtelMetricsFileSuffix, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return nil, base.NewGinfraErrorChildA("Could not open otel metrics export file", err, base.LM_FILEPATH,
			filepath.Join(path, OtelMetricsFileSuffix))
	}

	return otelstdmetric.New(otelstdmetric.WithWriter(f))
}

func setupOtelMetricsExporter(collector string) (metric.Exporter, error) {
	if strings.HasPrefix(collector, "grpc:") {
		return setupOtelMetricsExporterGrcp(collector)
	} else if strings.HasPrefix(collector, "file:") {
		return setupOtelMetricsExporterFile(collector[5:])
	}
	return nil, base.NewGinfraErrorA("Unknown collector.", base.LM_URL, collector)
}

// ==========================================================================================================
// = LOG EXPORTER

func setupOtelLogsExporterGrcp(loc string) (*otlploggrpc.Exporter, error) {

	u, err := url.Parse(loc)
	if err != nil {
		return nil, base.NewGinfraErrorChild("OpenTelemetry: bad collector url.", err)
	}

	conn, err := grpc.NewClient(u.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, base.NewGinfraErrorChild("OpenTelemetry: failed to create gRPC connection to collector.", err)
	}

	// Set up a metrics exporter
	te, err := otlploggrpc.New(context.Background(), otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, base.NewGinfraErrorChild("OpenTelemetry: failed to create log exporter.", err)
	}
	return te, nil
}

func setupOtelLogsExporterFile(path string) (sdklog.Exporter, error) {

	f, err := os.OpenFile(path+OtelLogsFileSuffix, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return nil, base.NewGinfraErrorChildA("Could not open otel log export file", err, base.LM_FILEPATH,
			filepath.Join(path, OtelLogsFileSuffix))
	}

	return otelstdlog.New(otelstdlog.WithWriter(f))
}

func setupOtelLogsExporter(collector string) (sdklog.Exporter, error) {
	if strings.HasPrefix(collector, "grpc:") {
		return setupOtelLogsExporterGrcp(collector)
	} else if strings.HasPrefix(collector, "file:") {
		return setupOtelLogsExporterFile(collector[5:])
	}
	return nil, base.NewGinfraErrorA("Unknown collector.", base.LM_URL, collector)
}

// ==========================================================================================================
// = SETUP

// SetupOtel initializes OpenTelemetry tracing and metrics for a given service name and collector endpoint.
// It configures the tracer provider, meter provider, and propagators for distributed traces and metrics collection.
// The function returns an error if there is any issue during setup.
func SetupOtel(name, collector string) error {

	var o OtelConfig

	// -- TRACE ---------------------------------------------------------------------------------------
	te, err := setupOtelTraceExporter(collector)
	if err != nil {
		return base.NewGinfraErrorChild("Could not get otel trace exporter", err)
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(name),
		),
	)
	if err != nil {
		return base.NewGinfraErrorChild("Could not get otel resource", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(te)
	o.TracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(r),
	)

	otel.SetTracerProvider(o.TracerProvider)

	o.Propagators = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(o.Propagators)

	o.Tracer = otel.Tracer(name)

	// -- METRIC ---------------------------------------------------------------------------------------
	me, err := setupOtelMetricsExporter(collector)
	if err != nil {
		return base.NewGinfraErrorChild("Could not get otel metrics exporter", err)
	}

	o.MeterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(me)),
		sdkmetric.WithResource(r),
	)
	otel.SetMeterProvider(o.MeterProvider)

	o.Meter = otel.Meter(name)

	// Standard metrics
	go OtelSystemResourceMetrics(o.Meter)

	// -- METRIC ---------------------------------------------------------------------------------------
	le, err := setupOtelLogsExporter(collector)
	if err != nil {
		return base.NewGinfraErrorChild("Could not get otel logs exporter", err)
	}
	o.LoggerProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(le)),
	)
	global.SetLoggerProvider(o.LoggerProvider)

	// TODO proxying through slog is a bit goofy, so eventually get rid of it and use a zap backed Handler.
	o.Logger = otelslog.NewLogger(name)

	otelCfg = &o
	return nil
}

// ShutdownOtel shuts down the OpenTelemetry tracer and meter providers, releasing all associated resources.
func ShutdownOtel() error {
	var err error
	if otelCfg != nil {
		err = otelCfg.TracerProvider.Shutdown(context.Background())
		e := otelCfg.MeterProvider.Shutdown(context.Background())
		if e != nil {
			// Only bothering to keep one of the errors.
			err = e
		}
		e = otelCfg.LoggerProvider.Shutdown(context.Background())
		if e != nil {
			// Only bothering to keep one of the errors.
			err = e
		}
		otelCfg = nil
	}
	return err
}
