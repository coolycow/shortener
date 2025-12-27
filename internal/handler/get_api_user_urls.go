package handler

import (
	"net/http"

	errors2 "github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

// GetAPIUserURLs Обрабатываем GET-запросы к серверу.
func GetAPIUserURLs(service service.URLService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getUserIDFromGinContext(c)
		if err != nil {
			_ = c.Error(errors2.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusUnauthorized,
			})
			return
		}

		// Получаем исходный URL по ключу из сервиса
		shortURLs, err := service.GetManyShortURLs(c.Request.Context(), userID)

		// Сервис возвращает CustomError
		if err != nil {
			_ = c.Error(err)
			return
		}

		if len(shortURLs) == 0 {
			c.JSON(http.StatusNoContent, nil)
			return
		}

		responses := make([]model.APIUserURLsResponse, 0, len(shortURLs))

		for _, u := range shortURLs {
			responses = append(responses, u.ToAPIUserURLsResponse(service.GetBaseURL()))
		}

		// Формируем ответ
		c.JSON(http.StatusOK, responses)
	}
}
