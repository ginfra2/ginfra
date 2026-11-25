/*
Package service
// >A#

You may license this file however you wish.

// >A#
 */
package service

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"<<<Module>>>/api/service/server"
	"<<<Module>>>/local"
	"<<<Module>>>/providers"

	// # YOUR IMPORTS START HERE
	// >B###############################################################################################################

	// >B###############################################################################################################
	// #  YOUR IMPORTS END
)

type Service<<<ServiceName>>> struct {
	Providers      *providers.ProvidersLoaded
	ServiceContext *local.GContext

	// # ADD ADDITIONAL SERVICE DATA START HERE
	// >C###############################################################################################################

	// >C###############################################################################################################
	// # ADD ADDITIONAL SERVICE DATA END HERE
}

func NewService<<<ServiceName>>>(gcontext *local.GContext, providers *providers.ProvidersLoaded) (*Service<<<ServiceName>>>, error) {
	return &Service<<<ServiceName>>>{
		Providers:      providers,
		ServiceContext: gcontext,
	}, nil
}

func ServiceSetup(service *Service<<<ServiceName>>>) error {

	var err error

	// # ADD ADDITIONAL SETUP CODE START HERE.  All configuration is complete and the providers are ready.
	// >D###############################################################################################################

	// >D###############################################################################################################
	// # ADD ADDITIONAL SETUP CODE END HERE

	return err
}

var specifics string

func getServiceSpecifics(gis *Service<<<ServiceName>>>) string {
	if specifics == "" {
		specifics = shared.LoadSpecifics(gis.ServiceContext.Scfg.ServiceHome)
	}
	return specifics
}

func RegisterHandlers(echoServer *echo.Echo, service *Service<<<ServiceName>>>) ([]*openapi3.T, error) {

	// Always register main service.
	sw, err := server.GetSwagger()
	s := []*openapi3.T{sw}
	s = append(s, sw)
	server.RegisterHandlers(echoServer, service)

	// NOTE: if you merge all services into one, the above registration is all you need.
	//
	// If you have multiple api services you host will need to be registered here.
	// If you have just a single api server, this should be enough:
	//   service.RegisterHandlers(echoServer, service)
	//
	// If you have more than one, you'll need to import them with unique names like this:
	//    v1 "gitlab.com/ginfra/gservices/service_otherservice/api/service_otherservice/v1/server"
	//    v2 "gitlab.com/ginfra/gservices/service_otherservice/api/service_otherservice/v2/server"
	//
	// and register them something like this:
	//   v1.RegisterHandlers(echoServer, service)
	//   v2.RegisterHandlers(echoServer, service)
	//
	// If you want schema validation, you must add the swagger definition to the array returned by this function.
	// It could look like this for a single api:
	//
	//      sw, err := myapiservice.GetSwagger()
	//      s = append(s, sw)
	//
	// Or this for multiple apis.
	//
	// 		sw, err := v1.GetSwagger()
	//		s = append(s, sw)
	//		sw, err = v2.GetSwagger()
	//		s = append(s, sw)
	//
	// #  REGISTRATIONS GO HERE
	// >E###############################################################################################################

	// >E###############################################################################################################
	// # END REGISTRATIONS

	return s, err
}

// ^^###################################################################################################################
// # ADDITIONAL CODE GOES BELOW HERE
// # Stubs can go here or in other files.

