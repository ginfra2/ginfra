package shared

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	config "gitlab.com/ginfra/ginfra/api/service_giconfig/client"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidatorManager(t *testing.T) {
	e := echo.New()
	rec := httptest.NewRecorder()

	csw, err := config.GetSwagger()
	assert.NoError(t, err)
	gsw, err := config.GetSwagger()
	assert.NoError(t, err)

	testCases := []struct {
		name          string
		schemas       []*openapi3.T
		handlerFunc   echo.HandlerFunc
		req           *http.Request
		expectedError string
	}{

		{
			name:    "multiple schemas valid request",
			schemas: []*openapi3.T{csw, gsw},
			req:     httptest.NewRequest(http.MethodGet, "/ginfra/specifics", nil),
			handlerFunc: func(context echo.Context) error {
				return context.String(http.StatusOK, "test response")
			},
			expectedError: "",
		},
		{
			name:    "multiple schemas invalid request",
			schemas: []*openapi3.T{csw, gsw},
			req:     httptest.NewRequest(http.MethodGet, "/configservice/query", nil), // Missing parames
			handlerFunc: func(context echo.Context) error {
				return context.String(http.StatusOK, "test response")
			},
			expectedError: "value is required but missing",
		},
		{
			name:    "empty schemas",
			schemas: []*openapi3.T{},
			req:     httptest.NewRequest(http.MethodGet, "/ginfra/specifics", nil),
			handlerFunc: func(context echo.Context) error {
				return context.String(http.StatusOK, "test response")
			},
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			context := e.NewContext(tc.req, rec)
			next := tc.handlerFunc

			vs, err := GetValidationSchemas(tc.schemas)
			assert.NoError(t, err)
			middleware := ValidatorManager(vs)
			err = middleware(next)(context)

			if tc.expectedError == "" {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
				assert.Less(t, 1, strings.Index(err.Error(), tc.expectedError))
			}
		})
	}
}
