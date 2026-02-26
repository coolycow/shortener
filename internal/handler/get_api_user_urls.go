package handler

import (
	"net/http"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

// GetAPIUserURLs возвращает обработчик GET /api/user/urls — список всех сокращённых URL текущего пользователя.
func GetAPIUserURLs(service service.URLService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := middleware.GetUserIDFromGinContext(c)
		if err != nil {
			_ = c.Error(error.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusInternalServerError,
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
