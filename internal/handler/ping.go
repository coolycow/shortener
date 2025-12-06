package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/error"
	"github.com/gin-gonic/gin"
)

// PingHandler Проверяет возможность соединения с БД
func PingHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)

		if err != nil {
			_ = c.Error(error.CustomError{
				Message:    "Database connection error",
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err = db.PingContext(ctx); err != nil {
			_ = c.Error(error.CustomError{
				Message:    "Database connection failed",
				StatusCode: http.StatusInternalServerError,
			})
			return
		}

		// Формируем ответ
		c.String(http.StatusOK, "Database connection succeeded")
	}
}
