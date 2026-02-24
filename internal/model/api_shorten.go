package model

// APIShortenRequest — тело запроса POST /api/shorten.
type APIShortenRequest struct {
	URL string `json:"url"`
}

// APIShortenResponse — тело ответа POST /api/shorten (полный короткий URL).
type APIShortenResponse struct {
	Result string `json:"result"`
}
