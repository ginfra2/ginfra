/*
Package gotel
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

OpenTelemetry middleware for Echo from labstack.  This will crete an event with standard values for an echo service call.
*/
package gotel

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	echo5 "github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func getHostPort(v string) (host, port string) {
	s := strings.Split(v, ":")
	if len(s) > 0 {
		host = s[0]
		if len(s) > 1 {
			port = s[1]
		}
	}
	return
}

func getEchoRequestAttrib(r *http.Request) []attribute.KeyValue {
	//var a []attribute.KeyValue

	h, p := getHostPort(r.RemoteAddr)
	lh, lp := getHostPort(r.Host)

	a := []attribute.KeyValue{
		{Key: OtelHttpMethod, Value: attribute.StringValue(r.Method)},
		{Key: OtelUrl, Value: attribute.StringValue(r.RequestURI)},
		{Key: OtelHttpUserAgent, Value: attribute.StringValue(r.Header.Get("User-Agent"))},
		{Key: OtelHostPeer, Value: attribute.StringValue(h)},
		{Key: OtelPortPeer, Value: attribute.StringValue(p)},
		{Key: OtelHostLocal, Value: attribute.StringValue(lh)},
		{Key: OtelPortLocal, Value: attribute.StringValue(lp)},
	}

	return a
}

func getEchoStatus(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Bad code: %d", code)
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}

// GetEchoMiddleware provides a middleware function for Echo to integrate OpenTelemetry tracing and metrics.
// It initializes a tracer and configures metrics for HTTP request latency and hit counts based on the Otel configuration.
// The middleware extracts context, starts a span, and records metrics for incoming HTTP requests.
// Parameters: `scope` specifies the tracer's scope, and `key` is used as a context key for storing the tracer.
func GetEchoMiddleware(scope string, key string) echo.MiddlewareFunc {

	oc := GetOtelConfig()
	if oc == nil {
		panic("Otel not setup before trying to acquire middleware.")
	}

	ldh, _ := oc.Meter.Int64Histogram(OtelHttpServerLatency, metric.WithUnit("ms"))

	t := oc.TracerProvider.Tracer(
		scope,
		oteltrace.WithInstrumentationVersion(OtelVersion),
	)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// If we need a different skipper, put here.
			//if Skipper(c) {
			//	return next(c)
			//}

			st := time.Now()

			// - TRACE --------------------------------------------------------------

			c.Set(key, t)
			request := c.Request()
			savedCtx := request.Context()
			defer func() {
				request = request.WithContext(savedCtx)
				c.SetRequest(request)
			}()
			ctx := oc.Propagators.Extract(savedCtx, propagation.HeaderCarrier(request.Header))
			opts := []oteltrace.SpanStartOption{
				oteltrace.WithAttributes(getEchoRequestAttrib(request)...),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}
			if path := c.Path(); path != "" {
				rAttr := semconv.HTTPRoute(path)
				opts = append(opts, oteltrace.WithAttributes(rAttr))
			}
			spanName := c.Path()
			if spanName == "" {
				spanName = fmt.Sprintf("HTTP %s route not found", request.Method)
			}

			ctx, span := t.Start(ctx, spanName, opts...)
			defer func() {
				span.End()
			}()

			c.SetRequest(request.WithContext(ctx))

			err := next(c)
			if err != nil {
				span.SetAttributes(attribute.String("echo.error", err.Error()))
				c.Error(err)
			}

			status := c.Response().Status
			span.SetStatus(getEchoStatus(status))
			if status > 0 {
				span.SetAttributes(semconv.HTTPStatusCode(status))
			}

			// - METRICS --------------------------------------------------------------
			metricAttributes := attribute.NewSet(
				attribute.String(OtelUrl, request.URL.String()),
				attribute.String(OtelHttpMethod, request.Method),
				attribute.Int(OtelHttpStatus, status),
			)
			ldh.Record(
				request.Context(),
				time.Since(st).Milliseconds(),
				metric.WithAttributeSet(metricAttributes),
			)

			// Hamfisted for now.  I just want to see something show up.
			attr := []attribute.KeyValue{
				attribute.Int("status", c.Response().Status),
			}
			ct, _ := oc.Meter.Int64Counter(c.Path(), metric.WithDescription("Hit counts"))
			ct.Add(ctx, 1, metric.WithAttributes(attr...))

			return err
		}
	}
}

// GetEchoMiddleware provides a middleware function for Echo to integrate OpenTelemetry tracing and metrics.
// It initializes a tracer and configures metrics for HTTP request latency and hit counts based on the Otel configuration.
// The middleware extracts context, starts a span, and records metrics for incoming HTTP requests.
// Parameters: `scope` specifies the tracer's scope, and `key` is used as a context key for storing the tracer.
func GetEchoMiddleware5(scope string, key string) echo5.MiddlewareFunc {

	oc := GetOtelConfig()
	if oc == nil {
		panic("Otel not setup before trying to acquire middleware.")
	}

	ldh, _ := oc.Meter.Int64Histogram(OtelHttpServerLatency, metric.WithUnit("ms"))

	t := oc.TracerProvider.Tracer(
		scope,
		oteltrace.WithInstrumentationVersion(OtelVersion),
	)
	return func(next echo5.HandlerFunc) echo5.HandlerFunc {
		return func(c *echo5.Context) error {

			// If we need a different skipper, put here.
			//if Skipper(c) {
			//	return next(c)
			//}

			st := time.Now()

			// - TRACE --------------------------------------------------------------

			c.Set(key, t)
			request := c.Request()
			savedCtx := request.Context()
			defer func() {
				request = request.WithContext(savedCtx)
				c.SetRequest(request)
			}()
			ctx := oc.Propagators.Extract(savedCtx, propagation.HeaderCarrier(request.Header))
			opts := []oteltrace.SpanStartOption{
				oteltrace.WithAttributes(getEchoRequestAttrib(request)...),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			}
			if path := c.Path(); path != "" {
				rAttr := semconv.HTTPRoute(path)
				opts = append(opts, oteltrace.WithAttributes(rAttr))
			}
			spanName := c.Path()
			if spanName == "" {
				spanName = fmt.Sprintf("HTTP %s route not found", request.Method)
			}

			ctx, span := t.Start(ctx, spanName, opts...)
			defer func() {
				span.End()
			}()

			c.SetRequest(request.WithContext(ctx))

			err := next(c)
			if err != nil {
				span.SetAttributes(attribute.String("echo.error", err.Error()))
			}

			var status int
			if resp, rerr := echo5.UnwrapResponse(c.Response()); rerr == nil {
				status = resp.Status
			}
			span.SetStatus(getEchoStatus(status))
			if status > 0 {
				span.SetAttributes(semconv.HTTPStatusCode(status))
			}

			// - METRICS --------------------------------------------------------------
			metricAttributes := attribute.NewSet(
				attribute.String(OtelUrl, request.URL.String()),
				attribute.String(OtelHttpMethod, request.Method),
				attribute.Int(OtelHttpStatus, status),
			)
			ldh.Record(
				request.Context(),
				time.Since(st).Milliseconds(),
				metric.WithAttributeSet(metricAttributes),
			)

			// Hamfisted for now.  I just want to see something show up.
			attr := []attribute.KeyValue{
				attribute.Int("status", status),
			}
			ct, _ := oc.Meter.Int64Counter(c.Path(), metric.WithDescription("Hit counts"))
			ct.Add(ctx, 1, metric.WithAttributes(attr...))

			return err
		}
	}
}

