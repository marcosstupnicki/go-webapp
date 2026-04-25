package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	app, err := gowebapp.New("local", "8080")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create webapp: %v\n", err)
		os.Exit(1)
	}

	// User routes with scoped middleware using Route (prefixed sub-tree).
	app.Route("/users", func(r chi.Router) {
		r.Use(middleware.Logger)

		r.Post("/", handlerPostUser)
		r.Get("/{id}", handlerGetUser)
		r.Put("/{id}", handlerUpdateUser)
	})

	// Operational endpoints using Group (no prefix, shared middleware).
	app.Group(func(r chi.Router) {
		r.Use(middleware.NoCache)
		r.Get("/health", handlerHealth)
		r.Get("/metrics", handlerMetrics)
	})

	if err := app.Run(); err != nil {
		fmt.Print("error booting application", err)
		os.Exit(ExitCodeFailToRunWebApp)
	}
}

func handlerPostUser(w http.ResponseWriter, r *http.Request) {
	dummyUser := User{ID: "uuid", Name: "dummy-user"}
	_ = gowebapp.RespondWithJSON(w, 201, dummyUser)
}

func handlerGetUser(w http.ResponseWriter, r *http.Request) {
	id := gowebapp.URLParam(r, "id")
	dummyUser := User{ID: id, Name: "dummy-user"}
	_ = gowebapp.RespondWithJSON(w, 200, dummyUser)
}

func handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	_ = gowebapp.RespondWithError(w, 500, "Access denied for user")
}

func handlerHealth(w http.ResponseWriter, r *http.Request) {
	_ = gowebapp.RespondWithJSON(w, 200, map[string]string{"status": "ok"})
}

func handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)
	_, _ = w.Write([]byte("requests_total 1\n"))
}
