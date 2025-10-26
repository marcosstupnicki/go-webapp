package gowebapp

type WebApp struct {
	*Router
	Scope
	Port string
}

type Scope struct {
	Environment string
}
