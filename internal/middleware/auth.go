package middleware

import (
	"context"
	"net/http"

	httpError "github.com/coolycow/shortener/internal/error"
	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

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

			c.Set("userID", user.ID)
		} else {
			// Проверяем валидность куки
			userID, err := userService.GetUserIDFromCookie(cookie)

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

			c.Set("userID", userID)
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

		c.Set("userID", userID)

		c.Next()
	}
}
