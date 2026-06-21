package repository

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/admin"
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
