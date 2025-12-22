package router

import (
	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/handler"
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

	r.GET("/:key", handler.GetHandler(srv))

	r.POST("/", handler.PostHandler(srv))

	r.POST("/api/shorten", handler.PostAPIShortenHandler(srv))

	r.POST("/api/shorten/batch", handler.PostAPIShortenBatchHandler(srv))

	r.GET("/ping", handler.PingHandler(srv))
}
