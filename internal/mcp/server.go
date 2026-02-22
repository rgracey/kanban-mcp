package mcp

import (
	"context"
	"net/http"
	"os"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rgracey/kanban-mcp/internal/store"
)

// NewServer creates a new MCP server with all tools registered.
func NewServer(s store.Store) *server.MCPServer {
	srv := server.NewMCPServer("kanban-mcp", "1.0.0",
		server.WithToolCapabilities(false),
	)
	registerTools(srv, s)
	return srv
}

// NewHTTPHandler returns an http.Handler for the MCP Streamable HTTP transport.
// Mount it at /mcp on your existing HTTP server.
func NewHTTPHandler(srv *server.MCPServer) http.Handler {
	return server.NewStreamableHTTPServer(srv)
}

// StartStdio runs the MCP stdio transport, blocking until the stream closes.
func StartStdio(srv *server.MCPServer) error {
	return server.NewStdioServer(srv).Listen(context.Background(), os.Stdin, os.Stdout)
}

// jsonResult marshals v to JSON and returns a text tool result.
func jsonResult(v any) (*mcpgo.CallToolResult, error) {
	result, err := mcpgo.NewToolResultJSON(v)
	if err != nil {
		return mcpgo.NewToolResultError(err.Error()), nil
	}
	return result, nil
}
