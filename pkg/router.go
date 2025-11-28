package gowebapp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Router struct {
	mux *chi.Mux
}

func (r *Router) Method(method string, pattern string, handler http.HandlerFunc) {
	r.mux.Method(method, pattern, handler)
}

func (r *Router) Get(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodGet, pattern, handler)
}

func (r *Router) Head(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodHead, pattern, handler)
}

func (r *Router) Post(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodPost, pattern, handler)
}

func (r *Router) Put(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodPut, pattern, handler)
}

func (r *Router) Patch(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodPatch, pattern, handler)
}

func (r *Router) Delete(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodDelete, pattern, handler)
}

func (r *Router) Connect(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodConnect, pattern, handler)
}

func (r *Router) Options(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodOptions, pattern, handler)
}

func (r *Router) Trace(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodTrace, pattern, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
