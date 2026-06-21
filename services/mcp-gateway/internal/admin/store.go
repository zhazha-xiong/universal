package admin

// Store 定义管理侧 service 和 tool 的存储能力。
type Store interface {
	UpsertService(serviceID string, payload ServicePayload) (ServiceItem, error)
	GetService(serviceID string) (ServiceItem, bool)
	SearchServices(query ServiceQuery) ([]ServiceItem, int)
	UpsertTool(serviceID string, toolID string, payload ToolPayload) (ToolItem, error)
	ListTools(serviceID string, query ToolQuery) ([]ToolItem, int)
}
