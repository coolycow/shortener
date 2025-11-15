package router

import (
	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, repo repository.URLRepository) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())

	setupURLRoutes(router, cfg, repo)

	return router
}
