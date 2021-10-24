package gowebapp

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"net/http"
)

func NewWebApplication(environment string) (*WebApplication, error) {
	router := newRouter()
	scope := newScope(environment)

	return &WebApplication{
		Router: router,
		Scope: scope,
	}, nil
}

func (wa *WebApplication)Run() error{
	return http.ListenAndServe(":8080", wa.mux)
}

func newScope(environment string) Scope{
	return Scope{
		Environment: environment,
	}
}

func newRouter() *Router{
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