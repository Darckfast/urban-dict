package main

import (
	"net/http"
	"urban-dict/pkg"

	"github.com/syumai/workers"
)

func main() {
	http.HandleFunc("GET /", pkg.Handler)
	workers.Serve(nil)
}
