/*
Package log
Copyright (c) 2023, 2025 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Logging facility.  I have decided to use go.uber.org/zap (https://github.com/uber-go/zap).  I have been a fan of
structured logging for a long time and have been adding it to my projects since 2002.
*/
package log

import (
	"encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log/slog"
	"path/filepath"
)

const BaseFileName = "ginfra.log"

// GetLogger creates and returns a configured Zap logger (https://github.com/uber-go/zap) for logging, storing logs in
// a file based on the given home directory.
func GetLogger(homedir string, debugging bool) *zap.Logger {
	gl := NewGLogger(debugging)
	gl.AddFileCore(homedir)
	return gl.GetLogger()
}

type GLogger struct {
	cores   []zapcore.Core
	encoder zapcore.Encoder
	level   zapcore.LevelEnabler
}

func NewGLogger(debugging bool) *GLogger {
	var level = zap.InfoLevel
	if debugging {
		level = zap.DebugLevel
	}
	l := zap.NewAtomicLevelAt(level)

	ecfg := zap.NewDevelopmentEncoderConfig()
	ecfg.TimeKey = "timestamp"
	ecfg.EncodeTime = zapcore.ISO8601TimeEncoder
	e := zapcore.NewJSONEncoder(ecfg)

	return &GLogger{encoder: e, level: l}
}

func (g *GLogger) AddFileCore(homedir string) {
	if homedir != "" {
		wrt := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Join(homedir, BaseFileName),
			MaxSize:    75,
			MaxBackups: 10,
			MaxAge:     7,
		})
		g.cores = append(g.cores, zapcore.NewCore(g.encoder, wrt, g.level))
	}
}

func (g *GLogger) AddWriterCore(w io.Writer) {
	if w != nil {
		g.cores = append(g.cores, zapcore.NewCore(g.encoder, zapcore.Lock(zapcore.AddSync(w)), g.level))
	}
}

type GLoggerWriter struct {
	actual *slog.Logger
}

func NewGLoggerWriter(logger *slog.Logger) *GLoggerWriter {
	return &GLoggerWriter{actual: logger}
}

func (lw *GLoggerWriter) Write(p []byte) (n int, err error) {

	var data map[string]interface{}
	if err := json.Unmarshal(p, &data); err != nil {
		return 0, err
	}

	lw.actual.Info("Logged", "log", data)

	return len(p), nil
}

func (g *GLogger) AddSlogCore(logger *slog.Logger) {
	if logger != nil {
		gw := NewGLoggerWriter(logger)
		g.cores = append(g.cores, zapcore.NewCore(g.encoder, zapcore.Lock(zapcore.AddSync(gw)), g.level))
	}
}

func (g *GLogger) GetLogger() *zap.Logger {
	tee := zapcore.NewTee(g.cores...)
	z := zap.New(tee)
	return z
}
