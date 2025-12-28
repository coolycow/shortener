package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/logger"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeleteAPIUserURLs удаляет сокращенные URL по их идентификатору
func DeleteAPIUserURLs(service service.URLService) gin.HandlerFunc {
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

		// Получаем ID пользователя из запроса.
		userID, err := getUserIDFromGinContext(c)
		if err != nil {
			_ = c.Error(error.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusUnauthorized,
			})
			return
		}

		// Читаем тело запроса
		var keys []string
		err = json.NewDecoder(c.Request.Body).Decode(&keys)

		if err != nil {
			logger.Log.Error("Error decoding request", zap.Error(err))
			_ = c.Error(err)
			return
		}

		err = service.DeleteManyURLs(c.Request.Context(), userID, keys)

		if err != nil {
			logger.Log.Error("Error deleting URLs", zap.Error(err))
			_ = c.Error(err)
			return
		}

		c.Status(http.StatusAccepted)
	}
}
