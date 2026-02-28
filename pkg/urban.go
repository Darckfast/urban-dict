//go:build js && wasm

package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Darckfast/workers-go/cloudflare/fetch"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

const (
	CACHE_TIME = "604800"
	BASE_URL   = "https://api.urbandictionary.com/v0"
)

var client = fetch.Client{
	Timeout: 3 * time.Second,
}

func Handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := otelslog.NewLogger("urban")

	w.Header().Add("content-type", "text/plain")
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	}

	term := r.URL.Query().Get("term")
	term, err := url.QueryUnescape(term)
	term = strings.TrimSpace(term)

	if err != nil {
		logger.ErrorContext(ctx, "Error unescaping query", "error", err.Error())
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

	hexValue := fmt.Sprintf("%x", term)

	var req *http.Request
	if term == "" || hexValue == "f3a08080" {
		logger.InfoContext(ctx, "Querying random entry")
		req, _ = http.NewRequestWithContext(ctx, "GET", BASE_URL+"/random", nil)
	} else {
		req, _ = http.NewRequestWithContext(ctx, "GET", BASE_URL+"/define?term="+url.QueryEscape(term), nil)
	}

	res, err = client.Do(req)
	if err != nil {
		logger.ErrorContext(ctx, "Error requesting urban API", "error", err.Error())
		w.Write([]byte(":( no definition found for: " + term))
		return
	}
	if res.StatusCode != 200 {
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		if res.StatusCode >= 500 {
			logger.ErrorContext(ctx, "urban api is unavailable", "status", res.StatusCode, "error", string(body))
			w.Write([]byte("ops, seems like urban is unavailable"))
		}

		return
	}

	var urbanDictRes UrbanDictRes

	json.NewDecoder(res.Body).Decode(&urbanDictRes)

	if len(urbanDictRes.List) == 0 {
		w.WriteHeader(200)
		w.Write([]byte(":( no definition found for: " + term))
		logger.InfoContext(ctx, "term searched but not found: "+term+", "+hexValue)

		return
	}

	if len(urbanDictRes.List) == 10 {
		page := 2
		for {
			url := fmt.Sprintf(BASE_URL+"/define?term=%s&page=%d",
				url.QueryEscape(term),
				page,
			)
			req, _ = http.NewRequestWithContext(r.Context(), "GET", url, nil)
			res, _ = client.Do(req)

			var pagination UrbanDictRes

			json.NewDecoder(res.Body).Decode(&pagination)

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
	word := fmt.Sprintf("%s%s: %s", atUser, urbanDictRes.List[0].Word, definition)

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

	logger.InfoContext(ctx, "request completed", "status", 200)
}
