package main

import (
	"context"
	"errors"
	"net/http"
	"urban-dict/pkg"
	"urban-dict/pkg/otel"

	"github.com/Darckfast/workers-go/cloudflare/fetch"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	otelShutdown, err := otel.SetupOTelSDK()
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	http.Handle("GET /", otelhttp.NewHandler(http.HandlerFunc(pkg.Handler), "get-entry"))
	http.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
	})

	fetch.ServeNonBlock(nil)

	<-make(chan struct{})
}
