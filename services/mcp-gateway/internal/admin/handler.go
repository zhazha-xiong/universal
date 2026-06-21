package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册管理侧 REST API 路由。
func RegisterRoutes(router gin.IRoutes, service *Service) {
	router.PUT("/api/admin/mcp/services/:service_id", func(ctx *gin.Context) {
		var payload ServicePayload
		if err := ctx.ShouldBindJSON(&payload); err != nil {
			writeError(ctx, http.StatusBadRequest, 40001, "请求参数错误")
			return
		}

		item, err := service.UpsertService(ctx.Param("service_id"), payload)
		if err != nil {
			writeError(ctx, http.StatusBadRequest, 40001, err.Error())
			return
		}
		writeSuccess(ctx, item)
	})

	router.GET("/api/admin/mcp/services/:service_id", func(ctx *gin.Context) {
		item, ok := service.GetService(ctx.Param("service_id"))
		if !ok {
			writeError(ctx, http.StatusNotFound, 40401, "service not found")
			return
		}
		writeSuccess(ctx, item)
	})

	router.GET("/api/admin/mcp/services", func(ctx *gin.Context) {
		query := ServiceQuery{
			ServiceID: ctx.Query("service_id"),
			Keyword:   ctx.Query("keyword"),
			Status:    ctx.Query("status"),
			OwnerID:   ctx.Query("owner_id"),
			Tag:       ctx.Query("tag"),
			Page:      parsePositiveInt(ctx.Query("page"), 1),
			PageSize:  parsePositiveInt(ctx.Query("page_size"), 20),
		}
		items, total := service.SearchServices(query)
		writeSuccess(ctx, gin.H{
			"items":     items,
			"page":      query.Page,
			"page_size": query.PageSize,
			"total":     total,
		})
	})

	router.PUT("/api/admin/mcp/services/:service_id/tools/:tool_id", func(ctx *gin.Context) {
		var payload ToolPayload
		if err := ctx.ShouldBindJSON(&payload); err != nil {
			writeError(ctx, http.StatusBadRequest, 40001, "请求参数错误")
			return
		}

		item, err := service.UpsertTool(ctx.Param("service_id"), ctx.Param("tool_id"), payload)
		if err != nil {
			writeError(ctx, http.StatusBadRequest, 40001, err.Error())
			return
		}
		writeSuccess(ctx, item)
	})

	router.GET("/api/admin/mcp/services/:service_id/tools", func(ctx *gin.Context) {
		query := ToolQuery{
			ToolID:   ctx.Query("tool_id"),
			Keyword:  ctx.Query("keyword"),
			Status:   ctx.Query("status"),
			Tag:      ctx.Query("tag"),
			Page:     parsePositiveInt(ctx.Query("page"), 1),
			PageSize: parsePositiveInt(ctx.Query("page_size"), 20),
		}
		items, total := service.ListTools(ctx.Param("service_id"), query)
		writeSuccess(ctx, gin.H{
			"items":     items,
			"page":      query.Page,
			"page_size": query.PageSize,
			"total":     total,
		})
	})
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func writeSuccess(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

func writeError(ctx *gin.Context, status int, code int, message string) {
	ctx.JSON(status, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
