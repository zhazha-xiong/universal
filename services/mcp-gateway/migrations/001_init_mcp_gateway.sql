CREATE TABLE IF NOT EXISTS mcp_service (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  service_id VARCHAR(128) NOT NULL,
  name VARCHAR(128) NOT NULL,
  description VARCHAR(512) NULL,
  base_url VARCHAR(512) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  tags VARCHAR(512) NULL,
  owner_id VARCHAR(128) NULL,
  ext VARCHAR(2048) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_mcp_service_service_id (service_id),
  UNIQUE KEY uk_mcp_service_name (name)
);

CREATE TABLE IF NOT EXISTS mcp_tool (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tool_id VARCHAR(128) NOT NULL,
  service_id VARCHAR(128) NOT NULL,
  name VARCHAR(128) NOT NULL,
  description VARCHAR(512) NULL,
  path VARCHAR(512) NOT NULL,
  method VARCHAR(16) NOT NULL DEFAULT 'POST',
  input_schema VARCHAR(4096) NULL,
  output_schema VARCHAR(4096) NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'active',
  tags VARCHAR(512) NULL,
  ext VARCHAR(2048) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_mcp_tool_tool_id (tool_id),
  UNIQUE KEY uk_mcp_tool_service_name (service_id, name),
  UNIQUE KEY uk_mcp_tool_service_path (service_id, path),
  KEY idx_mcp_tool_service_id (service_id)
);
