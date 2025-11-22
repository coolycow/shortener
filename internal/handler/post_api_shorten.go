package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PostAPIShortenHandler обрабатывает POST-запросы к серверу
func PostAPIShortenHandler(service service.URLService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, что тип контента - application/json
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			logger.Log.Debug("content type is not application/json")
			_ = c.Error(error.CustomError{
				Message:    "Content type not allowed",
				StatusCode: http.StatusUnsupportedMediaType,
			})
			return
		}

		// Читаем тело запроса
		var req model.APIShortenRequest
		dec := json.NewDecoder(c.Request.Body)

		if err := dec.Decode(&req); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			_ = c.Error(err)
			return
		}

		// Извлекаем строку из тела запроса и проверяем, что она не пуста
		trimURL := strings.TrimSpace(req.URL)

		if trimURL == "" {
			logger.Log.Debug("trim URL is empty")
			_ = c.Error(error.CustomError{
				Message:    "Empty URL",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Парсим URL из строки (фактически проверяем, что это действительно URL)
		validURL, err := url.ParseRequestURI(trimURL)

		if err != nil {
			logger.Log.Debug("invalid URL", zap.Error(err))
			_ = c.Error(error.CustomError{
				Message:    "Invalid URL",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		shortURL, err := service.CreateShortURL(validURL.String())

		if err != nil {
			logger.Log.Debug("cannot create short URL", zap.Error(err))
			_ = c.Error(err)
			return
		}

		logger.Log.Debug("create short url", zap.String("url", shortURL))

		resp := model.APIShortenResponse{
			Result: shortURL,
		}

		c.JSON(http.StatusCreated, resp)
	}
}
