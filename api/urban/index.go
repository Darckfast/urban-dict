package urban

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type UrbanDictRes struct {
	List []struct {
		Definition  string    `json:"definition"`
		Permalink   string    `json:"permalink"`
		ThumbsUp    int       `json:"thumbs_up"`
		Author      string    `json:"author"`
		Word        string    `json:"word"`
		Defid       int       `json:"defid"`
		CurrentVote string    `json:"current_vote"`
		WrittenOn   time.Time `json:"written_on"`
		Example     string    `json:"example"`
		ThumbsDown  int       `json:"thumbs_down"`
	} `json:"list"`
}

func Handler(writer http.ResponseWriter, request *http.Request) {
	term := request.URL.Query().Get("term")

	var res *http.Response

	hexValue := fmt.Sprintf("%x", term)

	if term == "" || hexValue == "f3a08080" {
		res, _ = http.Get("https://api.urbandictionary.com/v0/random")
	} else {
		res, _ = http.Get("https://api.urbandictionary.com/v0/define?term=" + term)
	}

	if res.StatusCode != 200 {
		writer.WriteHeader(res.StatusCode)

		defer res.Body.Close()

		body, _ := io.ReadAll(res.Body)
		log.Println("Error calling urban api", res.StatusCode, string(body))
		return
	}

	var urbanDictRes UrbanDictRes

	json.NewDecoder(res.Body).Decode(&urbanDictRes)

	if len(urbanDictRes.List) == 0 {
		writer.WriteHeader(200)
		writer.Write([]byte(":( no definition found for: " + term))

		return
	}

	definition := urbanDictRes.List[0].Definition
	definition = strings.ReplaceAll(definition, "[", "")
	definition = strings.ReplaceAll(definition, "]", "")
	word := fmt.Sprintf("%s: %s", urbanDictRes.List[0].Word, definition)

	writer.WriteHeader(200)
	writer.Write([]byte(word))
}
