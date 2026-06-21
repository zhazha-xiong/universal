package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const protocolVersion = "2025-06-18"

// NewRouter 创建 MCP 运行时 HTTP 路由。
func NewRouter(service *Service) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	RegisterRoutes(router, service)
	return router
}

type request struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  params `json:"params,omitempty"`
}

type params struct {
	Name      string         `json:"name,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

func handleMCP(ctx *gin.Context, service *Service) {
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
	case "tools/list":
		ctx.JSON(http.StatusOK, gin.H{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result": gin.H{
				"tools": service.ListTools(),
			},
		})
	case "tools/call":
		result, err := service.CallTool(req.Params.Name, req.Params.Arguments)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": gin.H{
					"content": []gin.H{
						{
							"type": "text",
							"text": fmt.Sprintf("{\"code\":40402,\"message\":%q,\"data\":null}", err.Error()),
						},
					},
					"isError": true,
				},
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result": gin.H{
				"content": []gin.H{
					{
						"type": "text",
						"text": fmt.Sprintf("{\"code\":0,\"message\":\"success\",\"data\":{\"output\":%s}}", result),
					},
				},
				"isError": false,
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

// Tool 表示对外暴露的 MCP 工具。
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema,omitempty"`
	Method      string         `json:"-"`
	URL         string         `json:"-"`
}

type catalog interface {
	ListTools() []Tool
	FindTool(name string) (Tool, bool)
}

// Service 负责 MCP 运行时工具查询与调用。
type Service struct {
	catalog catalog
	client  *http.Client
}

// NewService 创建运行时服务。
func NewService(catalog catalog, client *http.Client) *Service {
	if client == nil {
		client = http.DefaultClient
	}
	return &Service{
		catalog: catalog,
		client:  client,
	}
}

// ListTools 返回当前可用工具列表。
func (service *Service) ListTools() []Tool {
	if service == nil || service.catalog == nil {
		return []Tool{}
	}
	return service.catalog.ListTools()
}

// CallTool 调用指定工具并返回上游 JSON 结果。
func (service *Service) CallTool(name string, arguments map[string]any) (string, error) {
	if service == nil || service.catalog == nil {
		return "", fmt.Errorf("tool not found")
	}

	tool, ok := service.catalog.FindTool(name)
	if !ok {
		return "", fmt.Errorf("tool not found")
	}

	body, err := buildUpstreamBody(name, arguments)
	if err != nil {
		return "", err
	}

	method := tool.Method
	if method == "" {
		method = http.MethodPost
	}

	request, err := http.NewRequest(method, tool.URL, strings.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := service.client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func buildUpstreamBody(name string, arguments map[string]any) (string, error) {
	payload := map[string]any{
		"input": arguments,
		"context": map[string]any{
			"tool_name": name,
		},
	}

	content, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
