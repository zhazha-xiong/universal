package main

import (
	"log"
	"net/http"
	"os"

	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/admin"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/config"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/repository"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/runtime"
	"github.com/zhazha-xiong/universal/services/mcp-gateway/internal/server"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("open mysql: %v", err)
	}

	repo := repository.NewRepository(db)
	router := server.NewRouter(admin.NewServiceWithStore(repo), runtime.NewService(repo, nil))
	if err := http.ListenAndServe(cfg.Server.Addr, router); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
