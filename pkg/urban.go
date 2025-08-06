package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	multilogger "github.com/Darckfast/multi_logger/pkg/multi_logger"
	"github.com/syumai/workers/cloudflare"
	"github.com/syumai/workers/cloudflare/fetch"
)

const (
	CACHE_TIME        = "604800"
	CACHE_RANDOM_TIME = "604800"
	BASE_URL          = "https://api.urbandictionary.com/v0"
)

var logger = slog.New(multilogger.NewHandler(os.Stdout))

func Handler(writer http.ResponseWriter, request *http.Request) {
	ctx, wg := multilogger.SetupContext(&multilogger.SetupOps{
		Request:     request,
		AxiomApiKey: cloudflare.Getenv("AXIOM_API_KEY"),
		ServiceName: cloudflare.Getenv("VERCEL_GIT_REPO_SLUG"),
		RequestGen: func(args multilogger.SendLogsArgs) {
			args.MaxQueue <- 1
			args.Wg.Add(1)

			req, _ := fetch.NewRequest(args.Ctx, args.Method, args.Url, bytes.NewBuffer(*args.Body))
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", args.Bearer)

			client := fetch.NewClient()

			go func() {
				defer args.Wg.Done()
				client.Do(req, nil)
				<-args.MaxQueue
			}()
		},
	})

	defer func() {
		wg.Wait()
		ctx.Done()
	}()

	writer.Header().Add("content-type", "text/plain")
	origin := request.Header.Get("Origin")
	if origin != "" {
		writer.Header().Set("Access-Control-Allow-Origin", request.Header.Get("Origin"))
	}

	term := request.URL.Query().Get("term")
	term, err := url.QueryUnescape(term)
	term = strings.TrimSpace(term)

	if err != nil {
		logger.ErrorContext(ctx, "Error unescaping query", "error", err.Error())
		writer.Write([]byte(":( no definition found for: " + term))
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

	isRandom := false
	client := fetch.NewClient()

	var req *fetch.Request
	if term == "" || hexValue == "f3a08080" {
		isRandom = true
		logger.InfoContext(ctx, "Querying random entry")
		req, _ = fetch.NewRequest(request.Context(), "GET", BASE_URL+"/random", nil)
	} else {
		req, _ = fetch.NewRequest(request.Context(), "GET", BASE_URL+"/define?term="+url.QueryEscape(term), nil)
	}

	res, err = client.Do(req, nil)
	if err != nil {
		logger.ErrorContext(ctx, "Error requesting urban API", "error", err.Error())
		writer.Write([]byte(":( no definition found for: " + term))
		return
	}
	if res.StatusCode != 200 {
		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		if res.StatusCode == 503 {
			logger.ErrorContext(ctx, "urban api is unavailable", "status", res.StatusCode, "error", string(body))
			writer.Write([]byte("ops, seems like urban is unavailable"))
		} else {
			writer.Write([]byte("ops, something went wrong, wake up @darckfast and fix this"))
			logger.ErrorContext(ctx, "Error calling urban api", "status", res.StatusCode, "error", string(body))
		}
		return
	}

	var urbanDictRes UrbanDictRes

	json.NewDecoder(res.Body).Decode(&urbanDictRes)

	if len(urbanDictRes.List) == 0 {
		writer.WriteHeader(200)
		writer.Write([]byte(":( no definition found for: " + term))
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
			req, _ = fetch.NewRequest(request.Context(), "GET", url, nil)
			res, _ = client.Do(req, nil)

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

	if len(word) > 400 {
		word = word[:397] + "..."
	}

	writer.WriteHeader(200)

	if isRandom {
		writer.Header().Set("Cache-Control", "public, max-age="+CACHE_RANDOM_TIME)
	} else {
		writer.Header().Set("Cache-Control", "public, max-age="+CACHE_TIME)
	}

	writer.Write([]byte(word))

	logger.InfoContext(ctx, "request completed", "status", 200)
}
