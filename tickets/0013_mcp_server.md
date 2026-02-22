# 0013 — MCP Server

## Goal

Implement the MCP server using `mark3labs/mcp-go`, exposing all 18 tools. Support `stdio`, `http` (SSE), or both transports simultaneously, driven by config.

## Dependencies

- Ticket 0004 (models + Store interface)
- Ticket 0005 (store implementation)
- Ticket 0002 (config — MCPTransport, MCPPort)

## Commands

```sh
go get github.com/mark3labs/mcp-go
```

## Tasks

### `internal/mcp/server.go`

```go
package mcp

import (
    "github.com/mark3labs/mcp-go/server"
    "github.com/rgracey/kanban-mcp/internal/store"
)

func NewServer(s store.Store) *server.MCPServer {
    srv := server.NewMCPServer("kanban-mcp", "1.0.0")
    registerTools(srv, s)
    return srv
}
```

Start function (called from `main.go`):

```go
func Start(srv *server.MCPServer, transport, mcpPort string) error
```

- `transport == "stdio"`: call `server.NewStdioServer(srv).Listen(ctx, os.Stdin, os.Stdout)`.
- `transport == "http"`: create an SSE server on `mcpPort` via `server.NewSSEServer(srv, ...)`.
- `transport == "both"`: start the SSE server in a goroutine, then start stdio (blocking).

### `internal/mcp/tools.go`

Register all 18 tools. For each tool:
- Call `srv.AddTool(mcp.NewTool(...), handler)`.
- Handler receives `(ctx, req)`, calls the store, returns `mcp.NewToolResultText(jsonString)` or an error.

#### Tool definitions

| Tool | Input schema fields | Store method |
|---|---|---|
| `list_boards` | — | `ListBoards` |
| `create_board` | `name` (string, required), `description` (string) | `CreateBoard` |
| `update_board` | `id` (string, required), `name` (string), `description` (string) | `UpdateBoard` |
| `delete_board` | `id` (string, required) | `DeleteBoard` |
| `get_board_summary` | `id` (string, required) | `GetBoardSummary` |
| `list_epics` | `board_id` (string, required) | `ListEpics` |
| `create_epic` | `board_id` (string, required), `title` (string, required), `description` (string) | `CreateEpic` |
| `update_epic` | `id` (string, required), `title` (string), `description` (string) | `UpdateEpic` |
| `delete_epic` | `id` (string, required) | `DeleteEpic` |
| `list_tickets` | `board_id` (string, required), `status` (string), `priority` (string), `epic_id` (string), `q` (string) | `ListTickets` |
| `create_ticket` | `board_id` (string, required), `title` (string, required), `description` (string), `status` (string), `priority` (string), `epic_id` (string) | `CreateTicket` |
| `update_ticket` | `id` (string, required), `title` (string), `description` (string), `status` (string), `priority` (string), `epic_id` (string) | `UpdateTicket` |
| `delete_ticket` | `id` (string, required) | `DeleteTicket` |
| `move_ticket` | `id` (string, required), `status` (string, required) | `UpdateTicket` (status field only) |
| `list_comments` | `ticket_id` (string, required) | `ListComments` |
| `add_comment` | `ticket_id` (string, required), `body` (string, required) | `CreateComment` |
| `update_comment` | `id` (string, required), `body` (string, required) | `UpdateComment` |
| `delete_comment` | `id` (string, required) | `DeleteComment` |

All tool result payloads are JSON-encoded strings of the relevant model(s).

### Wire into `main.go`

```go
mcpSrv := mcp.NewServer(store)
go func() {
    if err := mcp.Start(mcpSrv, cfg.MCPTransport, cfg.MCPPort); err != nil {
        slog.Error("MCP server error", "err", err)
    }
}()
```

The HTTP server (ticket 0008) continues running on the main goroutine.

## Acceptance Criteria

- `go build ./...` passes.
- Running `echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./kanban-mcp` (with `--mcp-transport stdio`) returns a JSON response listing all 18 tools.
- Running with `--mcp-transport http` starts an SSE endpoint on `--mcp-port` (default 8081).
- Running with `--mcp-transport both` starts both.
- Integration test in `internal/mcp/mcp_test.go`:
  - Creates a store backed by a temp SQLite DB.
  - Calls each tool handler directly (not over the wire) and asserts a non-error result.
  - Asserts `create_board` → `list_boards` returns the created board.
  - Asserts `create_ticket` → `move_ticket` → `list_tickets?status=in_progress` returns the ticket.
- Tests pass: `go test ./internal/mcp/...`
