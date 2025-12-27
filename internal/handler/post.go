package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	errors2 "github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

func getUserIDFromGinContext(c *gin.Context) (int, error) {
	value, exists := c.Get("userID")

	if !exists || value == nil {
		return 0, errors.New("user ID not found in context")
	}

	userID, ok := value.(int)
	fmt.Println(value)
	if !ok {
		return 0, errors.New("incorrect user ID in context")
	}

	return userID, nil
}

// PostHandler обрабатывает POST-запросы к серверу
func PostHandler(service service.URLService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, что тип контента - text/plain
		// Учитываем, что "Content-Type" может содержать и другие значения, например, charset=utf-8
		contentType := c.GetHeader("Content-Type")
		if !strings.HasPrefix(contentType, "text/plain") {
			_ = c.Error(errors2.CustomError{
				Message:    "Content type not allowed",
				StatusCode: http.StatusUnsupportedMediaType,
			})
			return
		}

		// Читаем тело запроса
		body, err := io.ReadAll(c.Request.Body)

		if err != nil {
			_ = c.Error(errors2.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Проверяем, что не пришла пустота
		if len(body) == 0 {
			_ = c.Error(errors2.CustomError{
				Message:    "Empty body",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Извлекаем строку из тела запроса и проверяем, что она не пуста
		trimBody := strings.TrimSpace(string(body))

		if trimBody == "" {
			_ = c.Error(errors2.CustomError{
				Message:    "Empty URL",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		// Парсим URL из строки (фактически проверяем, что это действительно URL)
		validURL, err := url.ParseRequestURI(trimBody)

		if err != nil {
			_ = c.Error(errors2.CustomError{
				Message:    "Invalid URL",
				StatusCode: http.StatusBadRequest,
			})
			return
		}

		userID, err := getUserIDFromGinContext(c)
		if err != nil {
			_ = c.Error(errors2.CustomError{
				Message:    err.Error(),
				StatusCode: http.StatusUnauthorized,
			})
			return
		}

		shortURL, err := service.CreateShortURL(c.Request.Context(), userID, validURL.String())

		if err != nil {
			_ = c.Error(err)
			return
		}

		// Формируем ответ
		c.String(http.StatusCreated, shortURL)
	}
}
