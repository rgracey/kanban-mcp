package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port         int
	DBPath       string
	MCPTransport string
	MCPPort      int
	LogLevel     string
}

// Load parses flags and environment variables to populate Config
func Load() Config {
	c := Config{}

	// Define custom flag types to track if flags were set
	portFlag := flag.Int("port", 8080, "Port to listen on")
	dbFlag := flag.String("db", "kanban.db", "Path to the database file")
	mcpTransportFlag := flag.String("mcp-transport", "stdio", "MCP transport method (stdio, http, both)")
	mcpPortFlag := flag.Int("mcp-port", 8081, "Port for MCP server")
	logLevelFlag := flag.String("log-level", "info", "Log level (debug, info, warn, error)")

	flag.Parse()

	// Helper function to get env var or default
	getEnvOr := func(key, fallback string) string {
		if val := os.Getenv(key); val != "" {
			return val
		}
		return fallback
	}

	// Helper function to get int env var or default
	getEnvIntOr := func(key string, fallback int) int {
		if val := os.Getenv(key); val != "" {
			if parsed, err := strconv.Atoi(val); err == nil {
				return parsed
			}
		}
		return fallback
	}

	// Apply flags, then env vars, then defaults
	c.Port = *portFlag
	if portFlag == nil || os.Getenv("KANBAN_PORT") != "" {
		c.Port = getEnvIntOr("KANBAN_PORT", 8080)
	}

	c.DBPath = *dbFlag
	if dbFlag == nil || os.Getenv("KANBAN_DB") != "" {
		c.DBPath = getEnvOr("KANBAN_DB", "kanban.db")
	}

	c.MCPTransport = *mcpTransportFlag
	if mcpTransportFlag == nil || os.Getenv("KANBAN_MCP_TRANSPORT") != "" {
		c.MCPTransport = getEnvOr("KANBAN_MCP_TRANSPORT", "stdio")
	}

	c.MCPPort = *mcpPortFlag
	if mcpPortFlag == nil || os.Getenv("KANBAN_MCP_PORT") != "" {
		c.MCPPort = getEnvIntOr("KANBAN_MCP_PORT", 8081)
	}

	c.LogLevel = *logLevelFlag
	if logLevelFlag == nil || os.Getenv("KANBAN_LOG_LEVEL") != "" {
		c.LogLevel = getEnvOr("KANBAN_LOG_LEVEL", "info")
	}

	// Validate MCPTransport
	validTransports := map[string]bool{
		"stdio": true,
		"http":  true,
		"both":  true,
	}
	if !validTransports[c.MCPTransport] {
		panic("invalid MCP transport: must be one of stdio, http, both")
	}

	// Validate LogLevel
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		panic("invalid log level: must be one of debug, info, warn, error")
	}

	return c
}
