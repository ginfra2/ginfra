/*
Package service
// >A#

You may license this file however you wish.

// >A#
*/
package service

import (
	"context"
	"github.com/labstack/echo/v4"
	"gitlab.com/ginfra/ginfra/common/service/gotel"
	"<<<Module>>>/api/service/models"
	"<<<Module>>>/local"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"net/http"

	// # YOUR IMPORTS START HERE
	// >B###############################################################################################################

	// >B###############################################################################################################
	// #  YOUR IMPORTS END
)

func (gis *Service<<<ServiceName>>>) GiControlDatasourceInit(ctx echo.Context, params models.GiControlDatasourceInitParams) error {
	var span trace.Span

	o := gotel.GetOtelConfig()
	if o != nil {
		_, span = o.Tracer.Start(context.Background(), "GiControlDatasourceInit")
		defer span.End()
	}

	// # DATA SOURCE INIT IMPLEMENTATION HERE
	// >C###############################################################################################################

	return nil

	// >C###############################################################################################################
	// # END DATA SOURCE INIT IMPLEMENTATION
}

func (gis *Service<<<ServiceName>>>) GiControlManageReset(ctx echo.Context, params models.GiControlManageResetParams) error {
	var (
		span trace.Span
		err  error
	)

	o := gotel.GetOtelConfig()
	if o != nil {
		_, span = o.Tracer.Start(context.Background(), "GiControlManageReset")
		defer span.End()
	}

	// # ADD ADDITIONAL PRE-RESET ACTIONS HERE
	// >D###############################################################################################################

	// >D###############################################################################################################
	// # END ADDITIONAL PRE-RESET ACTIONS

	// Reset not required and is up to the implementation.

	// # ADD ADDITIONAL POST-RESET ACTIONS HERE
	// >E###############################################################################################################

	NoteNoReset := "No reset necessary."
	r := &models.ControlActionResponse{
		Note: &NoteNoReset,
	}
	err = ctx.JSON(http.StatusOK, r)

	// >E###############################################################################################################
	// # END ADDITIONAL POST-RESET ACTIONS

	return err
}

func (gis *Service<<<ServiceName>>>) GiControlManageStop(ctx echo.Context, params models.GiControlManageStopParams) error {
	var (
		span trace.Span
		err  error
	)

	o := gotel.GetOtelConfig()
	if o != nil {
		_, span = o.Tracer.Start(context.Background(), "GiControlManageStop")
		defer span.End()
	}

	// # ADD ADDITIONAL PRE-STOP ACTIONS HERE
	// >F##############################################################################################################

	// >F###############################################################################################################
	// # END ADDITIONAL PRE-STOP ACTIONS

	gctx, err := local.GetContext()
	if err != nil {
		panic(err)
	}

	local.Logger.Info("auth", zap.String("param", *params.Auth), zap.String("auth", gctx.Scfg.Auth))
	if *params.Auth == gctx.Scfg.Auth {

		local.Logger.Info("Shutdown ordered.")
		_ = StopService() // The error is sort of meaningless since the water is already under the bridge here.

		NoteShutdownUnderway := "Shutdown underway"
		r := &models.ControlActionResponse{
			Note: &NoteShutdownUnderway,
		}
		err = ctx.JSON(http.StatusOK, r)

	} else {
		err = local.PostErrorUnauthorizedResponse(ctx, "stop", nil)
	}

	// # ADD ADDITIONAL POST-STOP ACTIONS HERE
	// >G###############################################################################################################

	// >G###############################################################################################################
	// # END ADDITIONAL POST-STOP ACTIONS

	return err
}

// GiGetSpecifics Get specifics for this service
// (GET /specifics)
func (gis *Service<<<ServiceName>>>) GiGetSpecifics(ctx echo.Context) error {
	var span trace.Span
	o := gotel.GetOtelConfig()
	if o != nil {
		_, span = o.Tracer.Start(context.Background(), "GiGetSpecifics")
		defer span.End()
	}

	var r = getServiceSpecifics(gis)
	return ctx.JSON(http.StatusOK, r)
}
