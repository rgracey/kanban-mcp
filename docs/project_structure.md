# Project Structure

```
kanban-mcp/
├── go.mod
├── go.sum
├── main.go                  # Entrypoint: parses config, wires components, starts servers
├── internal/
│   ├── config/
│   │   └── config.go        # Config struct, flag + env parsing
│   ├── db/
│   │   ├── db.go            # Open/close SQLite, run migrations
│   │   └── migrations/
│   │       ├── 0001_init.up.sql
│   │       └── 0001_init.down.sql
│   ├── models/
│   │   └── models.go        # Board, Epic, Ticket, Comment structs
│   ├── store/
│   │   ├── store.go         # Store interface
│   │   ├── board.go
│   │   ├── epic.go
│   │   ├── ticket.go
│   │   └── comment.go
│   ├── api/
│   │   ├── router.go        # chi router setup, middleware
│   │   ├── boards.go
│   │   ├── epics.go
│   │   ├── tickets.go
│   │   └── comments.go
│   └── mcp/
│       ├── server.go        # MCP server setup, transport selection
│       └── tools.go         # All MCP tool registrations
└── frontend/
    ├── package.json
    ├── vite.config.ts
    ├── tailwind.config.ts
    ├── src/
    │   ├── main.ts
    │   ├── App.svelte
    │   ├── lib/
    │   │   ├── api.ts       # Typed fetch wrappers for REST API
    │   │   └── types.ts     # TypeScript types mirroring Go models
    │   └── components/
    │       ├── BoardSwitcher.svelte
    │       ├── KanbanBoard.svelte
    │       ├── KanbanColumn.svelte
    │       ├── TicketCard.svelte
    │       ├── TicketDetail.svelte
    │       ├── EpicFilter.svelte
    │       └── modals/
    │           ├── CreateBoard.svelte
    │           ├── CreateEpic.svelte
    │           └── CreateTicket.svelte
    └── dist/                # Built output — embedded into Go binary
```

## Build

Build the frontend first, then the Go binary:

```sh
cd frontend && npm run build && cd ..
go build -o kanban-mcp ./...
```

## Embed

`internal/api/router.go` (or a dedicated `embed.go`) contains:

```go
//go:embed all:frontend/dist
var frontendFS embed.FS
```

The Go binary serves the SPA from this embedded FS at `/`, with the REST API at `/api/v1`.
