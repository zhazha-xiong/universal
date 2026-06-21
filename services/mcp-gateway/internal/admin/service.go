package admin

import "time"

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

// Service 编排管理侧 service 和 tool 的业务逻辑。
type Service struct {
	store Store
}

// NewService 创建管理侧服务。
func NewService() *Service {
	return NewServiceWithStore(NewMemoryStore())
}

// NewServiceWithStore 使用指定存储创建管理侧服务。
func NewServiceWithStore(store Store) *Service {
	return &Service{store: store}
}

// UpsertService 创建或更新 service。
func (service *Service) UpsertService(serviceID string, payload ServicePayload) (ServiceItem, error) {
	return service.store.UpsertService(serviceID, payload)
}

// GetService 查询单个 service。
func (service *Service) GetService(serviceID string) (ServiceItem, bool) {
	return service.store.GetService(serviceID)
}

// SearchServices 搜索 service 列表。
func (service *Service) SearchServices(query ServiceQuery) ([]ServiceItem, int) {
	return service.store.SearchServices(query)
}

// UpsertTool 创建或更新指定 service 下的 tool。
func (service *Service) UpsertTool(serviceID string, toolID string, payload ToolPayload) (ToolItem, error) {
	return service.store.UpsertTool(serviceID, toolID, payload)
}

// ListTools 查询指定 service 下的 tool 列表。
func (service *Service) ListTools(serviceID string, query ToolQuery) ([]ToolItem, int) {
	return service.store.ListTools(serviceID, query)
}
