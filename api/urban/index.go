package urban

import (
	"encoding/json"
	"log"
	"net/http"
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
	if term == "" {
		res, _ = http.Get("https://api.urbandictionary.com/v0/random")
	} else {
		res, _ = http.Get("https://api.urbandictionary.com/v0/define?term=" + term)
	}

	if res.StatusCode != 200 {
		writer.WriteHeader(res.StatusCode)

		log.Println("Error calling urban api", res.StatusCode)
		return
	}

	var urbanDictRes UrbanDictRes

	json.NewDecoder(res.Body).Decode(&urbanDictRes)

	log.Println(urbanDictRes)

	writer.WriteHeader(200)
	writer.Write([]byte("it works!"))
}
