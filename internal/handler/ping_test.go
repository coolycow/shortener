package handler

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPingHandler(t *testing.T) {
	type want struct {
		code int
	}

	tests := []struct {
		name   string
		method string
		dsn    string
		want   want
	}{
		{
			name:   "Valid DSN",
			method: "GET",
			dsn:    "TEST_DATABASE_DSN",
			want: want{
				code: 200,
			},
		},
		{
			name:   "Invalid DSN",
			method: "GET",
			dsn:    "host=localhost port=5432 user=shortener password=shortener dbname=shortener sslmode=disable",
			want: want{
				code: 500,
			},
		},
		{
			name:   "Method not allowed",
			method: "POST",
			dsn:    "AAbbCC2",
			want: want{
				code: 404,
			},
		},
	}

	cfg, _ := config.InitConfig()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.dsn == "TEST_DATABASE_DSN" {
				test.dsn = os.Getenv("TEST_DATABASE_DSN")
			}

			if test.dsn == "" {
				t.Skip("Skipping integration test. TEST_DATABASE_DSN not set")
			}

			cfg.DatabaseDSN = test.dsn

			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(middleware.RequestLogger())
			router.Use(middleware.ErrorHandler())
			router.GET("/ping", PingHandler(cfg))

			request := httptest.NewRequest(test.method, "/ping", nil)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, request)

			result := w.Result()
			if err := result.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, test.want.code, result.StatusCode)
		})
	}
}
