package gowebapp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	golog "github.com/marcosstupnicki/go-log"
)

func NewWebApp(environment string, port string) (*WebApp, error) {
	scope := newScope(environment)
	logger, err := newLogger()
	if err != nil {
		return nil, err
	}

	router := newRouter(logger)

	return &WebApp{
		Router: router,
		Scope:  scope,
		Port:   port,
		Logger: logger,
	}, nil
}

func (wa *WebApp) Run() error {
	return http.ListenAndServe(":"+wa.Port, wa.Router.mux)
}

// Group creates a temporary scope for mounting middleware before route definitions.
// Note: Middlewares must be registered before adding routes inside the fn.
func (wa *WebApp) Group(fn func(r chi.Router)) {
	wa.Router.mux.Group(fn)
}

// Route mounts a sub-router under the given pattern, enabling route-scoped middleware.
// Example:
//
//	wa.Route("/users", func(r chi.Router) {
//	    r.Use(middleware.Logger)
//	    r.Get("/{id}", handler)
//	})
func (wa *WebApp) Route(pattern string, fn func(r chi.Router)) {
	wa.Router.mux.Route(pattern, fn)
}
func newScope(environment string) Scope {
	return Scope{
		Environment: environment,
	}
}

func newRouter(logger golog.Logger) *Router {
	mux := chi.NewRouter()

	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(logRequestResponse(logger))

	pingHandlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	mux.Get("/ping", pingHandlerFunc)

	router := &Router{
		mux: mux,
	}
	return router
}

func newLogger() (golog.Logger, error) {
	return golog.New()
}
