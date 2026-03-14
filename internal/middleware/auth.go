// Package middleware содержит HTTP-middleware: аутентификация, gzip, логирование ошибок.
package middleware

import (
	"context"
	"errors"
	"net/http"

	httpError "github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

type ginKey string

const (
	UserIDKey ginKey = "userID"
)

func GetUserIDFromGinContext(c *gin.Context) (string, error) {
	value, exists := c.Get(string(UserIDKey))

	if !exists || value == nil {
		return "", errors.New("user ID not found in context")
	}

	userID, ok := value.(string)

	if !ok {
		return "", errors.New("incorrect user ID in context")
	}

	return userID, nil
}

func createUserAndCookieValue(ctx context.Context, userService service.UserService) (model.User, string, error) {
	user, err := userService.CreateUser(ctx)
	if err != nil {
		return model.User{}, "", err
	}

	cookieValue, err := userService.GetCookieValueByUser(user)

	if err != nil {
		return model.User{}, "", err
	}

	return user, cookieValue, nil
}

func OptionalAuthMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("auth")

		if err != nil {
			user, cookieValue, err := createUserAndCookieValue(context.Background(), userService)

			if err != nil {
				_ = c.Error(httpError.CustomError{
					Message:    err.Error(),
					StatusCode: http.StatusInternalServerError,
				})
				c.Abort()
				return
			}

			http.SetCookie(c.Writer, &http.Cookie{
				Name:     "auth",
				Value:    cookieValue,
				Path:     "/",
				HttpOnly: true,
			})

			c.Set(string(UserIDKey), user.ID)
		} else {
			// Проверяем валидность куки
			userID, err := userService.GetUserIDFromCookie(cookie)

			if err != nil {
				user, cookieValue, createErr := createUserAndCookieValue(context.Background(), userService)

				if createErr != nil {
					_ = c.Error(httpError.CustomError{
						Message:    createErr.Error(),
						StatusCode: http.StatusInternalServerError,
					})
					c.Abort()
					return
				}

				http.SetCookie(c.Writer, &http.Cookie{
					Name:     "auth",
					Value:    cookieValue,
					Path:     "/",
					HttpOnly: true,
				})

				userID = user.ID
			}

			_, err = userService.GetUser(c.Request.Context(), userID)
			if err != nil {
				_ = c.Error(httpError.CustomError{
					Message:    "User with this ID does not exist",
					StatusCode: http.StatusUnauthorized,
				})
				c.Abort()
				return
			}

			c.Set(string(UserIDKey), userID)
		}

		c.Next()
	}
}

func RequiredAuthMiddleware(userService service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Request.Cookie("auth")

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userID, err := userService.GetUserIDFromCookie(cookie)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid cookie"})
			c.Abort()
			return
		}

		_, err = userService.GetUser(c.Request.Context(), userID)
		if err != nil {
			_ = c.Error(httpError.CustomError{
				Message:    "User with this ID does not exist",
				StatusCode: http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		c.Set(string(UserIDKey), userID)

		c.Next()
	}
}
