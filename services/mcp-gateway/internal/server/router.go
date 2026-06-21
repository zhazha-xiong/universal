package server

import (
	"github.com/gin-gonic/gin"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/admin"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/runtime"
)

// NewRouter 创建并组装网关顶层 HTTP 路由。
func NewRouter(adminService *admin.Service, runtimeService *runtime.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	admin.RegisterRoutes(router, adminService)
	runtime.RegisterRoutes(router, runtimeService)
	return router
}
