package gowebapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireJSONContentType(t *testing.T) {
	next := RequireJSONContentType(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name       string
		method     string
		content    string
		wantStatus int
	}{
		{
			name:       "allows json content type",
			method:     http.MethodPost,
			content:    "application/json; charset=utf-8",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "allows missing content type",
			method:     http.MethodPost,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "rejects non-json content type",
			method:     http.MethodPost,
			content:    "text/plain",
			wantStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:       "ignores get requests",
			method:     http.MethodGet,
			content:    "text/plain",
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", strings.NewReader("{}"))
			if tt.content != "" {
				req.Header.Set("Content-Type", tt.content)
			}
			rr := httptest.NewRecorder()

			next.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantStatus == http.StatusUnsupportedMediaType {
				var body map[string]string
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &body))
				assert.Equal(t, "Content-Type must be application/json", body["message"])
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	next := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	next.ServeHTTP(rr, req)

	assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rr.Header().Get("X-Frame-Options"))
}
