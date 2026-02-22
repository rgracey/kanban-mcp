package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rgracey/kanban-mcp/internal/api"
	"github.com/rgracey/kanban-mcp/internal/config"
	"github.com/rgracey/kanban-mcp/internal/db"
	internalmcp "github.com/rgracey/kanban-mcp/internal/mcp"
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

	slog.Info("database opened", "path", cfg.DBPath)

	// Create store, MCP server, and API router
	store := store.NewSQLiteStore(dbConn)
	mcpSrv := internalmcp.NewServer(store)
	hub := api.NewHub()

	// When using HTTP transport, mount the MCP handler on /mcp within the main
	// router so both the SPA/API and MCP share a single port.
	var mcpHandler http.Handler
	if cfg.MCPTransport == "http" || cfg.MCPTransport == "both" {
		mcpHandler = internalmcp.NewHTTPHandler(mcpSrv)
		slog.Info("MCP HTTP handler mounted", "path", "/mcp", "port", cfg.Port)
	}

	router := api.NewRouter(store, hub, mcpHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		slog.Info("HTTP server listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// For stdio and both transports, start the stdio MCP listener.
	if cfg.MCPTransport == "stdio" || cfg.MCPTransport == "both" {
		go func() {
			slog.Info("MCP stdio transport starting")
			if err := internalmcp.StartStdio(mcpSrv); err != nil {
				slog.Error("MCP stdio error", "err", err)
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
	slog.Info("shutdown complete")

	dbConn.Close()
}
