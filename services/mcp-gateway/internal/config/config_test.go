package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExpandsEnvironmentPlaceholders(t *testing.T) {
	t.Setenv("MCP_GATEWAY_ADDR", ":18080")
	t.Setenv("MYSQL_DSN", "user:pass@tcp(127.0.0.1:3306)/universal")

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("server:\n  addr: \"${MCP_GATEWAY_ADDR}\"\nmysql:\n  dsn: \"${MYSQL_DSN}\"\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Server.Addr != ":18080" {
		t.Fatalf("server addr = %q, want %q", cfg.Server.Addr, ":18080")
	}
	if cfg.MySQL.DSN != "user:pass@tcp(127.0.0.1:3306)/universal" {
		t.Fatalf("mysql dsn = %q", cfg.MySQL.DSN)
	}
}

func TestLoadFailsWhenEnvironmentPlaceholderIsMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("server:\n  addr: \"${MISSING_ADDR}\"\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("load config succeeded, want missing environment error")
	}
}

func TestLoadReadsDotEnvFromServiceRoot(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "configs")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	envPath := filepath.Join(root, ".env")
	envContent := []byte("MCP_GATEWAY_ADDR=:28080\nMYSQL_DSN=test:test@tcp(127.0.0.1:3306)/mcp_gateway\n")
	if err := os.WriteFile(envPath, envContent, 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := []byte("server:\n  addr: \"${MCP_GATEWAY_ADDR}\"\nmysql:\n  dsn: \"${MYSQL_DSN}\"\n")
	if err := os.WriteFile(configPath, configContent, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Server.Addr != ":28080" {
		t.Fatalf("server addr = %q, want %q", cfg.Server.Addr, ":28080")
	}
	if cfg.MySQL.DSN != "test:test@tcp(127.0.0.1:3306)/mcp_gateway" {
		t.Fatalf("mysql dsn = %q", cfg.MySQL.DSN)
	}
}

func TestLoadPrefersProcessEnvironmentOverDotEnv(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "configs")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	envPath := filepath.Join(root, ".env")
	envContent := []byte("MCP_GATEWAY_ADDR=:28080\nMYSQL_DSN=dotenv:dotenv@tcp(127.0.0.1:3306)/mcp_gateway\n")
	if err := os.WriteFile(envPath, envContent, 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	t.Setenv("MCP_GATEWAY_ADDR", ":38080")
	t.Setenv("MYSQL_DSN", "process:process@tcp(127.0.0.1:3306)/mcp_gateway")

	configPath := filepath.Join(configDir, "config.yaml")
	configContent := []byte("server:\n  addr: \"${MCP_GATEWAY_ADDR}\"\nmysql:\n  dsn: \"${MYSQL_DSN}\"\n")
	if err := os.WriteFile(configPath, configContent, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Server.Addr != ":38080" {
		t.Fatalf("server addr = %q, want %q", cfg.Server.Addr, ":38080")
	}
	if cfg.MySQL.DSN != "process:process@tcp(127.0.0.1:3306)/mcp_gateway" {
		t.Fatalf("mysql dsn = %q", cfg.MySQL.DSN)
	}
}
