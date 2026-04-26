package gowebapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	assert.False(t, cfg.corsEnabled)
	assert.Empty(t, cfg.corsOrigins)
	assert.Equal(t, []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"}, cfg.corsAllowedHeaders)
	assert.Empty(t, cfg.securityHeaders)
}

func TestOptions(t *testing.T) {
	tests := []struct {
		name  string
		opt   Option
		check func(t *testing.T, cfg webAppConfig)
	}{
		{
			name: "WithCORS enables CORS",
			opt:  WithCORS([]string{"https://example.com"}),
			check: func(t *testing.T, cfg webAppConfig) {
				assert.True(t, cfg.corsEnabled)
				assert.Equal(t, []string{"https://example.com"}, cfg.corsOrigins)
			},
		},
		{
			name: "WithCORSAllowedHeaders configures allowed headers",
			opt:  WithCORSAllowedHeaders([]string{"Authorization", "X-Org-ID"}),
			check: func(t *testing.T, cfg webAppConfig) {
				assert.Equal(t, []string{"Authorization", "X-Org-ID"}, cfg.corsAllowedHeaders)
			},
		},
		{
			name: "WithSecurityHeaders enables security headers",
			opt:  WithSecurityHeaders(map[string]string{"X-Test": "ok"}),
			check: func(t *testing.T, cfg webAppConfig) {
				assert.Equal(t, map[string]string{"X-Test": "ok"}, cfg.securityHeaders)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			tt.opt(&cfg)
			tt.check(t, cfg)
		})
	}
}

func TestNew_Endpoints(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantJSON   map[string]interface{}
	}{
		{
			name:       "ping endpoint returns 200",
			method:     http.MethodGet,
			path:       "/ping",
			wantStatus: http.StatusOK,
			wantJSON:   map[string]interface{}{"status": "ok"},
		},
		{
			name:       "not found returns 404",
			method:     http.MethodGet,
			path:       "/nonexistent",
			wantStatus: http.StatusNotFound,
		},
	}

	webapp := mustNew(t, "test", "0")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			webapp.Router.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantJSON != nil {
				var got map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &got)
				require.NoError(t, err)
				for k, v := range tt.wantJSON {
					assert.Equal(t, v, got[k])
				}
			}
		})
	}
}

func TestNew_GroupMiddleware(t *testing.T) {
	webapp := mustNew(t, "test", "0")

	webapp.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Custom", "test-value")
				next.ServeHTTP(w, r)
			})
		})
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	webapp.Router.ServeHTTP(rr, req)

	assert.Equal(t, "test-value", rr.Header().Get("X-Custom"))
}

func TestNew_RouterMux(t *testing.T) {
	webapp := mustNew(t, "test", "0")
	assert.NotNil(t, webapp.Router.Mux())
}

func TestNew_SecurityHeaders(t *testing.T) {
	webapp := mustNew(t, "test", "0", WithSecurityHeaders(DefaultSecurityHeaders()))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	webapp.Router.ServeHTTP(rr, req)

	assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rr.Header().Get("X-Frame-Options"))
	assert.Equal(t, "default-src 'self';", rr.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", rr.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "strict-origin-when-cross-origin", rr.Header().Get("Referrer-Policy"))
	assert.Equal(t, "geolocation=(), camera=()", rr.Header().Get("Permissions-Policy"))
}

func TestNew_CORSAllowedHeaders(t *testing.T) {
	webapp := mustNew(t, "test", "0",
		WithCORS([]string{"https://dashboard.example"}),
		WithCORSAllowedHeaders([]string{"Authorization", "X-Org-ID"}),
	)

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "https://dashboard.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	req.Header.Set("Access-Control-Request-Headers", "Authorization, X-Org-ID")
	rr := httptest.NewRecorder()

	webapp.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "X-Org-Id")
}
