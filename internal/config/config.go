package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration.
// Precedence (highest to lowest): CLI flag > environment variable > default.
type Config struct {
	Port         int
	DBPath       string
	MCPTransport string
	MCPPort      int
	LogLevel     string
}

// Load parses CLI flags and environment variables into a Config.
// It writes a diagnostic message to stderr and calls os.Exit(1) on invalid values.
func Load() Config {
	cfg, err := load(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

// load is the testable core: it parses args and returns an error on bad input.
func load(args []string) (Config, error) {
	fs := flag.NewFlagSet("kanban-mcp", flag.ContinueOnError)

	port := fs.Int("port", envInt("KANBAN_PORT", 8080), "HTTP server port")
	dbPath := fs.String("db", envStr("KANBAN_DB", "kanban.db"), "path to SQLite database file")
	mcpTransport := fs.String("mcp-transport", envStr("KANBAN_MCP_TRANSPORT", "stdio"), "MCP transport (stdio, http, both)")
	mcpPort := fs.Int("mcp-port", envInt("KANBAN_MCP_PORT", 8081), "MCP HTTP server port")
	logLevel := fs.String("log-level", envStr("KANBAN_LOG_LEVEL", "info"), "log level (debug, info, warn, error)")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Port:         *port,
		DBPath:       *dbPath,
		MCPTransport: *mcpTransport,
		MCPPort:      *mcpPort,
		LogLevel:     *logLevel,
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	switch c.MCPTransport {
	case "stdio", "http", "both":
	default:
		return fmt.Errorf("invalid MCP transport %q: must be stdio, http, or both", c.MCPTransport)
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid log level %q: must be debug, info, warn, or error", c.LogLevel)
	}

	return nil
}

// envStr returns the value of the environment variable key, or fallback if unset/empty.
func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envInt returns the integer value of the environment variable key, or fallback if unset/empty/invalid.
func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
