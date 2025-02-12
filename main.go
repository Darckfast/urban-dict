package main

import (
	"net/http"

	"urban-dict/api/v1/urban"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("GET /api/v1/urban", urban.Handler)
	workers.Serve(nil)
}
