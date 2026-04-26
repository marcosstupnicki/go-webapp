package gowebapp

import (
	"context"
	"net/http"

	golog "github.com/marcosstupnicki/go-log"
)

// WebApp is the main application structure that holds the router,
// server configuration, scope, and logger.
type WebApp struct {
	*Router
	Scope
	Port   string
	Logger golog.Logger

	server *http.Server
	ctx    context.Context
	cancel context.CancelFunc
}

// Scope contains environment metadata for the application.
type Scope struct {
	Environment string
}
