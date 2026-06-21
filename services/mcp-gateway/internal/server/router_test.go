package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/admin"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/runtime"
)

func TestNewRouterRegistersAdminAndRuntimeRoutes(t *testing.T) {
	router := NewRouter(admin.NewService(), runtime.NewService(nil, nil))

	adminReq := httptest.NewRequest(http.MethodGet, "/api/admin/mcp/services", nil)
	adminRec := httptest.NewRecorder()
	router.ServeHTTP(adminRec, adminReq)
	if adminRec.Code != http.StatusOK {
		t.Fatalf("admin status = %d, want %d", adminRec.Code, http.StatusOK)
	}

	runtimeReq := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18"}}`))
	runtimeReq.Header.Set("Content-Type", "application/json")
	runtimeRec := httptest.NewRecorder()
	router.ServeHTTP(runtimeRec, runtimeReq)
	if runtimeRec.Code != http.StatusOK {
		t.Fatalf("runtime status = %d, want %d", runtimeRec.Code, http.StatusOK)
	}
}
