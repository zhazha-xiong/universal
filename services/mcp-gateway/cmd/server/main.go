package main

import (
	"log"
	"net/http"
	"os"

	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/admin"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/config"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/runtime"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/server"
)

func main() {
	path := os.Getenv("MCP_GATEWAY_CONFIG")
	if path == "" {
		path = "configs/config.yaml"
	}

	cfg, err := config.Load(path)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	router := server.NewRouter(admin.NewService(), runtime.NewService(nil, nil))
	if err := http.ListenAndServe(cfg.Server.Addr, router); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
