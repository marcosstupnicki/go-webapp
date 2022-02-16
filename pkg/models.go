package gowebapp

type WebApp struct {
	*Router
	Scope
}

type Scope struct {
	Environment string
}