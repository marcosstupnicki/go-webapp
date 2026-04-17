package gowebapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	golog "github.com/marcosstupnicki/go-log"
)

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	assert.Equal(t, "/healthz", cfg.healthPath)
	assert.Equal(t, 30*time.Second, cfg.readTimeout)
	assert.Equal(t, 30*time.Second, cfg.writeTimeout)
	assert.Equal(t, 4096, cfg.maxBodyLogSize)
	assert.False(t, cfg.corsEnabled)
	assert.False(t, cfg.skipRequestBody)
	assert.False(t, cfg.skipResponseBody)
	assert.Nil(t, cfg.notFoundHandler)
	assert.Nil(t, cfg.methodNotAllowedHandler)
}

func TestWithCORS(t *testing.T) {
	cfg := defaultConfig()
	origins := []string{"https://acklane.com", "https://app.acklane.com"}
	WithCORS(origins)(&cfg)

	assert.True(t, cfg.corsEnabled)
	assert.Equal(t, origins, cfg.corsOrigins)
}

func TestWithTimeout(t *testing.T) {
	cfg := defaultConfig()
	WithTimeout(10 * time.Second)(&cfg)

	assert.Equal(t, 10*time.Second, cfg.readTimeout)
	assert.Equal(t, 10*time.Second, cfg.writeTimeout)
}

func TestWithHealthPath(t *testing.T) {
	cfg := defaultConfig()
	WithHealthPath("/health")(&cfg)

	assert.Equal(t, "/health", cfg.healthPath)
}

func TestWithNotFoundHandler(t *testing.T) {
	cfg := defaultConfig()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("custom 404"))
	}
	WithNotFoundHandler(handler)(&cfg)

	assert.NotNil(t, cfg.notFoundHandler)
}

func TestWithMaxBodyLogSize(t *testing.T) {
	cfg := defaultConfig()
	WithMaxBodyLogSize(8192)(&cfg)

	assert.Equal(t, 8192, cfg.maxBodyLogSize)
}

func TestWithSkipBodyLogging(t *testing.T) {
	cfg := defaultConfig()
	WithSkipBodyLogging(true, false)(&cfg)

	assert.True(t, cfg.skipRequestBody)
	assert.False(t, cfg.skipResponseBody)
}

func TestNew_WithOptions(t *testing.T) {
	webapp := mustNew(t, "test", "0",
		WithCORS([]string{"*"}),
		WithTimeout(5*time.Second),
		WithHealthPath("/ready"),
	)

	assert.True(t, webapp.config.corsEnabled)
	assert.Equal(t, 5*time.Second, webapp.config.readTimeout)
	assert.Equal(t, "/ready", webapp.config.healthPath)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rr := httptest.NewRecorder()
	webapp.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

	var body map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
}

func TestNew_DefaultHealthEndpoint(t *testing.T) {
	webapp := mustNew(t, "test", "0")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	webapp.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var body map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
}

func TestNew_WithLogger(t *testing.T) {
	logger, err := golog.New("test", golog.WithLevel(golog.DebugLevel))
	require.NoError(t, err)
	webapp := mustNew(t, "test", "0", WithLogger(logger))
	assert.NotNil(t, webapp.Logger)
}

func TestNew_UseMiddleware(t *testing.T) {
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

func TestNew_NotFoundHandler(t *testing.T) {
	webapp := mustNew(t, "test", "0")

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rr := httptest.NewRecorder()
	webapp.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Type"), "application/problem+json")

	var problem Problem
	err := json.Unmarshal(rr.Body.Bytes(), &problem)
	require.NoError(t, err)
	assert.Equal(t, 404, problem.Status)
}

func TestRouterMux(t *testing.T) {
	webapp := mustNew(t, "test", "0")
	mux := webapp.Router.Mux()
	assert.NotNil(t, mux)
}
