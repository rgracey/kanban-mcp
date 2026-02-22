package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration.
type Config struct {
	Port         int
	DBPath       string
	MCPTransport string
	MCPPort      int
	LogLevel     string
}

// flagSet tracks whether a flag was explicitly set by the user.
type flagSet struct {
	value int
	set   bool
}

func (f *flagSet) Set(val string) error {
	i, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	f.value = i
	f.set = true
	return nil
}

func (f *flagSet) String() string {
	return strconv.Itoa(f.value)
}

type stringFlagSet struct {
	value string
	set   bool
}

func (f *stringFlagSet) Set(val string) error {
	f.value = val
	f.set = true
	return nil
}

func (f *stringFlagSet) String() string {
	return f.value
}

// LoadConfigWithArgs parses flags and environment variables to populate Config.
// This is the internal function that takes a flag.FlagSet for testability.
func LoadConfigWithArgs(fs *flag.FlagSet, args []string) Config {
	// Initialize flag tracking structs
	var portFlag flagSet
	var dbFlag stringFlagSet
	var mcpTransportFlag stringFlagSet
	var mcpPortFlag flagSet
	var logLevelFlag stringFlagSet

	// Set default values
	portFlag.value = 8080
	dbFlag.value = "kanban.db"
	mcpTransportFlag.value = "stdio"
	mcpPortFlag.value = 8081
	logLevelFlag.value = "info"

	// Register custom flags with the provided FlagSet
	fs.Var(&portFlag, "port", "Port to listen on")
	fs.Var(&dbFlag, "db", "Path to the database file")
	fs.Var(&mcpTransportFlag, "mcp-transport", "MCP transport method (stdio, http, both)")
	fs.Var(&mcpPortFlag, "mcp-port", "Port for MCP server")
	fs.Var(&logLevelFlag, "log-level", "Log level (debug, info, warn, error)")

	fs.Parse(args)

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

	c := Config{}

	// For each field: if flag was set, use it; else check env var; else use default
	if portFlag.set {
		c.Port = portFlag.value
	} else {
		c.Port = getEnvIntOr("KANBAN_PORT", 8080)
	}

	if dbFlag.set {
		c.DBPath = dbFlag.value
	} else {
		c.DBPath = getEnvOr("KANBAN_DB", "kanban.db")
	}

	if mcpTransportFlag.set {
		c.MCPTransport = mcpTransportFlag.value
	} else {
		c.MCPTransport = getEnvOr("KANBAN_MCP_TRANSPORT", "stdio")
	}

	if mcpPortFlag.set {
		c.MCPPort = mcpPortFlag.value
	} else {
		c.MCPPort = getEnvIntOr("KANBAN_MCP_PORT", 8081)
	}

	if logLevelFlag.set {
		c.LogLevel = logLevelFlag.value
	} else {
		c.LogLevel = getEnvOr("KANBAN_LOG_LEVEL", "info")
	}

	// Validate MCPTransport
	validTransports := map[string]bool{
		"stdio": true,
		"http":  true,
		"both":  true,
	}
	if !validTransports[c.MCPTransport] {
		fmt.Fprintf(os.Stderr, "error: invalid MCP transport '%s': must be one of stdio, http, both\n", c.MCPTransport)
		os.Exit(1)
	}

	// Validate LogLevel
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		fmt.Fprintf(os.Stderr, "error: invalid log level '%s': must be one of debug, info, warn, error\n", c.LogLevel)
		os.Exit(1)
	}

	return c
}

// Load parses flags and environment variables to populate Config.
// Flags take precedence over environment variables, which take precedence over defaults.
func Load() Config {
	return LoadConfigWithArgs(flag.CommandLine, os.Args[1:])
}
