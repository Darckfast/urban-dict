package utils

import "time"

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
