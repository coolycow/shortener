package handler

import (
	"net/http"
	"strings"

	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

// GetHandler Обрабатываем GET-запросы к серверу.
func GetHandler(service service.URLService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем ключ из URL
		key := strings.TrimSpace(strings.TrimPrefix(c.Param("key"), "/"))

		// Получаем исходный URL по ключу из сервиса
		rawURL, err := service.GetOriginalURL(key)

		// Сервис возвращает CustomError
		if err != nil {
			_ = c.Error(err)
			return
		}

		// Формируем ответ
		c.Header("Location", rawURL)
		c.Status(http.StatusTemporaryRedirect)
	}
}
