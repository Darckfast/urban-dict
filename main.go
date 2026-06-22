//go:build js && wasm

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"urban-dict/pkg"
	"urban-dict/pkg/otel"

	"codeberg.org/darckfast/workers-go/platform/cloudflare/fetch"
	"codeberg.org/darckfast/workers-go/platform/cloudflare/lifecycle"
	"github.com/julienschmidt/httprouter"
)

func main() {
	otelShutdown, err := otel.SetupOTelSDK()

	if err != nil {
		slog.Error("error setting otel", "err", err)
	}

	defer lifecycle.Ctx.WaitUntil(func() error {
		return errors.Join(err, otelShutdown(context.Background()))
	})

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
