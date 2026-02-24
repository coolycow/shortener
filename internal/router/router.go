package router

import (
	"github.com/coolycow/shortener/internal/config"
	"github.com/coolycow/shortener/internal/middleware"
	"github.com/coolycow/shortener/internal/observer/audit"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
)

func NewRouter(cfg *config.Config, repo repository.URLRepository, auditNotifier *audit.Notifier) *gin.Engine {
	router := gin.Default()

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequestGzip())

	setupURLRoutes(router, cfg, repo, auditNotifier)

	return router
}
