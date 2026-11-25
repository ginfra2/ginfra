/*
Package app
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Application context.  Most of the values are set by the cli manager.
*/
package app

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"sync"
)

// #####################################################################################################################
// # GINFRA CONTEXT
// # This is meant for the app and not services.

var (
	// rootContext is the root, initial context created for the entire application during startup.
	rootContext  *GContext
	savedContext GContext
	rootMutex    sync.Mutex
)

func init() {
	rootContext, _ = NewContext(nil)
}

type GContext struct {

	// ---------------------------------------------------------------------------------
	// - CONFIGURATION - these are loaded during application start, typically by the command line processor.

	Debugging bool
	Blob      bool
	Target    string
	Source    string

	InterfaceDir string
	NoControl    bool

	// Deployment related
	Bundle string

	// Service related
	ServiceHomePath string
	ServicePath     string // Optional service path used for things like docker name.  It defaults to the service name.
	ServiceClass    string
	Deployment      string
	Service         string
	Host            string
	Port            int
	Specifics       string // json
	Clear           bool
	DontStart       bool // Don't actually start the service.  Useful for IDE testing or container, so the service is registered, the container is started, associated servers are started and etc.
	NoRollback      bool // Do not rollback anything in case of error.
	CommandLine     string
	Debugger        int //  Tell the target (container) to attach a debugger.  This value will be the listen port.  0 means don't do it.
	Authtoken       string

	// Container related
	InitContainer string
	Gpus          string // GPU specification

	// Platform
	PortalHostnameUrl string // This could be host with docker engine or any entry point for managed platforms.

	// ---------------------------------------------------------------------------------
	// - RUNTIME

	logger *zap.SugaredLogger
	alog   zap.AtomicLevel

	// ---------------------------------------------------------------------------------
	// - FOR TESTING

	MostRecentOutput string
	HostDirPath      string
}

// NewContext creates a new GContext with the specified alternate io.Writer for logging.
// If alt is nil, the logging output is directed only to os.Stdout.
// The function returns the created GContext and an error if any.
func NewContext(alt io.Writer) (*GContext, error) {
	gctx := &GContext{
		Debugging: false,
	}

	gctx.alog = zap.NewAtomicLevel()
	gctx.alog.SetLevel(zapcore.InfoLevel)

	//ecfg := zap.NewProductionEncoderConfig()
	ecfg := zap.NewDevelopmentEncoderConfig()

	var l *zap.Logger
	if alt == nil {
		l = zap.New(zapcore.NewCore(
			zapcore.NewJSONEncoder(ecfg),
			zapcore.Lock(os.Stdout),
			gctx.alog,
		))
	} else {
		core := zapcore.NewTee(
			zapcore.NewCore(zapcore.NewJSONEncoder(ecfg), zapcore.Lock(os.Stdout), gctx.alog),
			zapcore.NewCore(zapcore.NewJSONEncoder(ecfg), zapcore.Lock(zapcore.AddSync(alt)), gctx.alog),
		)
		l = zap.New(core)
	}

	defer func(l *zap.Logger) {
		_ = l.Sync()
	}(l)
	gctx.logger = l.Sugar()

	return gctx, nil
}

// GetRootContext returns the singleton root GContext instance, providing access to shared configuration and state.   The root
// will be the first context created at startup.
func GetRootContext() *GContext {
	return rootContext
}

// GetCopyRootContext returns a copy of the root GContext instance.  It is a shallow copy so all resources are the same.
func GetCopyRootContext() GContext {
	return *rootContext
}

// TurnDebuggingOn sets the debugging flag GContext and adjusts the log level to Debug.  It is up to the individual
// components to support the flag or not.
func (gctx *GContext) TurnDebuggingOn() {
	gctx.Debugging = true
	gctx.alog.SetLevel(zapcore.DebugLevel)
}

// GetLogger gets the context's logger.
func (gctx *GContext) GetLogger() *zap.SugaredLogger {
	if !gctx.Debugging {
		gctx.TurnDebuggingOn()
	}
	return gctx.logger
}

// UseRootContext is for testing only and is very dangerous to play with otherwise.
func UseRootContext() *GContext {
	rootMutex.Lock()
	savedContext = *rootContext
	return rootContext
}

func ReleaseRootContext() {
	rootContext = &savedContext
	rootMutex.Unlock()
}
