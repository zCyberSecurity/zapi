package main

import (
	"log"

	"github.com/zCyberSecurity/zapi/internal/config"
	"github.com/zCyberSecurity/zapi/internal/database"
	"github.com/zCyberSecurity/zapi/internal/router"
)

func main() {
	cfg := config.Load()
	db := database.Init(cfg.DBPath)
	r := router.New(db, cfg)
	log.Printf("zAPI listening on %s", cfg.Addr)
	log.Fatal(r.Run(cfg.Addr))
}
