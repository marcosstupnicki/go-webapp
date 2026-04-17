package gowebapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	golog "github.com/marcosstupnicki/go-log"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()
	assert.False(t, cfg.corsEnabled)
	assert.Nil(t, cfg.logger)
}

func TestOptions(t *testing.T) {
	tests := []struct {
		name   string
		opt    Option
		check  func(t *testing.T, cfg webAppConfig)
	}{
		{
			name: "WithCORS enables CORS",
			opt:  WithCORS([]string{"https://acklane.com"}),
			check: func(t *testing.T, cfg webAppConfig) {
				assert.True(t, cfg.corsEnabled)
				assert.Equal(t, []string{"https://acklane.com"}, cfg.corsOrigins)
			},
		},
		{
			name: "WithLogger sets logger",
			opt: func() Option {
				l, err := golog.New("test")
				require.NoError(t, err)
				return WithLogger(l)
			}(),
			check: func(t *testing.T, cfg webAppConfig) {
				assert.NotNil(t, cfg.logger)
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
			name:       "health endpoint returns 200",
			method:     http.MethodGet,
			path:       "/health",
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
