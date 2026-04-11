package handler

import (
	"net"
	"net/http"
	"strings"

	"github.com/coolycow/shortener/internal/model"
	"github.com/coolycow/shortener/internal/repository"
	"github.com/gin-gonic/gin"
)

// GetAPIInternalStatsHandler возвращает обработчик GET /api/internal/stats.
// Доступ разрешён только если заголовок X-Real-IP содержит IP из доверенной подсети (trusted).
// При trusted == nil (пустая trusted_subnet в конфиге) отвечает 403 для любого запроса.
func GetAPIInternalStatsHandler(repo repository.URLRepository, trusted *net.IPNet) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Если trusted == nil (пустая trusted_subnet в конфиге), отвечаем 403 для любого запроса
		if trusted == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Получаем IP из заголовка X-Real-IP
		clientIP := strings.TrimSpace(c.GetHeader("X-Real-IP"))
		ip := net.ParseIP(clientIP)

		// Если IP невалидный или не в доверенной подсети, отвечаем 403
		if ip == nil || !trusted.Contains(ip) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Получаем контекст запроса
		ctx := c.Request.Context()

		// Формируем ответ
		c.JSON(http.StatusOK, model.APIInternalStats{
			Urls:  repo.GetSize(ctx),
			Users: repo.GetUsersCount(ctx),
		})
	}
}
