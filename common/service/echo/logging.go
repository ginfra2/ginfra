/*
Package echo
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Labstack echo logging middleware.
*/

package echo

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Echo2ZapLogger for logging to zap which most of ginfra does.
func Echo2ZapLogger(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var (
				st  = time.Now()
				req = c.Request()
				res = c.Response()
			)

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			fields := []zapcore.Field{
				zap.String("time", time.Since(st).String()),
				zap.String("target.host", req.Host),
				zap.String("source.ip", c.RealIP()),
				zap.String("request", fmt.Sprintf("%s %s", req.Method, req.URL)),
				zap.Int("status", res.Status),
				zap.Int64("size", res.Size),
			}

			switch {
			case res.Status >= 500:
				log.With(zap.Error(err)).Error("Server error", fields...)
			case res.Status >= 400:
				log.With(zap.Error(err)).Warn("Bad request", fields...)
			case res.Status >= 300:
				log.Info("Redirect", fields...)
			default:
				log.Info("Success", fields...)
			}

			return nil
		}
	}
}
