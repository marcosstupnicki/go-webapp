package gowebapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	golog "github.com/marcosstupnicki/go-log"
)

const defaultShutdownTimeout = 15 * time.Second

// New creates a WebApp with the given environment, port, and options.
// Returns an error if the logger cannot be built.
//
//	app, err := gowebapp.New("local", "8080",
//	    gowebapp.WithCORS([]string{"http://localhost:3001"}),
//	)
func New(environment string, port string, opts ...Option) (*WebApp, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	logger, err := golog.New(environment)
	if err != nil {
		return nil, fmt.Errorf("gowebapp: create logger: %w", err)
	}

	scope := Scope{Environment: environment}
	router := newRouter(logger, cfg)

	return &WebApp{
		Router: router,
		Scope:  scope,
		Port:   port,
		Logger: logger,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: router.mux,
		},
	}, nil
}

// Run starts the HTTP server and blocks until the server stops or the process
// receives an interrupt or termination signal.
func (wa *WebApp) Run() error {
	wa.Logger.Info(context.Background(), "http server starting", golog.Field("port", wa.Port))

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- wa.server.ListenAndServe()
	}()

	return wa.handleShutdown(serverErr)
}

func (wa *WebApp) handleShutdown(serverErr <-chan error) error {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	select {
	case err := <-serverErr:
		return normalizeServerError(err)
	case <-signalCh:
		wa.Logger.Info(context.Background(), "shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := wa.server.Shutdown(shutdownCtx); err != nil {
		if closeErr := wa.server.Close(); closeErr != nil {
			return fmt.Errorf("gowebapp: shutdown server: %w; close: %v", err, closeErr)
		}
		return fmt.Errorf("gowebapp: shutdown server: %w", err)
	}

	if err := <-serverErr; err != nil {
		return normalizeServerError(err)
	}

	wa.Logger.Info(context.Background(), "http server stopped")
	return nil
}

func normalizeServerError(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Group creates a temporary scope for mounting middleware before route definitions.
func (wa *WebApp) Group(fn func(r chi.Router)) {
	wa.Router.mux.Group(fn)
}

// Route mounts a sub-router under the given pattern, enabling
// route-scoped middleware.
func (wa *WebApp) Route(pattern string, fn func(r chi.Router)) {
	wa.Router.mux.Route(pattern, fn)
}

func newRouter(logger golog.Logger, cfg webAppConfig) *Router {
	mux := chi.NewRouter()

	// Core middleware
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)

	// Context enrichment: adds request_id, method, path to context
	// so downstream logger.Info(ctx, ...) calls include these fields.
	mux.Use(enrichContextMiddleware)

	// Request/response logging wraps recoverer so panic responses are logged.
	mux.Use(logRequestResponse(logger))
	mux.Use(middleware.Recoverer)

	if hasSecurityHeaders(cfg.securityHeaders) {
		mux.Use(securityHeadersMiddleware(cfg.securityHeaders))
	}

	// CORS middleware (if configured)
	if cfg.corsEnabled {
		mux.Use(cors.Handler(cors.Options{
			AllowedOrigins:   cfg.corsOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   cfg.corsAllowedHeaders,
			ExposedHeaders:   []string{"Link", "X-Request-Id"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	mux.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	return &Router{mux: mux}
}

// enrichContextMiddleware adds request-scoped fields to the context
// using golog.Enrich so that any downstream log call automatically
// includes request_id, method, and path.
func enrichContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		ctx := golog.Enrich(r.Context(),
			golog.Field("request_id", requestID),
			golog.Field("method", r.Method),
			golog.Field("path", r.URL.Path),
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
