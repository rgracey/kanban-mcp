package mcp

import (
	"context"
	"fmt"
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

// Start launches the MCP server using the configured transport.
// For "stdio" it blocks; for "http" it starts a Streamable HTTP server; for "both" it
// runs Streamable HTTP in a goroutine then blocks on stdio.
func Start(srv *server.MCPServer, transport, mcpPort string) error {
	switch transport {
	case "stdio":
		return server.NewStdioServer(srv).Listen(context.Background(), os.Stdin, os.Stdout)

	case "http":
		addr := fmt.Sprintf(":%s", mcpPort)
		httpSrv := server.NewStreamableHTTPServer(srv)
		return httpSrv.Start(addr)

	case "both":
		addr := fmt.Sprintf(":%s", mcpPort)
		httpSrv := server.NewStreamableHTTPServer(srv)
		go func() {
			if err := httpSrv.Start(addr); err != nil {
				fmt.Fprintf(os.Stderr, "MCP HTTP server error: %v\n", err)
			}
		}()
		return server.NewStdioServer(srv).Listen(context.Background(), os.Stdin, os.Stdout)

	default:
		return fmt.Errorf("unknown transport %q: must be stdio, http, or both", transport)
	}
}

// jsonResult marshals v to JSON and returns a text tool result.
func jsonResult(v any) (*mcpgo.CallToolResult, error) {
	result, err := mcpgo.NewToolResultJSON(v)
	if err != nil {
		return mcpgo.NewToolResultError(err.Error()), nil
	}
	return result, nil
}
