package config

import (
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"KANBAN_PORT", "KANBAN_DB",
		"KANBAN_MCP_TRANSPORT", "KANBAN_LOG_LEVEL",
	} {
		t.Setenv(k, "") // t.Setenv restores on cleanup; setting "" clears the lookup
	}
}

func TestLoad_Defaults(t *testing.T) {
	clearEnv(t)

	cfg, err := load([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port: want 8080, got %d", cfg.Port)
	}
	if cfg.DBPath != "kanban.db" {
		t.Errorf("DBPath: want kanban.db, got %s", cfg.DBPath)
	}
	if cfg.MCPTransport != "stdio" {
		t.Errorf("MCPTransport: want stdio, got %s", cfg.MCPTransport)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel: want info, got %s", cfg.LogLevel)
	}
}

func TestLoad_EnvVars(t *testing.T) {
	t.Setenv("KANBAN_PORT", "9000")
	t.Setenv("KANBAN_DB", "/path/to/db.sqlite")
	t.Setenv("KANBAN_MCP_TRANSPORT", "http")
	t.Setenv("KANBAN_LOG_LEVEL", "debug")

	cfg, err := load([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 9000 {
		t.Errorf("Port: want 9000, got %d", cfg.Port)
	}
	if cfg.DBPath != "/path/to/db.sqlite" {
		t.Errorf("DBPath: want /path/to/db.sqlite, got %s", cfg.DBPath)
	}
	if cfg.MCPTransport != "http" {
		t.Errorf("MCPTransport: want http, got %s", cfg.MCPTransport)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel: want debug, got %s", cfg.LogLevel)
	}
}

func TestLoad_FlagOverridesEnv(t *testing.T) {
	t.Setenv("KANBAN_PORT", "9000")
	t.Setenv("KANBAN_MCP_TRANSPORT", "http")

	cfg, err := load([]string{"-port=7777", "-mcp-transport=both"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 7777 {
		t.Errorf("Port: flag should override env: want 7777, got %d", cfg.Port)
	}
	if cfg.MCPTransport != "both" {
		t.Errorf("MCPTransport: flag should override env: want both, got %s", cfg.MCPTransport)
	}
}

func TestLoad_ValidTransports(t *testing.T) {
	clearEnv(t)
	for _, transport := range []string{"stdio", "http", "both"} {
		cfg, err := load([]string{"-mcp-transport=" + transport})
		if err != nil {
			t.Errorf("transport %q: unexpected error: %v", transport, err)
			continue
		}
		if cfg.MCPTransport != transport {
			t.Errorf("MCPTransport: want %s, got %s", transport, cfg.MCPTransport)
		}
	}
}

func TestLoad_InvalidTransport(t *testing.T) {
	clearEnv(t)
	_, err := load([]string{"-mcp-transport=invalid"})
	if err == nil {
		t.Error("expected error for invalid transport, got nil")
	}
}

func TestLoad_ValidLogLevels(t *testing.T) {
	clearEnv(t)
	for _, level := range []string{"debug", "info", "warn", "error"} {
		cfg, err := load([]string{"-log-level=" + level})
		if err != nil {
			t.Errorf("level %q: unexpected error: %v", level, err)
			continue
		}
		if cfg.LogLevel != level {
			t.Errorf("LogLevel: want %s, got %s", level, cfg.LogLevel)
		}
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	clearEnv(t)
	_, err := load([]string{"-log-level=verbose"})
	if err == nil {
		t.Error("expected error for invalid log level, got nil")
	}
}

func TestLoad_Combined(t *testing.T) {
	t.Setenv("KANBAN_PORT", "1234")
	t.Setenv("KANBAN_DB", "test.db")
	t.Setenv("KANBAN_MCP_TRANSPORT", "both")
	t.Setenv("KANBAN_LOG_LEVEL", "warn")

	cfg, err := load([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 1234 {
		t.Errorf("Port: want 1234, got %d", cfg.Port)
	}
	if cfg.DBPath != "test.db" {
		t.Errorf("DBPath: want test.db, got %s", cfg.DBPath)
	}
	if cfg.MCPTransport != "both" {
		t.Errorf("MCPTransport: want both, got %s", cfg.MCPTransport)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel: want warn, got %s", cfg.LogLevel)
	}
}
