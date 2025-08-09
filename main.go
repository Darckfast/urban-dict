package main

import (
	"net/http"
	"urban-dict/pkg"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("GET /", pkg.Handler)
	http.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	})

	workers.Serve(nil)
}
