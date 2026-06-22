package repository

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/admin"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/runtime"
)

// Repository 提供基于 GORM 的管理侧数据访问能力。
type Repository struct {
	db *gorm.DB
}

type serviceRecord struct {
	ID          uint64 `gorm:"primaryKey"`
	ServiceID   string `gorm:"column:service_id;size:128;uniqueIndex"`
	Name        string `gorm:"column:name;size:128;uniqueIndex"`
	Description string `gorm:"column:description;size:512"`
	BaseURL     string `gorm:"column:base_url;size:512"`
	Status      string `gorm:"column:status;size:32"`
	Tags        string `gorm:"column:tags;size:512"`
	OwnerID     string `gorm:"column:owner_id;size:128"`
	Ext         string `gorm:"column:ext;size:2048"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (serviceRecord) TableName() string {
	return "mcp_service"
}

type toolRecord struct {
	ID           uint64 `gorm:"primaryKey"`
	ToolID       string `gorm:"column:tool_id;size:128;uniqueIndex"`
	ServiceID    string `gorm:"column:service_id;size:128;index;uniqueIndex:uk_service_name,priority:1;uniqueIndex:uk_service_path,priority:1"`
	Name         string `gorm:"column:name;size:128;uniqueIndex:uk_service_name,priority:2"`
	Description  string `gorm:"column:description;size:512"`
	Path         string `gorm:"column:path;size:512;uniqueIndex:uk_service_path,priority:2"`
	Method       string `gorm:"column:method;size:16"`
	InputSchema  string `gorm:"column:input_schema;size:4096"`
	OutputSchema string `gorm:"column:output_schema;size:4096"`
	Status       string `gorm:"column:status;size:32"`
	Tags         string `gorm:"column:tags;size:512"`
	Ext          string `gorm:"column:ext;size:2048"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (toolRecord) TableName() string {
	return "mcp_tool"
}

// NewRepository 创建 GORM repository。
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// AutoMigrate 自动迁移当前 repository 使用的表结构。
func (repository *Repository) AutoMigrate() error {
	return repository.db.AutoMigrate(&serviceRecord{}, &toolRecord{})
}

// UpsertService 创建或更新 service。
func (repository *Repository) UpsertService(serviceID string, payload admin.ServicePayload) (admin.ServiceItem, error) {
	if serviceID == "" || payload.Name == "" || payload.BaseURL == "" {
		return admin.ServiceItem{}, fmt.Errorf("service_id、name、base_url 不能为空")
	}
	if strings.Contains(payload.Name, "/") {
		return admin.ServiceItem{}, fmt.Errorf("service name 不能包含 /")
	}

	var duplicated int64
	if err := repository.db.Model(&serviceRecord{}).
		Where("name = ? AND service_id <> ?", payload.Name, serviceID).
		Count(&duplicated).Error; err != nil {
		return admin.ServiceItem{}, err
	}
	if duplicated > 0 {
		return admin.ServiceItem{}, fmt.Errorf("service name 已存在")
	}

	now := time.Now()
	var record serviceRecord
	err := repository.db.Where("service_id = ?", serviceID).First(&record).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return admin.ServiceItem{}, err
	}
	if err == gorm.ErrRecordNotFound {
		record.ServiceID = serviceID
		record.CreatedAt = now
	}

	record.Name = payload.Name
	record.Description = payload.Description
	record.BaseURL = payload.BaseURL
	record.Status = normalizeStatus(payload.Status)
	record.Tags = encodeStrings(payload.Tags)
	record.OwnerID = payload.OwnerID
	record.Ext = encodeMap(payload.Ext)
	record.UpdatedAt = now

	if err := repository.db.Save(&record).Error; err != nil {
		return admin.ServiceItem{}, err
	}
	return mapServiceRecord(record), nil
}

// GetService 查询单个 service。
func (repository *Repository) GetService(serviceID string) (admin.ServiceItem, bool) {
	var record serviceRecord
	if err := repository.db.Where("service_id = ?", serviceID).First(&record).Error; err != nil {
		return admin.ServiceItem{}, false
	}
	return mapServiceRecord(record), true
}

