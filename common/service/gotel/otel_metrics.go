/*
Package gotel
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

OpenTelemetry specific to metrics.
*/
package gotel

import (
	"context"
	"go.opentelemetry.io/otel/metric"
	"runtime"
	"time"
)

const OtelHttpServerLatency = "http.server.latency"
const OtelMetricProcessMemory = "process.memory"
const OtelMetricProcessGoroutines = "process.goroutines"

const OtelMetricMeasureInterval = 10

func OtelSystemResourceMetrics(meter metric.Meter) {

	it := time.NewTicker(OtelMetricMeasureInterval * time.Second)
	for {
		select {
		case <-it.C:
			// Memory usage.
			_, _ = meter.Float64ObservableGauge(OtelMetricProcessMemory,
				metric.WithFloat64Callback(
					func(ctx context.Context, fo metric.Float64Observer) error {
						var ms runtime.MemStats
						runtime.ReadMemStats(&ms)
						allocatedMemoryInMB := float64(ms.Alloc) / float64(1_048_576)
						fo.Observe(allocatedMemoryInMB)
						return nil
					}))
			// Go routines in flight.
			_, _ = meter.Int64ObservableGauge(OtelMetricProcessMemory,
				metric.WithInt64Callback(
					func(ctx context.Context, fo metric.Int64Observer) error {
						fo.Observe(int64(runtime.NumGoroutine()))
						return nil
					}))
		}
	}
}
