# go-webapp

**go-webapp** is a lightweight router for building Go HTTP services. **webapplication** is a chi wrapper that provides shortcuts to design and build REST API servers.

---
## Example

**As easy as:**

```go
package main

import (
	"fmt"
	"github.com/marcosstupnicki/go-webapp/pkg"
	"net/http"
	"os"
)

const (
	ExitCodeFailToCreateWebApplication = iota
	ExitCodeFailToRunWebApplication
)

func main()  {
	app, err := gowebapp.NewWebApplication("local")
	if err != nil {
		os.Exit(ExitCodeFailToCreateWebApplication)
	}

	userGroup := app.Group("/users")
	userGroup.Post("/", handlerPostUser)
	userGroup.Get("/{id}", handlerGetUser)

	err = app.Run()
	if err != nil {
		fmt.Print("error booting application", err)
		os.Exit(ExitCodeFailToRunWebApplication)
	}
}

func handlerPostUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, POST")
}

func handlerGetUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, GET")
}
```