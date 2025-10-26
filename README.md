# go-webapp

go-webapp is a lightweight wrapper around chi for building HTTP services with simple routing helpers and sensible middleware defaults.

---
## Quick Start

```go
app := gowebapp.NewWebApp("local", "8080")

// Preferred: Route - group a sub-tree under a prefix
app.Route("/users", func(r chi.Router) {
    r.Get("/{id}", handlerGetUser)
})

// Alternative: Group - apply shared middleware without changing the prefix
app.Group(func(r chi.Router) {
    r.Get("/health", handlerHealth)
    r.Get("/metrics", handlerMetrics)
})

_ = app.Run()
```

## Route vs Group
- Route: mounts a subrouter under a prefix (e.g., `/users`). Middlewares declared inside apply to the whole sub-tree. Prefer Route when endpoints share a clear prefix.
- Group: creates a block without a prefix to apply middleware to multiple absolute paths. Use it when there is no natural common prefix.
- chi rule: call `r.Use(...)` before declaring routes inside the same block; chi does not allow adding middleware after routes.

Tip: for a single route with middleware, use `r.With(mw...).Get("/path", h)`.

Project convention: Prefer `Route` for prefixed sections, use `Group` sparingly when there is no shared prefix, and use `r.With(...)` for one-off route middleware.

## Commands
- Build: `go build -o bin/webapp .`
- Run: `go run .`
- Test: `go test ./... -v`
- Format/Vet: `go fmt ./...` and `go vet ./...`
