/*
Package shared
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Shared infrastructure for api services.
*/
package shared

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	echo5 "github.com/labstack/echo/v5"
	echomiddleware5 "github.com/labstack/echo/v5/middleware"
	apiechomiddleware "github.com/oapi-codegen/echo-middleware"
	"gitlab.com/ginfra/ginfra/base"
)

// #####################################################################################################################
// # GENERAL

type SchemaValidator struct {
	r routers.Router
}

// GetNewSchemaValidator initializes and returns a new SchemaValidator configured with the provided OpenAPI specification.
func GetNewSchemaValidator(swagger *openapi3.T) (*SchemaValidator, error) {
	var (
		sv  SchemaValidator
		err error
	)

	sv.r, err = gorillamux.NewRouter(swagger)
	if err != nil {
		// May unpanic this later.
		panic(err)
	}

	return &sv, err
}

// GetValidationSchemas creates and returns a slice of SchemaValidator instances for the given OpenAPI specs or an error.
func GetValidationSchemas(sw []*openapi3.T) ([]*SchemaValidator, error) {
	var (
		sv []*SchemaValidator
	)

	for _, i := range sw {
		if nsv, err := GetNewSchemaValidator(i); err == nil {
			sv = append(sv, nsv)
		} else {
			return nil, err
		}
	}

	return sv, nil
}

// #####################################################################################################################
// # ECHO MIDDLEWARE

// ValidatorManager is a middleware function that validates requests against provided schemas using Echo middleware.
func ValidatorManager(schemas []*SchemaValidator) echo.MiddlewareFunc {
	s := schemas // for debugging
	var skip echomiddleware.Skipper = echomiddleware.DefaultSkipper

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var err error

			// TODO HAX: Need a more generic way to register skips for validation.  For now, just ignore websocket connections.
			if c.Request().RequestURI != "/ws" && c.Request().RequestURI != "/message" {
				for _, i := range s {
					if !skip(c) {
						e := apiechomiddleware.ValidateRequestFromContext(c, i.r, nil)
						if e == nil {
							err = nil
							break
						}
						err = base.NewGinfraErrorA(e.Error(), base.LM_CODE, e.Code, base.LM_ERROR,
							fmt.Sprintf("%s", e.Message))
					}
				}

			}

			if err == nil {
				return next(c)
			}
			return err
		}
	}
}

// ValidatorManager5 is a middleware function that validates requests against provided schemas using Echo v5 middleware.
func ValidatorManager5(schemas []*SchemaValidator) echo5.MiddlewareFunc {
	s := schemas // for debugging
	var skip echomiddleware5.Skipper = echomiddleware5.DefaultSkipper

	return func(next echo5.HandlerFunc) echo5.HandlerFunc {
		return func(c *echo5.Context) error {
			var err error

			// TODO HAX: Need a more generic way to register skips for validation.  For now, just ignore websocket connections.
			if c.Request().RequestURI != "/ws" && c.Request().RequestURI != "/message" {
				for _, i := range s {
					if !skip(c) {
						e := validateRequestFromContext5(c, i.r)
						if e == nil {
							err = nil
							break
						}
						err = e
					}
				}

			}

			if err == nil {
				return next(c)
			}
			return err
		}
	}
}

// validateRequestFromContext5 performs OpenAPI request validation for an echo/v5 context,
// replicating the logic of apiechomiddleware.ValidateRequestFromContext without requiring echo/v4.
func validateRequestFromContext5(c *echo5.Context, router routers.Router) *echo.HTTPError {
	req := c.Request()
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		switch e := err.(type) {
		case interface{ Error() string }:
			return echo.NewHTTPError(http.StatusNotFound, e.Error())
		default:
			return echo.NewHTTPError(http.StatusInternalServerError,
				fmt.Sprintf("error validating route: %s", err.Error()))
		}
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	if err = openapi3filter.ValidateRequest(context.Background(), validationInput); err != nil {
		switch e := err.(type) {
		case *openapi3filter.RequestError:
			return &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  e.Error(),
				Internal: err,
			}
		case *openapi3filter.SecurityRequirementsError:
			return &echo.HTTPError{
				Code:     http.StatusForbidden,
				Message:  e.Error(),
				Internal: err,
			}
		default:
			return &echo.HTTPError{
				Code:     http.StatusInternalServerError,
				Message:  err.Error(),
				Internal: err,
			}
		}
	}

	return nil
}