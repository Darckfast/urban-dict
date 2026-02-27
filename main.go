//go:build js && wasm

package main

import (
	"net/http"
	"urban-dict/pkg"

	"github.com/Darckfast/workers-go/cloudflare/fetch"
)

func main() {

	http.HandleFunc("GET /", pkg.Handler)
	http.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	})

	fetch.ServeNonBlock(nil)

	<-make(chan struct{})
}
