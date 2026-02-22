package main

import (
	"fmt"
	"log"

	"github.com/rgracey/kanban-mcp/internal/config"
	"github.com/rgracey/kanban-mcp/internal/db"
)

func main() {
	cfg := config.Load()
	fmt.Printf("Config loaded: %+v\n", cfg)

	// Open database with migrations
	dbConn, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer dbConn.Close()

	fmt.Printf("Database opened successfully at %s\n", cfg.DBPath)
}
