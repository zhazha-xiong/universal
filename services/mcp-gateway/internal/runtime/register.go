package runtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册 MCP 运行时路由。
func RegisterRoutes(router gin.IRoutes, service *Service) {
	router.POST("/mcp", func(ctx *gin.Context) {
		handleMCP(ctx, service)
	})
	router.GET("/mcp", func(ctx *gin.Context) {
		ctx.Status(http.StatusMethodNotAllowed)
	})
}
