package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/rgracey/kanban-mcp/internal/api"
	"github.com/rgracey/kanban-mcp/internal/config"
	"github.com/rgracey/kanban-mcp/internal/db"
	"github.com/rgracey/kanban-mcp/internal/store"
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

	// Create store and API router
	store := store.NewSQLiteStore(dbConn)
	router := api.NewRouter(store)

	fmt.Printf("Starting server on port %d\n", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
