package main

import (
	"net/http"

	"urban-dict/api/v1"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("GET /", urban.Handler)
	http.HandleFunc("HEAD /", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	workers.Serve(nil)
}
