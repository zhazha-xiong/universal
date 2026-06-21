package admin

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"
)

// Service 管理管理侧 service 和 tool 的内存数据。
type Service struct {
	mu       sync.RWMutex
	services map[string]ServiceItem
	tools    map[string]map[string]ToolItem
}

// ServicePayload 描述 service 的写入参数。
type ServicePayload struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	BaseURL     string         `json:"base_url"`
	Status      string         `json:"status"`
	Tags        []string       `json:"tags"`
	OwnerID     string         `json:"owner_id"`
	Ext         map[string]any `json:"ext"`
}

// ServiceItem 描述管理侧 service 对象。
type ServiceItem struct {
	ServiceID   string         `json:"service_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	BaseURL     string         `json:"base_url"`
	Status      string         `json:"status"`
	Tags        []string       `json:"tags,omitempty"`
	OwnerID     string         `json:"owner_id,omitempty"`
	Ext         map[string]any `json:"ext,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ServiceQuery 描述 service 搜索条件。
type ServiceQuery struct {
	ServiceID string
	Keyword   string
	Status    string
	OwnerID   string
	Tag       string
	Page      int
	PageSize  int
}

// ToolPayload 描述 tool 的写入参数。
type ToolPayload struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Path         string         `json:"path"`
	Method       string         `json:"method"`
	InputSchema  map[string]any `json:"input_schema"`
	OutputSchema map[string]any `json:"output_schema"`
	Status       string         `json:"status"`
	Tags         []string       `json:"tags"`
	Ext          map[string]any `json:"ext"`
}

// ToolItem 描述管理侧 tool 对象。
type ToolItem struct {
	ToolID       string         `json:"tool_id"`
	ServiceID    string         `json:"service_id"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Path         string         `json:"path"`
	Method       string         `json:"method"`
	InputSchema  map[string]any `json:"input_schema,omitempty"`
	OutputSchema map[string]any `json:"output_schema,omitempty"`
	Status       string         `json:"status"`
	Tags         []string       `json:"tags,omitempty"`
	Ext          map[string]any `json:"ext,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// ToolQuery 描述 tool 查询条件。
type ToolQuery struct {
	ToolID   string
	Keyword  string
	Status   string
	Tag      string
	Page     int
	PageSize int
}

// NewService 创建管理侧内存服务。
func NewService() *Service {
	return &Service{
		services: map[string]ServiceItem{},
		tools:    map[string]map[string]ToolItem{},
	}
}

// UpsertService 创建或更新 service。
func (service *Service) UpsertService(serviceID string, payload ServicePayload) (ServiceItem, error) {
	if serviceID == "" || payload.Name == "" || payload.BaseURL == "" {
		return ServiceItem{}, fmt.Errorf("service_id、name、base_url 不能为空")
	}
	if strings.Contains(payload.Name, "/") {
		return ServiceItem{}, fmt.Errorf("service name 不能包含 /")
	}

	service.mu.Lock()
	defer service.mu.Unlock()

	for id, item := range service.services {
		if id != serviceID && item.Name == payload.Name {
			return ServiceItem{}, fmt.Errorf("service name 已存在")
		}
	}

	now := time.Now()
	current, ok := service.services[serviceID]
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

	service.services[serviceID] = current
	return current, nil
}

// GetService 查询单个 service。
func (service *Service) GetService(serviceID string) (ServiceItem, bool) {
	service.mu.RLock()
	defer service.mu.RUnlock()

	item, ok := service.services[serviceID]
	return item, ok
}

// SearchServices 搜索 service 列表。
func (service *Service) SearchServices(query ServiceQuery) ([]ServiceItem, int) {
	service.mu.RLock()
	defer service.mu.RUnlock()

	var items []ServiceItem
	for _, item := range service.services {
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

// UpsertTool 创建或更新指定 service 下的 tool。
func (service *Service) UpsertTool(serviceID string, toolID string, payload ToolPayload) (ToolItem, error) {
	if serviceID == "" || toolID == "" || payload.Name == "" || payload.Path == "" {
		return ToolItem{}, fmt.Errorf("service_id、tool_id、name、path 不能为空")
	}
	if strings.Contains(payload.Name, "/") {
		return ToolItem{}, fmt.Errorf("tool name 不能包含 /")
	}

	service.mu.Lock()
	defer service.mu.Unlock()

	if _, ok := service.services[serviceID]; !ok {
		return ToolItem{}, fmt.Errorf("service not found")
	}

	if _, ok := service.tools[serviceID]; !ok {
		service.tools[serviceID] = map[string]ToolItem{}
	}
	for id, item := range service.tools[serviceID] {
		if id != toolID && item.Name == payload.Name {
			return ToolItem{}, fmt.Errorf("tool name 已存在")
		}
		if id != toolID && item.Path == payload.Path {
			return ToolItem{}, fmt.Errorf("tool path 已存在")
		}
	}

	now := time.Now()
	current, ok := service.tools[serviceID][toolID]
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

	service.tools[serviceID][toolID] = current
	return current, nil
}

// ListTools 查询指定 service 下的 tool 列表。
func (service *Service) ListTools(serviceID string, query ToolQuery) ([]ToolItem, int) {
	service.mu.RLock()
	defer service.mu.RUnlock()

	toolsByService := service.tools[serviceID]
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

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
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
