package repository

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/admin"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/runtime"
)

func TestRepositoryUpsertsAndGetsService(t *testing.T) {
	repository := newTestRepository(t)

	item, err := repository.UpsertService("search-service", admin.ServicePayload{
		Name:        "search",
		Description: "搜索服务",
		BaseURL:     "https://search.example.com",
		Status:      "active",
		Tags:        []string{"search"},
		OwnerID:     "bear",
	})
	if err != nil {
		t.Fatalf("upsert service: %v", err)
	}
	if item.ServiceID != "search-service" {
		t.Fatalf("service_id = %q, want %q", item.ServiceID, "search-service")
	}

	got, ok := repository.GetService("search-service")
	if !ok {
		t.Fatal("get service = not found, want found")
	}
	if got.Name != "search" {
		t.Fatalf("name = %q, want %q", got.Name, "search")
	}
}

func TestRepositorySearchesServices(t *testing.T) {
	repository := newTestRepository(t)
	_, _ = repository.UpsertService("search-service", admin.ServicePayload{
		Name:        "search",
		Description: "搜索服务",
		BaseURL:     "https://search.example.com",
		Status:      "active",
	})
	_, _ = repository.UpsertService("ticket-service", admin.ServicePayload{
		Name:        "ticket",
		Description: "工单服务",
		BaseURL:     "https://ticket.example.com",
		Status:      "inactive",
	})

	items, total := repository.SearchServices(admin.ServiceQuery{
		Status:   "active",
		Keyword:  "搜索",
		Page:     1,
		PageSize: 20,
	})
	if total != 1 {
		t.Fatalf("total = %d, want %d", total, 1)
	}
	if len(items) != 1 || items[0].ServiceID != "search-service" {
		t.Fatalf("items = %+v", items)
	}
}

func TestRepositoryUpsertsAndListsTools(t *testing.T) {
	repository := newTestRepository(t)
	_, _ = repository.UpsertService("search-service", admin.ServicePayload{
		Name:        "search",
		Description: "搜索服务",
		BaseURL:     "https://search.example.com",
		Status:      "active",
	})

	item, err := repository.UpsertTool("search-service", "search-web-search", admin.ToolPayload{
		Name:        "web_search",
		Description: "搜索网页",
		Path:        "/tools/web-search",
		Method:      "POST",
		Status:      "active",
		Tags:        []string{"search"},
	})
	if err != nil {
		t.Fatalf("upsert tool: %v", err)
	}
	if item.ToolID != "search-web-search" {
		t.Fatalf("tool_id = %q, want %q", item.ToolID, "search-web-search")
	}

	items, total := repository.ListTools("search-service", admin.ToolQuery{
		Status:   "active",
		Keyword:  "网页",
		Page:     1,
		PageSize: 20,
	})
	if total != 1 {
		t.Fatalf("total = %d, want %d", total, 1)
	}
	if len(items) != 1 || items[0].ToolID != "search-web-search" {
		t.Fatalf("items = %+v", items)
	}
}

func TestRepositoryListsRuntimeTools(t *testing.T) {
	repository := newTestRepository(t)
	_, _ = repository.UpsertService("search-service", admin.ServicePayload{
		Name:        "search",
		Description: "搜索服务",
		BaseURL:     "https://search.example.com/api",
		Status:      "active",
	})
	_, _ = repository.UpsertTool("search-service", "search-web-search", admin.ToolPayload{
		Name:        "web_search",
		Description: "搜索网页",
		Path:        "/tools/web-search",
		Method:      "POST",
		Status:      "active",
		InputSchema: map[string]any{
			"type": "object",
		},
	})

	tools := repository.ListRuntimeTools()
	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d, want %d", len(tools), 1)
	}
	if tools[0].Name != "search/web_search" {
		t.Fatalf("name = %q, want %q", tools[0].Name, "search/web_search")
	}
	if tools[0].URL != "https://search.example.com/api/tools/web-search" {
		t.Fatalf("url = %q", tools[0].URL)
	}
}

func TestRepositoryFindsRuntimeToolOnlyWhenActive(t *testing.T) {
	repository := newTestRepository(t)
	_, _ = repository.UpsertService("search-service", admin.ServicePayload{
		Name:        "search",
		Description: "搜索服务",
		BaseURL:     "https://search.example.com",
		Status:      "inactive",
	})
	_, _ = repository.UpsertTool("search-service", "search-web-search", admin.ToolPayload{
		Name:        "web_search",
		Description: "搜索网页",
		Path:        "/tools/web-search",
		Method:      "POST",
		Status:      "active",
	})
	_, _ = repository.UpsertService("ticket-service", admin.ServicePayload{
		Name:        "ticket",
		Description: "工单服务",
		BaseURL:     "https://ticket.example.com",
		Status:      "active",
	})
	_, _ = repository.UpsertTool("ticket-service", "ticket-query", admin.ToolPayload{
		Name:        "query",
		Description: "查询工单",
		Path:        "/query",
		Method:      "POST",
		Status:      "inactive",
	})
	_, _ = repository.UpsertService("knowledge-service", admin.ServicePayload{
		Name:        "knowledge",
		Description: "知识库服务",
		BaseURL:     "https://knowledge.example.com/api",
		Status:      "active",
	})
	_, _ = repository.UpsertTool("knowledge-service", "knowledge-recall", admin.ToolPayload{
		Name:        "recall",
		Description: "召回知识",
		Path:        "/recall",
		Method:      "POST",
		Status:      "active",
	})

	if _, ok := repository.FindRuntimeTool("search/web_search"); ok {
		t.Fatal("inactive service tool = found, want not found")
	}
	if _, ok := repository.FindRuntimeTool("ticket/query"); ok {
		t.Fatal("inactive tool = found, want not found")
	}

	got, ok := repository.FindRuntimeTool("knowledge/recall")
	if !ok {
		t.Fatal("active tool = not found, want found")
	}
	if got.URL != "https://knowledge.example.com/api/recall" {
		t.Fatalf("url = %q", got.URL)
	}
}

var _ interface {
	ListRuntimeTools() []runtime.Tool
	FindRuntimeTool(name string) (runtime.Tool, bool)
} = (*Repository)(nil)

func newTestRepository(t *testing.T) *Repository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&serviceRecord{}, &toolRecord{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return NewRepository(db)
}
