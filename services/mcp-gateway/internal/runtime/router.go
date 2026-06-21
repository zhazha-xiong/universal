package runtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const protocolVersion = "2025-06-18"

// NewRouter creates the MCP runtime HTTP router.
func NewRouter() http.Handler {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.POST("/mcp", handleMCP)
	router.GET("/mcp", func(ctx *gin.Context) {
		ctx.Status(http.StatusMethodNotAllowed)
	})

	return router
}

type request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method"`
}

func handleMCP(ctx *gin.Context) {
	var req request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"jsonrpc": "2.0",
			"error": gin.H{
				"code":    -32700,
				"message": "parse error",
			},
			"id": nil,
		})
		return
	}

	switch req.Method {
	case "initialize":
		ctx.JSON(http.StatusOK, gin.H{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result": gin.H{
				"protocolVersion": protocolVersion,
				"serverInfo": gin.H{
					"name":    "universal-mcp-gateway",
					"version": "0.1.0",
				},
				"capabilities": gin.H{
					"tools": gin.H{
						"listChanged": true,
					},
				},
			},
		})
	case "notifications/initialized":
		ctx.Status(http.StatusAccepted)
	default:
		ctx.JSON(http.StatusOK, gin.H{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"error": gin.H{
				"code":    -32601,
				"message": "method not found",
			},
		})
	}
}
