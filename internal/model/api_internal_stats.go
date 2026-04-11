package model

// APIInternalStats — ответ GET /api/internal/stats.
type APIInternalStats struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}
