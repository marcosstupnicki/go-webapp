package gowebapp

import (
    "net/http"

    "github.com/go-chi/chi"
    "github.com/go-chi/chi/middleware"
)

func NewWebApp(environment string, port string) *WebApp {
	router := newRouter()
	scope := newScope(environment)

	return &WebApp{
		Router: router,
		Scope:  scope,
		Port:   port,
	}
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
//  wa.Route("/users", func(r chi.Router) {
//      r.Use(middleware.Logger)
//      r.Get("/{id}", handler)
//  })
func (wa *WebApp) Route(pattern string, fn func(r chi.Router)) {
    wa.Router.mux.Route(pattern, fn)
}

func newScope(environment string) Scope {
	return Scope{
		Environment: environment,
	}
}

func newRouter() *Router {
	mux := chi.NewRouter()

	// A good base middleware stack
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	pingHandlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	mux.Get("/ping", pingHandlerFunc)

	router := &Router{
		mux: mux,
	}
	return router
}
