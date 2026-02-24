package handler

import (
	"net/http"
	"strings"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

// GetHandler Обрабатываем GET-запросы к серверу.
func GetHandler(service service.URLService, auditNotifier *audit.Notifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем ключ из URL
		key := strings.TrimSpace(strings.TrimPrefix(c.Param("key"), "/"))

		// Получаем идентификатор пользователя из контекста
		userID, err := middleware.GetUserIDFromGinContext(c)
		if err != nil {
			_ = c.Error(error.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		// Получаем исходный URL по ключу из сервиса
		shortURL, err := service.GetShortURL(c.Request.Context(), key)

		// Сервис возвращает CustomError
		if err != nil {
			_ = c.Error(err)
			return
		}

		if shortURL.DeletedAt != nil {
			c.Status(http.StatusGone)
			return
		}

		// Формируем ответ
		c.Header("Location", shortURL.OriginalURL)
		c.Status(http.StatusTemporaryRedirect)

		// Отправляем событие аудита
		event := audit.NewEvent(audit.ActionFollow, userID, shortURL.OriginalURL)
		auditNotifier.Notify(event)
	}
}
