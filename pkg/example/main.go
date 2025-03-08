package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	gowebapp "github.com/marcosstupnicki/go-webapp/pkg"
)

const (
	ExitCodeFailToRunWebApp = iota
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	app, err := gowebapp.NewWebApp("local", "8080")
	if err != nil {
		fmt.Print("error creating webapp", err)
		os.Exit(ExitCodeFailToRunWebApp)
	}
	// User routes with scoped middleware using Route (prefixed sub-tree)
	app.Route("/users", func(r chi.Router) {
		// Apply middlewares before defining routes
		r.Use(middleware.Logger)

		r.Post("/", handlerPostUser)
		r.Get("/{id}", handlerGetUser)
		r.Put("/{id}", handlerUpdateUser)
	})

	// Operational endpoints using Group (no prefix, shared middleware)
	app.Group(func(r chi.Router) {
		r.Use(middleware.NoCache)
		r.Get("/health", handlerHealth)
		r.Get("/metrics", handlerMetrics)
	})

	err = app.Run()
	if err != nil {
		fmt.Print("error booting application", err)
		os.Exit(ExitCodeFailToRunWebApp)
	}
}

func handlerPostUser(w http.ResponseWriter, r *http.Request) {
	dummyUser := User{ID: "uuid", Name: "dummy-user"}
	gowebapp.RespondWithJSON(w, 201, dummyUser)
}

func handlerGetUser(w http.ResponseWriter, r *http.Request) {
	id := gowebapp.URLParam(r, "id")
	dummyUser := User{ID: id, Name: "dummy-user"}
	gowebapp.RespondWithJSON(w, 200, dummyUser)
}

func handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	// an internal error occurred ...
	gowebapp.RespondWithError(w, 500, "Access denied for user")
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	_ = gowebapp.RespondWithJSON(w, 200, map[string]string{"status": "ok"})
}

func handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	_, _ = w.Write([]byte("requests_total 1\n"))
}
