package model

type ApiShortenRequest struct {
	URL string `json:"url"`
}

type ApiShortenResponse struct {
	Result string `json:"result"`
}
