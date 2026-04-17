package gowebapp

import (
	"context"
	"encoding/json"
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

// New creates a WebApp with the given environment, port, and options.
// Returns an error if the logger cannot be built.
//
//	app, err := gowebapp.New("local", "8080",
//	    gowebapp.WithCORS([]string{"http://localhost:3001"}),
//	    gowebapp.WithTimeout(30 * time.Second),
//	)
func New(environment string, port string, opts ...Option) (*WebApp, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	// Resolve logger: explicit > default
	var logger golog.Logger
	if cfg.logger != nil {
		logger = *cfg.logger
	} else {
		var err error
		logger, err = golog.New(environment)
		if err != nil {
			return nil, fmt.Errorf("gowebapp: create logger: %w", err)
		}
	}

	scope := Scope{Environment: environment}
	router := newRouter(logger, cfg)

	return &WebApp{
		Router: router,
		Scope:  scope,
		Port:   port,
		Logger: logger,
		config: cfg,
	}, nil
}

// Run starts the HTTP server with graceful shutdown on SIGINT/SIGTERM.
// It blocks until the server is shut down.
func (wa *WebApp) Run() error {
	wa.server = &http.Server{
		Addr:         ":" + wa.Port,
		Handler:      wa.Router.mux,
		ReadTimeout:  wa.config.readTimeout,
		WriteTimeout: wa.config.writeTimeout,
	}

	errCh := make(chan error, 1)

	go func() {
		if err := wa.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := wa.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}

// Use appends one or more middleware to the router stack.
// Must be called before registering routes.
func (wa *WebApp) Use(middlewares ...func(http.Handler) http.Handler) {
	wa.Router.Use(middlewares...)
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
	mux.Use(middleware.Recoverer)

	// CORS middleware (if configured)
	if cfg.corsEnabled {
		mux.Use(cors.Handler(cors.Options{
			AllowedOrigins:   cfg.corsOrigins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
			ExposedHeaders:   []string{"Link", "X-Request-ID"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	// Context enrichment middleware: adds request_id, method, path to context
	// so any logger.Info(ctx, ...) call downstream includes these fields.
	mux.Use(enrichContextMiddleware)

	// Logging middleware with body size limits
	mux.Use(logRequestResponse(logger, cfg))

	// Custom error handlers
	if cfg.notFoundHandler != nil {
		mux.NotFound(cfg.notFoundHandler)
	} else {
		mux.NotFound(defaultNotFoundHandler)
	}

	if cfg.methodNotAllowedHandler != nil {
		mux.MethodNotAllowed(cfg.methodNotAllowedHandler)
	} else {
		mux.MethodNotAllowed(defaultMethodNotAllowedHandler)
	}

	// Health check endpoint (JSON response)
	mux.Get(cfg.healthPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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

// defaultNotFoundHandler returns a Problem Details 404.
func defaultNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	WriteError(w, NotFound("the requested resource was not found"))
}

// defaultMethodNotAllowedHandler returns a Problem Details 405.
func defaultMethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	writeProblem(w, http.StatusMethodNotAllowed, "Method Not Allowed", "the HTTP method is not allowed for this resource")
}
