package middleware

import (
	"net/http"

	"github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

func OptionalAuthMiddleware(cookieService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("user")

		if err != nil {
			user, err := cookieService.CreateUser(c.Request.Context())

			if err != nil {
				_ = c.Error(error.CustomError{
					Message:    err.Error(),
					StatusCode: http.StatusInternalServerError,
				})
				c.Abort()
				return
			}

			cookieValue, err := cookieService.GetCookieValueByUserID(user.ID)

			if err != nil {
				_ = c.Error(error.CustomError{
					Message:    err.Error(),
					StatusCode: http.StatusInternalServerError,
				})
				c.Abort()
				return
			}

			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "user",
				Value:    cookieValue,
				Path:     "/",
				HttpOnly: true,
			})

			c.Set("userID", user.ID)
		} else {
			// Проверяем валидность куки
			userID, err := cookieService.GetUserIDFromCookie(cookie)

			if err != nil {
				// Создаем нового пользователя, если кука невалидна
				user, err := cookieService.CreateUser(c.Request.Context())

				if err != nil {
					_ = c.Error(error.CustomError{
						Message:    err.Error(),
						StatusCode: http.StatusInternalServerError,
					})
					c.Abort()
					return
				}

				cookieValue, err := cookieService.GetCookieValueByUserID(user.ID)

				if err != nil {
					_ = c.Error(error.CustomError{
						Message:    err.Error(),
						StatusCode: http.StatusInternalServerError,
					})
					c.Abort()
					return
				}

				http.SetCookie(c.Writer, &http.Cookie{
					Name:     "user",
					Value:    cookieValue,
					Path:     "/",
					HttpOnly: true,
				})

				userID = user.ID
			}

			c.Set("userID", userID)
		}

		c.Next()
	}
}

func RequiredAuthMiddleware(cookieService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("user")

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userID, err := cookieService.GetUserIDFromCookie(cookie)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid cookie"})
			c.Abort()
			return
		}

		c.Set("userID", userID)

		c.Next()
	}
}
