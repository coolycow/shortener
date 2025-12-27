package model

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	error2 "github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/logger"
	"go.uber.org/zap"
)

type ShortURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	Key           string `json:"key"`
}

// UnmarshalJSON нужна для специальной проверки входных данных
func (s *ShortURL) UnmarshalJSON(data []byte) (err error) {
	type ShortURLAlias ShortURL

	aliasValue := &struct {
		*ShortURLAlias
		OriginalURL string `json:"original_url"`
	}{
		ShortURLAlias: (*ShortURLAlias)(s),
	}

	if err = json.Unmarshal(data, aliasValue); err != nil {
		return
	}

	s.OriginalURL, err = parseOriginalURL(aliasValue.OriginalURL)

	return
}

// ToAPIShortenBatchResponse переводит в подходящую структуру ответа
func (s *ShortURL) ToAPIShortenBatchResponse(baseURL string) APIShortenBatchResponse {
	return APIShortenBatchResponse{
		CorrelationID: s.CorrelationID,
		ShortURL:      baseURL + `/` + s.Key,
	}
}

// ToAPIUserURLsResponse переводит в подходящую структуру ответа
func (s *ShortURL) ToAPIUserURLsResponse(baseURL string) APIUserURLsResponse {
	return APIUserURLsResponse{
		ShortURL:    baseURL + `/` + s.Key,
		OriginalURL: s.OriginalURL,
	}
}

// parseOriginalURL отдельная функция для парсинга исходной URL
func parseOriginalURL(originalURL string) (string, error) {
	// Извлекаем строку из тела запроса и проверяем, что она не пуста
	trimURL := strings.TrimSpace(originalURL)

	if trimURL == "" {
		logger.Log.Debug("trim URL is empty")
		return "", error2.CustomError{
			Message:    "Empty URL",
			StatusCode: http.StatusBadRequest,
		}
	}

	// Парсим URL из строки (фактически проверяем, что это действительно URL)
	validURL, err := url.ParseRequestURI(trimURL)

	if err != nil {
		logger.Log.Debug("invalid URL", zap.Error(err))
		return "", error2.CustomError{
			Message:    "Invalid URL",
			StatusCode: http.StatusBadRequest,
		}
	}

	return validURL.String(), nil
}
