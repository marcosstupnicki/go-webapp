package gowebapp

import golog "github.com/marcosstupnicki/go-log"

type WebApp struct {
	*Router
	Scope
	Port   string
	Logger golog.Logger
}

type Scope struct {
	Environment string
}
