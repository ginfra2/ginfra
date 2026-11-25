//go:build functional_test

/*
Package poke
Copyright (c) 2025 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Test the get.
*/
package poke

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name              string
		url               string
		auth              string
		args              []string
		expectedBody      string
		expectedError     error
		expectedErrorText string
		responseBody      string
		responseStatus    int
	}{
		{
			name:           "valid request",
			url:            "http://localhost",
			auth:           "valid-auth",
			args:           []string{"key", "value"},
			expectedBody:   `{"success": true}`,
			expectedError:  nil,
			responseBody:   `{"success": true}`,
			responseStatus: http.StatusOK,
		},
		{
			name:              "invalid URL",
			url:               "invalid-url",
			auth:              "",
			args:              nil,
			expectedBody:      "",
			expectedError:     nil,
			expectedErrorText: "invalid-url",
			responseBody:      "",
			responseStatus:    http.StatusInternalServerError,
		},
		{
			name:           "server error response",
			url:            "http://localhost",
			auth:           "valid-auth",
			args:           nil,
			expectedBody:   "",
			expectedError:  base.NewGinfraErrorA("Poke GET failed.", base.LM_CODE, http.StatusInternalServerError, base.LM_CAUSE, "Internal Server Error"),
			responseBody:   "Internal Server Error",
			responseStatus: http.StatusInternalServerError,
		},
		{
			name:           "no auth header",
			url:            "http://localhost",
			auth:           "",
			args:           []string{"key1", "value1", "key2", "value2"},
			expectedBody:   `{"success": true}`,
			expectedError:  nil,
			responseBody:   `{"success": true}`,
			responseStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for i := 0; i < len(tt.args); i += 2 {
					key := tt.args[i]
					value := tt.args[i+1]
					if r.Header.Get(key) != value {
						http.Error(w, "Header mismatch", http.StatusBadRequest)
						return
					}
				}
				if tt.auth != "" && r.Header.Get("auth") != tt.auth {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				w.WriteHeader(tt.responseStatus)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Use the mock server URL if URL is valid
			url := tt.url
			if url == "http://localhost" {
				url = server.URL
			}

			body, err := Get(app.GetRootContext(), url, tt.auth, tt.args)

			// Verify the results
			if tt.expectedErrorText != "" {
				if !strings.Contains(err.Error(), tt.expectedErrorText) {
					t.Errorf("expected error text to find: %s, got: %s", tt.expectedErrorText, err.Error())
				}
			} else if !reflect.DeepEqual(err, tt.expectedError) {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if body != tt.expectedBody {
				t.Errorf("expected body: %s, got: %s", tt.expectedBody, body)
			}

		})
	}
}

func TestGetCases(t *testing.T) {
	t.Run("bad url", func(t *testing.T) {
		_, err := Get(app.GetRootContext(), "smurf:\000///sdfgsfdg", "xxxx", []string{})
		assert.ErrorContains(t, err, "parse")
	})
}