// SearchServices 搜索 service 列表。
func (repository *Repository) SearchServices(query admin.ServiceQuery) ([]admin.ServiceItem, int) {
	db := repository.db.Model(&serviceRecord{})
	if query.ServiceID != "" {
		db = db.Where("service_id = ?", query.ServiceID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.OwnerID != "" {
		db = db.Where("owner_id = ?", query.OwnerID)
	}
	if query.Tag != "" {
		db = db.Where("tags LIKE ?", "%\""+query.Tag+"\"%")
	}
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ?", keyword, keyword)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return []admin.ServiceItem{}, 0
	}

	limit := max(query.PageSize, 20)
	offset := (max(query.Page, 1) - 1) * limit

	var records []serviceRecord
	if err := db.Order("service_id ASC").Offset(offset).Limit(limit).Find(&records).Error; err != nil {
		return []admin.ServiceItem{}, 0
	}

	items := make([]admin.ServiceItem, 0, len(records))
	for _, record := range records {
		items = append(items, mapServiceRecord(record))
	}
	return items, int(total)
}

// UpsertTool 创建或更新 tool。
func (repository *Repository) UpsertTool(serviceID string, toolID string, payload admin.ToolPayload) (admin.ToolItem, error) {
	if serviceID == "" || toolID == "" || payload.Name == "" || payload.Path == "" {
		return admin.ToolItem{}, fmt.Errorf("service_id、tool_id、name、path 不能为空")
	}
	if strings.Contains(payload.Name, "/") {
		return admin.ToolItem{}, fmt.Errorf("tool name 不能包含 /")
	}

	var serviceCount int64
	if err := repository.db.Model(&serviceRecord{}).Where("service_id = ?", serviceID).Count(&serviceCount).Error; err != nil {
		return admin.ToolItem{}, err
	}
	if serviceCount == 0 {
		return admin.ToolItem{}, fmt.Errorf("service not found")
	}

	var duplicated int64
	if err := repository.db.Model(&toolRecord{}).
		Where("service_id = ? AND name = ? AND tool_id <> ?", serviceID, payload.Name, toolID).
		Count(&duplicated).Error; err != nil {
		return admin.ToolItem{}, err
	}
	if duplicated > 0 {
		return admin.ToolItem{}, fmt.Errorf("tool name 已存在")
	}

	duplicated = 0
	if err := repository.db.Model(&toolRecord{}).
		Where("service_id = ? AND path = ? AND tool_id <> ?", serviceID, payload.Path, toolID).
		Count(&duplicated).Error; err != nil {
		return admin.ToolItem{}, err
	}
	if duplicated > 0 {
		return admin.ToolItem{}, fmt.Errorf("tool path 已存在")
	}

	now := time.Now()
	var record toolRecord
	err := repository.db.Where("tool_id = ?", toolID).First(&record).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return admin.ToolItem{}, err
	}
	if err == gorm.ErrRecordNotFound {
		record.ToolID = toolID
		record.CreatedAt = now
	}

	record.ServiceID = serviceID
	record.Name = payload.Name
	record.Description = payload.Description
	record.Path = payload.Path
	record.Method = normalizeMethod(payload.Method)
	record.InputSchema = encodeMap(payload.InputSchema)
	record.OutputSchema = encodeMap(payload.OutputSchema)
	record.Status = normalizeStatus(payload.Status)
	record.Tags = encodeStrings(payload.Tags)
	record.Ext = encodeMap(payload.Ext)
	record.UpdatedAt = now

	if err := repository.db.Save(&record).Error; err != nil {
		return admin.ToolItem{}, err
	}
	return mapToolRecord(record), nil
}

// ListTools 查询指定 service 下的 tool 列表。
func (repository *Repository) ListTools(serviceID string, query admin.ToolQuery) ([]admin.ToolItem, int) {
	db := repository.db.Model(&toolRecord{}).Where("service_id = ?", serviceID)
	if query.ToolID != "" {
		db = db.Where("tool_id = ?", query.ToolID)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Tag != "" {
		db = db.Where("tags LIKE ?", "%\""+query.Tag+"\"%")
	}
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ?", keyword, keyword)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return []admin.ToolItem{}, 0
	}

	limit := max(query.PageSize, 20)
	offset := (max(query.Page, 1) - 1) * limit

	var records []toolRecord
	if err := db.Order("tool_id ASC").Offset(offset).Limit(limit).Find(&records).Error; err != nil {
		return []admin.ToolItem{}, 0
	}

	items := make([]admin.ToolItem, 0, len(records))
	for _, record := range records {
		items = append(items, mapToolRecord(record))
	}
	return items, int(total)
}

