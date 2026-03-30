package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/livingopensource/bind-dns-api/internal/api"
	"github.com/livingopensource/bind-dns-api/internal/bind"
	"github.com/livingopensource/bind-dns-api/internal/config"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Ensure zone directory exists
	if err := os.MkdirAll(cfg.BIND.ZoneDirectory, 0755); err != nil {
		log.Fatalf("Failed to create zone directory: %v", err)
	}

	// Create BIND manager
	manager := bind.NewManager(&cfg.BIND)

	// Create API handler
	handler := api.NewHandler(manager)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Register routes
	handler.RegisterRoutes(router)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting BIND DNS API server on %s", addr)
	log.Printf("Zone directory: %s", cfg.BIND.ZoneDirectory)
	log.Printf("Configuration: %s", *configPath)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
