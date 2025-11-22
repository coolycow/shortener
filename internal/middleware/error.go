package middleware

import (
	"errors"
	"strings"

	"github.com/coolycow/shortener/internal/error"
	"github.com/gin-gonic/gin"
)

// ErrorHandler captures error and returns a consistent JSON error response
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Process other handlers

		if len(c.Errors) > 0 {
			// Iterate through collected error
			for _, ginErr := range c.Errors {
				var customErr error.CustomError
				if errors.As(ginErr.Err, &customErr) {
					if strings.HasPrefix(c.Request.URL.Path, "/api/") {
						c.JSON(customErr.StatusCode, gin.H{"error": customErr.Message})
					} else {
						c.String(customErr.StatusCode, customErr.Message)
					}
					return
				}
			}
		}
	}
}