// ListRuntimeTools 返回当前可暴露给 MCP 运行时的工具列表。
func (repository *Repository) ListRuntimeTools() []runtime.Tool {
	rows := make([]runtimeToolRow, 0)
	if err := repository.runtimeToolQuery().Order("mcp_service.name ASC, mcp_tool.name ASC").Find(&rows).Error; err != nil {
		return []runtime.Tool{}
	}

	tools := make([]runtime.Tool, 0, len(rows))
	for _, row := range rows {
		tools = append(tools, row.toRuntimeTool())
	}
	return tools
}

// FindRuntimeTool 按运行时暴露名查询可调用工具。
func (repository *Repository) FindRuntimeTool(name string) (runtime.Tool, bool) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return runtime.Tool{}, false
	}

	var row runtimeToolRow
	err := repository.runtimeToolQuery().
		Where("mcp_service.name = ? AND mcp_tool.name = ?", parts[0], parts[1]).
		Take(&row).Error
	if err != nil {
		return runtime.Tool{}, false
	}
	return row.toRuntimeTool(), true
}

type runtimeToolRow struct {
	ServiceName string `gorm:"column:service_name"`
	BaseURL     string `gorm:"column:base_url"`
	ToolName    string `gorm:"column:tool_name"`
	Description string `gorm:"column:description"`
	Path        string `gorm:"column:path"`
	Method      string `gorm:"column:method"`
	InputSchema string `gorm:"column:input_schema"`
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

func encodeStrings(items []string) string {
	if len(items) == 0 {
		return ""
	}
	content, err := json.Marshal(items)
	if err != nil {
		return ""
	}
	return string(content)
}

func decodeStrings(raw string) []string {
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func encodeMap(input map[string]any) string {
	if len(input) == 0 {
		return ""
	}
	content, err := json.Marshal(input)
	if err != nil {
		return ""
	}
	return string(content)
}

func decodeMap(raw string) map[string]any {
	if raw == "" {
		return nil
	}
	var output map[string]any
	if err := json.Unmarshal([]byte(raw), &output); err != nil {
		return nil
	}
	return output
}

func mapServiceRecord(record serviceRecord) admin.ServiceItem {
	return admin.ServiceItem{
		ServiceID:   record.ServiceID,
		Name:        record.Name,
		Description: record.Description,
		BaseURL:     record.BaseURL,
		Status:      record.Status,
		Tags:        decodeStrings(record.Tags),
		OwnerID:     record.OwnerID,
		Ext:         decodeMap(record.Ext),
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}
}

func mapToolRecord(record toolRecord) admin.ToolItem {
	return admin.ToolItem{
		ToolID:       record.ToolID,
		ServiceID:    record.ServiceID,
		Name:         record.Name,
		Description:  record.Description,
		Path:         record.Path,
		Method:       record.Method,
		InputSchema:  decodeMap(record.InputSchema),
		OutputSchema: decodeMap(record.OutputSchema),
		Status:       record.Status,
		Tags:         decodeStrings(record.Tags),
		Ext:          decodeMap(record.Ext),
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}
}

func (repository *Repository) runtimeToolQuery() *gorm.DB {
	return repository.db.Table("mcp_tool").
		Select([]string{
			"mcp_service.name AS service_name",
			"mcp_service.base_url AS base_url",
			"mcp_tool.name AS tool_name",
			"mcp_tool.description AS description",
			"mcp_tool.path AS path",
			"mcp_tool.method AS method",
			"mcp_tool.input_schema AS input_schema",
		}).
		Joins("JOIN mcp_service ON mcp_service.service_id = mcp_tool.service_id").
		Where("mcp_service.status = ? AND mcp_tool.status = ?", "active", "active")
}

func (row runtimeToolRow) toRuntimeTool() runtime.Tool {
	return runtime.Tool{
		Name:        row.ServiceName + "/" + row.ToolName,
		Description: row.Description,
		InputSchema: decodeMap(row.InputSchema),
		Method:      row.Method,
		URL:         joinURL(row.BaseURL, row.Path),
	}
}

func joinURL(baseURL string, path string) string {
	if baseURL == "" {
		return path
	}
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
}

func max(value int, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
