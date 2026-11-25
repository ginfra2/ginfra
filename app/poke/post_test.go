//go:build functional_test

/*
Package poke
Copyright (c) 2025 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Test the post.
*/
package poke

import (
	"gitlab.com/ginfra/ginfra/app"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPost(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		auth           string
		url            string
		mockResponse   string
		mockStatusCode int
		expectError    bool
		expectedBody   string
	}{
		{
			name:           "valid request",
			args:           []string{"param1", "value1", "param2", "value2"},
			auth:           "testauth",
			url:            "http://example.com",
			mockResponse:   `{"result": "success"}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedBody:   `{"result": "success"}`,
		},
		{
			name:           "empty args",
			args:           []string{},
			auth:           "testauth",
			url:            "http://example.com",
			mockResponse:   `{"result": "success"}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedBody:   `{"result": "success"}`,
		},
		{
			name:           "marshal error",
			args:           []string{"param1"},
			auth:           "testauth",
			url:            "http://example.com",
			mockResponse:   "",
			mockStatusCode: http.StatusOK,
			expectError:    true,
			expectedBody:   "",
		},
		{
			name:           "server error response",
			args:           []string{"param1", "value1"},
			auth:           "testauth",
			url:            "http://example.com",
			mockResponse:   `{"error": "something went wrong"}`,
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedBody:   "",
		},
		{
			name:           "invalid URL",
			args:           []string{"param1", "value1"},
			auth:           "testauth",
			url:            ":invalid-url",
			mockResponse:   "",
			mockStatusCode: http.StatusOK,
			expectError:    true,
			expectedBody:   "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.mockStatusCode)
				_, _ = w.Write([]byte(test.mockResponse))
			}))
			defer server.Close()

			url := test.url
			if test.name != "invalid URL" {
				url = server.URL
			}

			gctx := &app.GContext{}
			body, err := Post(gctx, url, test.auth, test.args)

			if test.expectError {
				if err == nil {
					t.Fatalf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("did not expect error but got: %v", err)
				}

				if body != test.expectedBody {
					t.Errorf("expected body %q, got %q", test.expectedBody, body)
				}
			}
		})
	}
}
