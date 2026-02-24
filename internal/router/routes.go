package router

import (
	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/handler"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

func setupURLRoutes(
	r *gin.Engine,
	cfg *config.Config,
	repo repository.URLRepository,
	auditNotifier *audit.Notifier,
) {
	srv := service.NewURLService(cfg, repo)
	cookieService := service.NewUserService(cfg, repo)

	r.GET("/ping", handler.PingHandler(srv))

	// Группа маршрутов с опциональной аутентификацией
	authGroup := r.Group("/")
	authGroup.Use(middleware.OptionalAuthMiddleware(cookieService))

	authGroup.POST("/", handler.PostHandler(srv, auditNotifier))
	authGroup.GET("/:key", handler.GetHandler(srv, auditNotifier))

	authGroup.POST("/api/shorten", handler.PostAPIShortenHandler(srv, auditNotifier))
	authGroup.POST("/api/shorten/batch", handler.PostAPIShortenBatchHandler(srv))

	authGroup.GET("/api/user/urls", handler.GetAPIUserURLs(srv))
	authGroup.DELETE("/api/user/urls", handler.DeleteAPIUserURLs(srv))
}
