package mcp

import (
	"context"
	"fmt"
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

// jsonResult marshals v to JSON and returns a structured tool result.
// v must be a JSON object (map or struct); use jsonListResult for slices.
func jsonResult(v any) (*mcpgo.CallToolResult, error) {
	result, err := mcpgo.NewToolResultJSON(v)
	if err != nil {
		return mcpgo.NewToolResultError(err.Error()), nil
	}
	return result, nil
}

// jsonListResult wraps a slice in {"items": [...]} so that structuredContent
// is always a JSON object, which is required by the MCP spec.
func jsonListResult(items any) (*mcpgo.CallToolResult, error) {
	return jsonResult(map[string]any{"items": items})
}

// resolveBoardID returns id unchanged if non-empty. If id is empty and name is
// non-empty it looks up the board by exact name and returns its ID. Returns an
// error string suitable for NewToolResultError if neither resolves.
func resolveBoardID(ctx context.Context, s store.Store, id, name string) (string, error) {
	if id != "" {
		return id, nil
	}
	if name != "" {
		b, err := s.GetBoardByName(ctx, name)
		if err != nil {
			return "", fmt.Errorf("board with name %q not found", name)
		}
		return b.ID, nil
	}
	return "", fmt.Errorf("id or name required")
}
