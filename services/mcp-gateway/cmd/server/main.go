package main

import (
	"log"
	"net/http"
	"os"

	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/config"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/runtime"
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

	if err := http.ListenAndServe(cfg.Server.Addr, runtime.NewRouter(runtime.NewService(nil, nil))); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
