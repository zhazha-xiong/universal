package admin

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
)

type memoryStore struct {
	mu       sync.RWMutex
	services map[string]ServiceItem
	tools    map[string]map[string]ToolItem
}

// NewMemoryStore 创建仅用于本地测试的内存存储。
func NewMemoryStore() Store {
	return &memoryStore{
		services: map[string]ServiceItem{},
		tools:    map[string]map[string]ToolItem{},
	}
}

func (store *memoryStore) UpsertService(serviceID string, payload ServicePayload) (ServiceItem, error) {
	if serviceID == "" || payload.Name == "" || payload.BaseURL == "" {
		return ServiceItem{}, fmt.Errorf("service_id、name、base_url 不能为空")
	}
	if strings.Contains(payload.Name, "/") {
		return ServiceItem{}, fmt.Errorf("service name 不能包含 /")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	for id, item := range store.services {
		if id != serviceID && item.Name == payload.Name {
			return ServiceItem{}, fmt.Errorf("service name 已存在")
		}
	}

	now := time.Now()
	current, ok := store.services[serviceID]
	if !ok {
		current.CreatedAt = now
	}
	current.ServiceID = serviceID
	current.Name = payload.Name
	current.Description = payload.Description
	current.BaseURL = payload.BaseURL
	current.Status = normalizeStatus(payload.Status)
	current.Tags = slices.Clone(payload.Tags)
	current.OwnerID = payload.OwnerID
	current.Ext = cloneMap(payload.Ext)
	current.UpdatedAt = now
	if current.CreatedAt.IsZero() {
		current.CreatedAt = now
	}

	store.services[serviceID] = current
	return current, nil
}

func (store *memoryStore) GetService(serviceID string) (ServiceItem, bool) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	item, ok := store.services[serviceID]
	return item, ok
}

func (store *memoryStore) SearchServices(query ServiceQuery) ([]ServiceItem, int) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	var items []ServiceItem
	for _, item := range store.services {
		if query.ServiceID != "" && item.ServiceID != query.ServiceID {
			continue
		}
		if query.Status != "" && item.Status != query.Status {
			continue
		}
		if query.OwnerID != "" && item.OwnerID != query.OwnerID {
			continue
		}
		if query.Tag != "" && !contains(item.Tags, query.Tag) {
			continue
		}
		if query.Keyword != "" && !strings.Contains(item.Name, query.Keyword) && !strings.Contains(item.Description, query.Keyword) {
			continue
		}
		items = append(items, item)
	}

	slices.SortFunc(items, func(left, right ServiceItem) int {
		return strings.Compare(left.ServiceID, right.ServiceID)
	})
	return paginate(items, query.Page, query.PageSize), len(items)
}

func (store *memoryStore) UpsertTool(serviceID string, toolID string, payload ToolPayload) (ToolItem, error) {
	if serviceID == "" || toolID == "" || payload.Name == "" || payload.Path == "" {
		return ToolItem{}, fmt.Errorf("service_id、tool_id、name、path 不能为空")
	}
	if strings.Contains(payload.Name, "/") {
		return ToolItem{}, fmt.Errorf("tool name 不能包含 /")
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.services[serviceID]; !ok {
		return ToolItem{}, fmt.Errorf("service not found")
	}

	if _, ok := store.tools[serviceID]; !ok {
		store.tools[serviceID] = map[string]ToolItem{}
	}
	for id, item := range store.tools[serviceID] {
		if id != toolID && item.Name == payload.Name {
			return ToolItem{}, fmt.Errorf("tool name 已存在")
		}
		if id != toolID && item.Path == payload.Path {
			return ToolItem{}, fmt.Errorf("tool path 已存在")
		}
	}

	now := time.Now()
	current, ok := store.tools[serviceID][toolID]
	if !ok {
		current.CreatedAt = now
	}
	current.ToolID = toolID
	current.ServiceID = serviceID
	current.Name = payload.Name
	current.Description = payload.Description
	current.Path = payload.Path
	current.Method = normalizeMethod(payload.Method)
	current.InputSchema = cloneMap(payload.InputSchema)
	current.OutputSchema = cloneMap(payload.OutputSchema)
	current.Status = normalizeStatus(payload.Status)
	current.Tags = slices.Clone(payload.Tags)
	current.Ext = cloneMap(payload.Ext)
	current.UpdatedAt = now
	if current.CreatedAt.IsZero() {
		current.CreatedAt = now
	}

	store.tools[serviceID][toolID] = current
	return current, nil
}

func (store *memoryStore) ListTools(serviceID string, query ToolQuery) ([]ToolItem, int) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	toolsByService := store.tools[serviceID]
	var items []ToolItem
	for _, item := range toolsByService {
		if query.ToolID != "" && item.ToolID != query.ToolID {
			continue
		}
		if query.Status != "" && item.Status != query.Status {
			continue
		}
		if query.Tag != "" && !contains(item.Tags, query.Tag) {
			continue
		}
		if query.Keyword != "" && !strings.Contains(item.Name, query.Keyword) && !strings.Contains(item.Description, query.Keyword) {
			continue
		}
		items = append(items, item)
	}

	slices.SortFunc(items, func(left, right ToolItem) int {
		return strings.Compare(left.ToolID, right.ToolID)
	})
	return paginate(items, query.Page, query.PageSize), len(items)
}

func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	content, err := json.Marshal(input)
	if err != nil {
		return nil
	}
	var output map[string]any
	if err := json.Unmarshal(content, &output); err != nil {
		return nil
	}
	return output
}

func paginate[T any](items []T, page int, pageSize int) []T {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	start := (page - 1) * pageSize
	if start >= len(items) {
		return []T{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func normalizeStatus(status string) string {
	if status == "" {
		return "active"
	}
	return status
}

func normalizeMethod(method string) string {
	if method == "" {
		return "POST"
	}
	return strings.ToUpper(method)
}
