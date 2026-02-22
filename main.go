package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/rgracey/kanban-mcp/internal/api"
	"github.com/rgracey/kanban-mcp/internal/config"
	"github.com/rgracey/kanban-mcp/internal/db"
	"github.com/rgracey/kanban-mcp/internal/store"
)

func main() {
	cfg := config.Load()

	// Set up structured logging with JSON output
	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	slog.Debug("config loaded", "config", cfg)

	// Open database with migrations
	dbConn, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		log.Fatal(err)
	}
	defer dbConn.Close()

	slog.Info("database opened", "path", cfg.DBPath)

	// Create store and API router
	store := store.NewSQLiteStore(dbConn)
	router := api.NewRouter(store)

	slog.Info("HTTP server listening", "port", cfg.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router); err != nil {
		slog.Error("server failed", "error", err)
		log.Fatal(err)
	}
}
