//go:build js && wasm

package main

import (
	"net/http"
	"urban-dict/pkg"

	"codeberg.org/darckfast/workers-go/platform/cloudflare/fetch"
	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	router.Handler("GET", "/api/urban", http.HandlerFunc(pkg.Handler))
	router.HandlerFunc("OPTIONS", "/api/urban", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	})

	fetch.ServeNonBlock(router)

	<-make(chan struct{})
}
