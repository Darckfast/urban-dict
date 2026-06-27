package pkg

//easyjson:json
type ResultList struct {
	Definition string `json:"definition"`
	Word       string `json:"word"`
	ThumbsUp   int    `json:"thumbs_up"`
	ThumbsDown int    `json:"thumbs_down"`
	Score      int
}

//easyjson:json
type UrbanDictRes struct {
	List []ResultList `json:"list"`
}
