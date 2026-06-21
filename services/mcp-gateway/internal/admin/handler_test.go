package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutesUpsertsService(t *testing.T) {
	service := NewService()
	router := newRouter(service)
	body := `{"name":"search","description":"搜索服务","base_url":"https://search.example.com","status":"active","tags":["search"],"owner_id":"bear"}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/mcp/services/search-service", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"service_id":"search-service"`) {
		t.Fatalf("response missing service id: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":0`) {
		t.Fatalf("response missing success code: %s", rec.Body.String())
	}
}

func TestRegisterRoutesGetsServiceDetail(t *testing.T) {
	service := NewService()
	_, err := service.UpsertService("search-service", ServicePayload{
		Name:        "search",
		Description: "搜索服务",
		BaseURL:     "https://search.example.com",
		Status:      "active",
	})
	if err != nil {
		t.Fatalf("upsert service: %v", err)
	}

	router := newRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/mcp/services/search-service", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"name":"search"`) {
		t.Fatalf("response missing service name: %s", rec.Body.String())
	}
}

func TestRegisterRoutesSearchesServices(t *testing.T) {
	service := NewService()
	_, _ = service.UpsertService("search-service", ServicePayload{
		Name:        "search",
		Description: "搜索服务",
		BaseURL:     "https://search.example.com",
		Status:      "active",
	})
	_, _ = service.UpsertService("ticket-service", ServicePayload{
		Name:        "ticket",
		Description: "工单服务",
		BaseURL:     "https://ticket.example.com",
		Status:      "inactive",
	})

	router := newRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/mcp/services?status=active&keyword=搜索", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"total":1`) {
		t.Fatalf("response missing total: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"service_id":"search-service"`) {
		t.Fatalf("response missing matched service: %s", rec.Body.String())
	}
}

func TestRegisterRoutesUpsertsTool(t *testing.T) {
	service := NewService()
	_, _ = service.UpsertService("search-service", ServicePayload{
		Name:        "search",
		Description: "搜索服务",
		BaseURL:     "https://search.example.com",
		Status:      "active",
	})

	router := newRouter(service)
	body := `{"name":"web_search","description":"搜索网页","path":"/tools/web-search","method":"POST","input_schema":{"type":"object"},"output_schema":{"type":"object"},"status":"active","tags":["search"]}`
	req := httptest.NewRequest(http.MethodPut, "/api/admin/mcp/services/search-service/tools/search-web-search", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"tool_id":"search-web-search"`) {
		t.Fatalf("response missing tool id: %s", rec.Body.String())
	}
}

func TestRegisterRoutesListsToolsByService(t *testing.T) {
	service := NewService()
	_, _ = service.UpsertService("search-service", ServicePayload{
		Name:        "search",
		Description: "搜索服务",
		BaseURL:     "https://search.example.com",
		Status:      "active",
	})
	_, _ = service.UpsertTool("search-service", "search-web-search", ToolPayload{
		Name:        "web_search",
		Description: "搜索网页",
		Path:        "/tools/web-search",
		Method:      "POST",
		Status:      "active",
	})

	router := newRouter(service)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/mcp/services/search-service/tools?status=active&keyword=网页", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"total":1`) {
		t.Fatalf("response missing total: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"tool_id":"search-web-search"`) {
		t.Fatalf("response missing tool id: %s", rec.Body.String())
	}
}

func newRouter(service *Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	RegisterRoutes(router, service)
	return router
}
