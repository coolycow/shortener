package model

// APIUserURLsResponse — элемент ответа GET /api/user/urls.
type APIUserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
