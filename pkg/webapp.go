package gowebapp

import (
	"context"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	return &WebApp{
		Router: router,
		Scope:  scope,
		Port:   port,
		Logger: logger,
		server: &http.Server{
			Addr:    ":" + port,
			Handler: router.mux,
		},
		ctx:  ctx,
		stop: stop,
	}, nil
}

// Context returns the root application context. It is canceled when the
// process receives an interrupt or termination signal.
func (wa *WebApp) Context() context.Context {
	return wa.ctx
}

// Run starts the HTTP server and blocks until the server stops or the process
// receives an interrupt or termination signal.
func (wa *WebApp) Run() error {
	defer wa.stop()

	wa.Router.mountSystemRoutes()

	errCh := make(chan error, 1)
	wa.Logger.Info(wa.ctx, "http server starting", golog.Field("port", wa.Port))

	go func() {
		if err := wa.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-wa.ctx.Done():
	}

	wa.Logger.Info(context.Background(), "shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := wa.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("gowebapp: shutdown server: %w", err)
	}

	if err := <-errCh; err != nil {
		return err
	}

	wa.Logger.Info(context.Background(), "http server stopped")
	return nil
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
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Org-ID", "X-Request-Id"},
			ExposedHeaders:   []string{"Link", "X-Request-Id"},
			AllowCredentials: true,
			MaxAge:           300,
		}))
	}

	if cfg.securityHeaders {
		mux.Use(SecurityHeaders)
	}

	// Context enrichment: adds request_id, method, path to context
	// so downstream logger.Info(ctx, ...) calls include these fields.
	mux.Use(enrichContextMiddleware)

	// Request/response logging.
	mux.Use(logRequestResponse(logger))

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
