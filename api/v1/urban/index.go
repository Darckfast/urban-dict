package urban

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"

	"urban-dict/pkg/utils"

	multilogger "github.com/Darckfast/multi_logger/pkg/multi_logger"
)

const (
	CACHE_TIME        = "604800"
	CACHE_RANDOM_TIME = "10"
	BASE_URL          = "https://api.urbandictionary.com/v0"
)

var logger = slog.New(multilogger.NewHandler(os.Stdout))

func Handler(writer http.ResponseWriter, request *http.Request) {
	ctx, wg := multilogger.SetupContext(&multilogger.SetupOps{
		Request:           request,
		BaselimeApiKey:    os.Getenv("BASELIME_API_KEY"),
		BetterStackApiKey: os.Getenv("BETTERSTACK_API_KEY"),
		AxiomApiKey:       os.Getenv("AXIOM_API_KEY"),
		ServiceName:       os.Getenv("VERCEL_GIT_REPO_SLUG"),
	})

	defer func() {
		wg.Wait()
		ctx.Done()
	}()

	writer.Header().Add("content-type", "text/plain")
	logger.InfoContext(ctx, "processing request")
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
	if term == "" || hexValue == "f3a08080" {
		isRandom = true
		logger.InfoContext(ctx, "Querying random entry")
		res, err = http.Get(BASE_URL + "/random")
	} else {
		res, err = http.Get(BASE_URL + "/define?term=" + url.QueryEscape(term))
	}

	if err != nil {
		logger.ErrorContext(ctx, "Error requesting urban API", "error", err.Error())
		writer.Write([]byte(":( no definition found for: " + term))
		return
	}
	if res.StatusCode != 200 {
		writer.WriteHeader(res.StatusCode)

		defer res.Body.Close()

		body, _ := io.ReadAll(res.Body)
		logger.ErrorContext(ctx, "Error calling urban api", "status", res.StatusCode, "error", string(body))
		writer.Write([]byte("ops, something went wrong, wake up @darckfast and fix this"))
		return
	}

	var urbanDictRes utils.UrbanDictRes

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
			res, _ = http.Get(url)

			var pagination utils.UrbanDictRes

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
