package main

import (
	"net/http"

	urban "urban-dict/api/v1"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("GET /", urban.Handler)
	workers.Serve(nil)
}
