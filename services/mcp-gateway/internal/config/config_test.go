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
