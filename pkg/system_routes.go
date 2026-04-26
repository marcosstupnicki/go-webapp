package gowebapp

import (
	"encoding/json"
	"net/http"
)

func (r *Router) mountSystemRoutes() {
	r.systemRoutes.Do(func() {
		r.mux.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		})
	})
}
