package main

import (
	"fmt"
	gowebapp "github.com/marcosstupnicki/go-webapp/pkg"
	"net/http"
	"os"
)

const (
	ExitCodeFailToCreateWebApplication = iota
	ExitCodeFailToRunWebApplication
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	app, err := gowebapp.NewWebApplication("local")
	if err != nil {
		os.Exit(ExitCodeFailToCreateWebApplication)
	}

	userGroup := app.Group("/users")
	userGroup.Post("/", handlerPostUser)
	userGroup.Get("/{id}", handlerGetUser)
	userGroup.Put("/{id}", handlerUpdateUser)

	err = app.Run()
	if err != nil {
		fmt.Print("error booting application", err)
		os.Exit(ExitCodeFailToRunWebApplication)
	}
}

func handlerPostUser(w http.ResponseWriter, r *http.Request) {
	dummyUser := User{Name: "dummy-user"}
	fmt.Printf("user request: %+v", dummyUser)

	gowebapp.RespondWithJSON(w, 201, dummyUser)
}

func handlerGetUser(w http.ResponseWriter, r *http.Request) {
	id := gowebapp.URLParam(r, "id")
	dummyUser := User{ID: id, Name: "dummy-user"}

	gowebapp.RespondWithJSON(w, 200, dummyUser)
}

func handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	// an internal error occurred ...
	gowebapp.RespondWithError(w, 500, "Access denied for user 'root'")
}
