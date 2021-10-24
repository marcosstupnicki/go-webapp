package gowebapp

import (
	"github.com/go-chi/chi"
	"net/http"
	"path"
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

func (r *Router) Post( pattern string, handler http.HandlerFunc) {
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
	r.mux.ServeHTTP(w,req)
}

func (r *Router) Group(pattern string) *RouterGroup {
	return &RouterGroup{router: r, path: pattern}
}

type RouterGroup struct {
	router *Router
	path string
}

func (r *RouterGroup) Method(method string, pattern string, handler http.HandlerFunc) {
	r.router.Method(method, path.Join(r.path, pattern), handler)
}

func (r *RouterGroup) Get(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodGet, pattern, handler)
}

func (r *RouterGroup) Head(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodHead, pattern, handler)
}

func (r *RouterGroup) Post( pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodPost, pattern, handler)
}

func (r *RouterGroup) Put(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodPut, pattern, handler)
}

func (r *RouterGroup) Patch(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodPatch, pattern, handler)
}

func (r *RouterGroup) Delete(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodDelete, pattern, handler)
}

func (r *RouterGroup) Connect(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodConnect, pattern, handler)
}

func (r *RouterGroup) Options(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodOptions, pattern, handler)
}

func (r *RouterGroup) Trace(pattern string, handler http.HandlerFunc) {
	r.Method(http.MethodTrace, pattern, handler)
}