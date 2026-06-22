//go:build js && wasm

package pkg

import (
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mailru/easyjson"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
)

const (
	CACHE_TIME = "604800"
	BASE_URL   = "https://api.urbandictionary.com/v0"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer("trace").Start(r.Context(), "http.server")
	defer span.End()
	slog.SetDefault(otelslog.NewLogger("urban"))

	w.Header().Add("content-type", "text/plain")
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	}

	term := r.URL.Query().Get("term")
	term, err := url.QueryUnescape(term)
	term = strings.TrimSpace(term)

	if err != nil {
		slog.ErrorContext(ctx, "Error unescaping query", "error", err.Error())
		w.Write([]byte(":( no definition found for: " + term))
		return
	}

	atUser := ""
	if len(term) > 0 {
		if term[0] == '!' {
			termSplitted := strings.Split(term, " ")
			termSplitted = termSplitted[1:]
			term = strings.Join(termSplitted, " ")
		}

		if strings.Contains(term, "@") {
			termSplitted := strings.Split(term, " ")
			term = ""

			for _, word := range termSplitted {
				if len(word) == 0 {
					continue
				}

				if word[0] == '@' {
					atUser = word + " "
					continue
				}

				term = term + word + " "
			}
		}
	}
	var res *http.Response

	hexValue := hex.EncodeToString([]byte(term))

	var req *http.Request
	if term == "" || hexValue == "f3a08080" {
		slog.InfoContext(ctx, "Querying random entry")
		req, err = http.NewRequestWithContext(ctx, "GET", BASE_URL+"/random", nil)

		if err != nil {
			slog.ErrorContext(ctx, "error creating request", "err", err)
		}
	} else {
		slog.InfoContext(ctx, "Querying entry"+term)
		req, _ = http.NewRequestWithContext(ctx, "GET", BASE_URL+"/define?term="+url.QueryEscape(term), nil)
	}

	client := http.Client{
		Timeout:   2 * time.Second,
		Transport: http.DefaultTransport,
	}

	res, err = client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "Error requesting urban API", "error", err.Error())
		w.Write([]byte(":( no definition found for: " + term))
		return
	}

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()

		slog.ErrorContext(ctx, "urban api return non 2XX", "status", res.StatusCode, "error", string(body))
		if res.StatusCode >= 500 {
			w.Write([]byte("ops, seems like urban is unavailable"))
		}

		return
	}

	var urbanDictRes UrbanDictRes
	easyjson.UnmarshalFromReader(res.Body, &urbanDictRes)
	res.Body.Close()

	if len(urbanDictRes.List) == 0 {
		w.WriteHeader(200)
		w.Write([]byte(":( no definition found for: " + term))
		slog.InfoContext(ctx, "term searched but not found: "+term+", "+hexValue)

		return
	}

	if len(urbanDictRes.List) == 10 && term != "" {
		page := 2
		for {
			url := BASE_URL + "/define?term=" + url.QueryEscape(term) + "&page=" + strconv.Itoa(page)
			req, _ = http.NewRequestWithContext(ctx, "GET", url, nil)
			res, err = client.Do(req)

			if err != nil {
				slog.ErrorContext(ctx, "error requesting Urban API", "err", err)
				break
			}

			if res.StatusCode > 299 {
				body, _ := io.ReadAll(res.Body)
				res.Body.Close()
				slog.ErrorContext(ctx, "Urban API returned non 2xx", "err", string(body))
				break
			}

			var pagination UrbanDictRes
			easyjson.UnmarshalFromReader(res.Body, &pagination)
			res.Body.Close()

			urbanDictRes.List = append(urbanDictRes.List, pagination.List...)

			if len(pagination.List) != 10 {
				break
			}
			page = page + 1

			if page == 3 {
				break
			}
		}
	}

	for i, entry := range urbanDictRes.List {
		if entry.ThumbsDown < 0 {
			entry.ThumbsDown = entry.ThumbsDown * -1
		}

		urbanDictRes.List[i].OriginalIndex = i
		urbanDictRes.List[i].Score = entry.ThumbsUp - entry.ThumbsDown
	}

	sort.Slice(urbanDictRes.List, func(i, j int) bool {
		return urbanDictRes.List[i].Score > urbanDictRes.List[j].Score
	})

	definition := urbanDictRes.List[0].Definition
	definition = strings.ReplaceAll(definition, "[", "")
	definition = strings.ReplaceAll(definition, "]", "")
	word := atUser + urbanDictRes.List[0].Word + ": " + definition

	if strings.HasPrefix(word, "/") {
		strings.Replace(word, "/", "", 1)
	}

	if strings.HasPrefix(word, "!") {
		strings.Replace(word, "!", "", 1)
	}

	if len(word) > 400 {
		word = word[:397] + "..."
	}

	w.WriteHeader(200)
	w.Header().Set("Cache-Control", "public, max-age="+CACHE_TIME)
	w.Write([]byte(word))

	slog.InfoContext(ctx, "request completed", "status", 200)
}
