package model

// APIShortenBatchResponse — элемент ответа POST /api/shorten/batch.
type APIShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
