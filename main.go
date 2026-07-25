//go:build js && wasm

package main

import (
	"log/slog"
	"net/http"
	"urban-dict/pkg"

	"codeberg.org/darckfast/workers-go/platform/cloudflare/durableobjects"
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

	fetch.ServeNonBlock(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stub, err := durableobjects.NewDurableObjectNamespace("RUSTY_LIMITER")
		if err != nil {
			slog.Error("error retrieving RUSTY_LIMITER", "err", err)
		} else {
			id := r.Header.Get("cf-connecting-ip")
			if id == "" {
				id = "127.0.0.1"
			}

			id += r.Header.Get("x-streamelements-channel")
			id += r.Header.Get("nightbot-channel")
			id += r.Header.Get("x-fossabot-channelid")
			id += r.Header.Get("moobot-channel-id")

			do := stub.GetByName(id)
			dummy_req, _ := http.NewRequest("GET", "http://dummy?max_reqs=20", nil)
			rs, err := do.Fetch(dummy_req)
			if err != nil {
				slog.Error("error checking rate-limit", "err", err)
				w.Write([]byte("ops, something went wrong"))
				return
			}

			if rs.StatusCode == 429 {
				slog.Warn("rate-limit exceeded", "id", id, "cooldown", rs.Header.Get("retry-after"))
				w.WriteHeader(429)
				w.Write([]byte("slowdown, we are getting overloaded"))
				w.Header().Set("retry-after", rs.Header.Get("retry-after"))
				return
			}
		}

		router.ServeHTTP(w, r)
	}))

	<-make(chan struct{})
}
