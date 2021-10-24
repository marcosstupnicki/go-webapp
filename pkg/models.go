package gowebapp

type WebApplication struct {
	*Router
	Scope
}

type Scope struct {
	Environment string
}