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
	"time"
)

type UrbanDictRes struct {
	List []struct {
		Definition    string    `json:"definition"`
		Permalink     string    `json:"permalink"`
		ThumbsUp      int       `json:"thumbs_up"`
		Author        string    `json:"author"`
		Word          string    `json:"word"`
		Defid         int       `json:"defid"`
		CurrentVote   string    `json:"current_vote"`
		WrittenOn     time.Time `json:"written_on"`
		Example       string    `json:"example"`
		ThumbsDown    int       `json:"thumbs_down"`
		Score         int
		OriginalIndex int
	} `json:"list"`
}

const (
	CACHE_TIME        = "604800"
	CACHE_RANDOM_TIME = "10"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func Handler(writer http.ResponseWriter, request *http.Request) {
	logger.Info("Incoming request")
	term := request.URL.Query().Get("term")
	term, _ = url.QueryUnescape(term)
	term = strings.TrimSpace(term)

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
		logger.Info("Querying random entry")
		res, _ = http.Get("https://api.urbandictionary.com/v0/random")
	} else {
		res, _ = http.Get("https://api.urbandictionary.com/v0/define?term=" + url.QueryEscape(term))
	}

	if res.StatusCode != 200 {
		writer.WriteHeader(res.StatusCode)

		defer res.Body.Close()

		body, _ := io.ReadAll(res.Body)
		logger.Error("Error calling urban api", "status", res.StatusCode, "body", string(body))
		writer.Write([]byte("ops, something went wrong, wake up @darckfast and fix this"))
		return
	}

	var urbanDictRes UrbanDictRes

	json.NewDecoder(res.Body).Decode(&urbanDictRes)

	if len(urbanDictRes.List) == 0 {
		writer.WriteHeader(200)
		writer.Write([]byte(":( no definition found for: " + term))
		logger.Info("term searched but not found: " + term + ", " + hexValue)

		return
	}

	if len(urbanDictRes.List) == 10 {
		page := 2
		for {
			url := fmt.Sprintf("https://api.urbandictionary.com/v0/define?term=%s&page=%d",
				url.QueryEscape(term),
				page,
			)
			res, _ = http.Get(url)

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
		writer.Header().Set("CDN-Cache-Control", "public, max-age="+CACHE_RANDOM_TIME)
		writer.Header().Set("Vercel-CDN-Cache-Control", "public, max-age="+CACHE_RANDOM_TIME)
	} else {
		writer.Header().Set("Cache-Control", "public, max-age="+CACHE_TIME)
		writer.Header().Set("CDN-Cache-Control", "public, max-age="+CACHE_TIME)
		writer.Header().Set("Vercel-CDN-Cache-Control", "public, max-age="+CACHE_TIME)
	}

	writer.Write([]byte(word))

	logger.Info("Sending response", "status", 200)
}
