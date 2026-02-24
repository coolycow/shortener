package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BenchmarkPostHandler измеряет скорость обработки POST / (создание короткой ссылки).
func BenchmarkPostHandler(b *testing.B) {
	cfg, _ := config.InitConfig()
	tmpFile := "bench_" + uuid.New().String() + ".json"
	defer os.Remove(tmpFile)

	repo := repository.NewDoubleMapsRepository(tmpFile)
	defer repo.Close()

	auditNotifier := audit.NewNotifier("", "")
	urlService := service.NewURLService(cfg, repo)
	userService := service.NewUserService(cfg, repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.OptionalAuthMiddleware(userService))
	router.POST("/", PostHandler(urlService, auditNotifier))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body := fmt.Sprintf("https://example.com/%d", i)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkGetHandler измеряет скорость обработки GET /:key (редирект по короткой ссылке).
func BenchmarkGetHandler(b *testing.B) {
	cfg, _ := config.InitConfig()
	tmpFile := "bench_" + uuid.New().String() + ".json"
	defer os.Remove(tmpFile)

	repo := repository.NewDoubleMapsRepository(tmpFile)
	defer repo.Close()

	_, _, _ = repo.SaveURL(context.Background(), "user1", "https://example.com/unique", "Abc123")

	auditNotifier := audit.NewNotifier("", "")
	urlService := service.NewURLService(cfg, repo)
	userService := service.NewUserService(cfg, repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.OptionalAuthMiddleware(userService))
	router.GET("/:key", GetHandler(urlService, auditNotifier))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/Abc123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
