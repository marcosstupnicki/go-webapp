package gowebapp

import (
	"encoding/json"
	"fmt"
	"net/http"

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
//	)
func New(environment string, port string, opts ...Option) (*WebApp, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

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
	}, nil
}

// Run starts the HTTP server. Blocks until the server returns an error.
func (wa *WebApp) Run() error {
	return http.ListenAndServe(":"+wa.Port, wa.Router.mux)
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

	// Context enrichment: adds request_id, method, path to context
	// so downstream logger.Info(ctx, ...) calls include these fields.
	mux.Use(enrichContextMiddleware)

	// Request/response logging.
	mux.Use(logRequestResponse(logger))

	// Health check endpoint.
	mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
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
