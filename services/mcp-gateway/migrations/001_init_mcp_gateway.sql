CREATE TABLE IF NOT EXISTS mcp_service (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  service_id VARCHAR(128) NOT NULL COMMENT '服务唯一标识',
  name VARCHAR(128) NOT NULL COMMENT '服务名称，对外暴露名，要求全局唯一',
  description VARCHAR(512) NULL COMMENT '服务描述',
  base_url VARCHAR(512) NOT NULL COMMENT '服务基础地址，用于拼接上游工具请求 URL',
  status VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '服务状态，例如 active 或 inactive',
  tags VARCHAR(512) NULL COMMENT '服务标签列表，JSON 字符串存储',
  owner_id VARCHAR(128) NULL COMMENT '服务负责人或所属人标识',
  ext VARCHAR(2048) NULL COMMENT '服务扩展信息，JSON 字符串存储',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_mcp_service_service_id (service_id),
  UNIQUE KEY uk_mcp_service_name (name)
) COMMENT='MCP 上游服务表';

CREATE TABLE IF NOT EXISTS mcp_tool (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  tool_id VARCHAR(128) NOT NULL COMMENT '工具唯一标识',
  service_id VARCHAR(128) NOT NULL COMMENT '所属服务标识，关联 mcp_service.service_id',
  name VARCHAR(128) NOT NULL COMMENT '工具名称，在服务内唯一',
  description VARCHAR(512) NULL COMMENT '工具描述',
  path VARCHAR(512) NOT NULL COMMENT '工具对应的上游请求路径',
  method VARCHAR(16) NOT NULL DEFAULT 'POST' COMMENT '工具对应的上游 HTTP 方法',
  input_schema VARCHAR(4096) NULL COMMENT '工具输入参数结构，JSON Schema 字符串存储',
  output_schema VARCHAR(4096) NULL COMMENT '工具输出结果结构，JSON Schema 字符串存储',
  status VARCHAR(32) NOT NULL DEFAULT 'active' COMMENT '工具状态，例如 active 或 inactive',
  tags VARCHAR(512) NULL COMMENT '工具标签列表，JSON 字符串存储',
  ext VARCHAR(2048) NULL COMMENT '工具扩展信息，JSON 字符串存储',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_mcp_tool_tool_id (tool_id),
  UNIQUE KEY uk_mcp_tool_service_name (service_id, name),
  UNIQUE KEY uk_mcp_tool_service_path (service_id, path),
  KEY idx_mcp_tool_service_id (service_id)
) COMMENT='MCP 工具元数据表';
