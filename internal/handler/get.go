package handler

import (
	"net/http"
	"strings"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/gin-gonic/gin"
)

// GetHandler Обрабатываем GET-запросы к серверу.
func GetHandler(repo repository.URLRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем id из URL
		key := strings.TrimSpace(strings.TrimPrefix(c.Param("key"), "/"))

		// Проверяем, что id не пустой
		if key == "" {
			_ = c.Error(error.CustomError{
				Message:    "Key is required",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Получаем исходный URL по id из репозитория
		rawURL, exists := repo.GetOriginalURL(key)

		// Проверяем, что URL существует
		if !exists {
			_ = c.Error(error.CustomError{
				Message:    "URL not found",
				StatusCode: http.StatusNotFound,
			})
			return
		}

		// Формируем ответ
		c.Header("Location", rawURL)
		c.Status(http.StatusTemporaryRedirect)
	}
}
