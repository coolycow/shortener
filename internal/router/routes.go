package router

import (
	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/handler"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/gin-gonic/gin"
)

func setupURLRoutes(
	r *gin.Engine,
	cfg *config.Config,
	repo repository.URLRepository,
) {
	r.GET("/:key", handler.GetHandler(repo))

	r.POST("/", handler.PostHandler(cfg, repo))
}
