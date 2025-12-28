package router

import (
	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/handler"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/coolycow/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

func setupURLRoutes(
	r *gin.Engine,
	cfg *config.Config,
	repo repository.URLRepository,
) {
	srv := service.NewURLService(cfg, repo)
	cookieService := service.NewUserService(cfg, repo)

	r.GET("/ping", handler.PingHandler(srv))

	r.POST("/", middleware.OptionalAuthMiddleware(cookieService), handler.PostHandler(srv))
	r.GET("/:key", middleware.OptionalAuthMiddleware(cookieService), handler.GetHandler(srv))

	r.POST("/api/shorten", middleware.OptionalAuthMiddleware(cookieService), handler.PostAPIShortenHandler(srv))
	r.POST("/api/shorten/batch", middleware.OptionalAuthMiddleware(cookieService), handler.PostAPIShortenBatchHandler(srv))

	r.GET("/api/user/urls", middleware.OptionalAuthMiddleware(cookieService), handler.GetAPIUserURLs(srv))
	r.DELETE("/api/user/urls", middleware.OptionalAuthMiddleware(cookieService), handler.DeleteAPIUserURLs(srv))
}
