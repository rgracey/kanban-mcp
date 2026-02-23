# kanban-mcp

A self-contained kanban board server that works for both humans and AI agents.

It ships as a single Go binary with no external dependencies: SQLite is embedded, the web UI is embedded, and the MCP server runs alongside the REST API on the same process. Spin it up and everything is immediately available.

---

## What it does

**For humans** — a visual kanban board accessible in any browser. Boards, epics, tickets with markdown descriptions, checklists, comments, assignees, priority levels, and a per-ticket audit history showing every change over time.

**For AI agents** — a full [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server. Claude (and any other MCP-compatible client) can create boards, manage tickets, write comments, check off tasks, and inspect history — all through natural conversation. The same data store backs both interfaces, so changes made by an agent are immediately visible in the UI, and vice versa.

---

## Why

Most project management tools are either too heavy to self-host or have no machine-readable interface. kanban-mcp is designed to be the task layer that an AI coding agent actually uses — lightweight enough to run locally alongside your editor, with an MCP interface that exposes the full ticket lifecycle rather than a read-only view.

---

## Features

- **Boards, epics, and tickets** with status (`todo` / `in progress` / `done`) and priority (`low` / `medium` / `high` / `critical`)
- **Checklists** — per-ticket task lists with progress tracking
- **Comments** with edit support
- **Assignees** — free-text assignee field per ticket
- **Markdown** rendering in ticket descriptions
- **Audit history** — every create, edit, move, comment, and task change is recorded and shown in a timeline on the ticket
- **Real-time updates** — the UI reacts to changes via Server-Sent Events (no polling, no refresh required)
- **Dark mode** — user toggle, persisted in localStorage
- **MCP server** — 6 action-dispatched tools covering the full data model (see [MCP tools](#mcp-tools) below)
- **Zero external dependencies** — SQLite, frontend assets, and MCP all embedded in one binary

---

## Quick start

### Prerequisites

- [Go](https://go.dev/dl/) 1.22 or later

### Download and run

```sh
go install github.com/rgracey/kanban-mcp@latest
kanban-mcp
```

Or build from source:

```sh
git clone https://github.com/rgracey/kanban-mcp
cd kanban-mcp
go build -o kanban-mcp .
./kanban-mcp
```

The web UI is available at `http://localhost:8080`.

---

## MCP setup

### Claude Desktop (recommended)

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "kanban": {
      "command": "/path/to/kanban-mcp"
    }
  }
}
```

The binary defaults to stdio transport, which is what Claude Desktop expects. No extra flags needed.

### Claude Code

```sh
claude mcp add kanban /path/to/kanban-mcp
```

### HTTP transport

If you need the MCP server reachable over HTTP (e.g. for a remote agent or to use with an MCP inspector):

```sh
kanban-mcp --mcp-transport http   # MCP at http://localhost:8080/mcp
kanban-mcp --mcp-transport both   # HTTP + stdio simultaneously
```

---

## MCP tools

All tools use an action-dispatch pattern: one tool per resource, with an `action` parameter selecting the operation.

| Tool | Actions | Key parameters |
|---|---|---|
| `board` | `list`, `get`, `create`, `update`, `delete`, `summary` | `id`, `name`, `description` |
| `epic` | `list`, `get`, `create`, `update`, `delete` | `id`, `board_id`, `title`, `description` |
| `ticket` | `list`, `get`, `create`, `update`, `delete`, `move` | `id`, `board_id`, `title`, `description`, `status`, `priority`, `epic_id`, `assignee`, filter/sort params, `include_comments`, `include_history` |
| `task` | `list`, `create`, `update`, `delete` | `id`, `ticket_id`, `title`, `done` |
| `comment` | `list`, `add`, `update`, `delete` | `id`, `ticket_id`, `body` |
| `ticket_history` | _(single action)_ | `ticket_id` |

Example — ask Claude: *"Create a board called 'Backend Rewrite', add an epic for authentication, and create three tickets under it."*

---

## Configuration

Flags take precedence over environment variables.

| Flag | Env var | Default | Description |
|---|---|---|---|
| `--port` | `KANBAN_PORT` | `8080` | HTTP listen port |
| `--db` | `KANBAN_DB` | `kanban.db` | Path to the SQLite database file |
| `--mcp-transport` | `KANBAN_MCP_TRANSPORT` | `stdio` | `stdio`, `http`, or `both` |
| `--log-level` | `KANBAN_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |

The database file is created automatically on first run. Migrations are applied automatically on startup.

---

## Development

### Prerequisites

- Go 1.22+
- Node.js 20+

### Running locally

Run the backend and the Vite dev server in separate terminals. The Vite dev server proxies `/api` requests to the Go backend.

```sh
# Terminal 1 — Go backend
go run .

# Terminal 2 — frontend with hot reload
cd frontend && npm install && npm run dev
```

Frontend: `http://localhost:5173`  
Backend API: `http://localhost:8080`

### Tests

```sh
go test ./...
```

### Building the frontend

The compiled frontend (`frontend/dist/`) is committed to the repository so that `go build` works without Node. If you change frontend source files, rebuild and commit the dist:

```sh
cd frontend && npm run build
git add frontend/dist/
```

### Building the binary

```sh
go build -o kanban-mcp .
```

---

## REST API

Base path: `/api/v1`. All responses are JSON. All IDs are UUID v4. Timestamps are RFC3339.

<details>
<summary>Boards</summary>

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/boards` | List all boards |
| `POST` | `/api/v1/boards` | Create a board (`name` required) |
| `GET` | `/api/v1/boards/:id` | Get a board |
| `PUT` | `/api/v1/boards/:id` | Update a board (partial) |
| `DELETE` | `/api/v1/boards/:id` | Delete a board (cascades to epics, tickets, comments) |
| `GET` | `/api/v1/boards/:id/summary` | Ticket counts by status + epic breakdown |

</details>

<details>
<summary>Epics</summary>

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/boards/:id/epics` | List epics for a board |
| `POST` | `/api/v1/boards/:id/epics` | Create an epic (`title` required) |
| `GET` | `/api/v1/epics/:id` | Get an epic |
| `PUT` | `/api/v1/epics/:id` | Update an epic (partial) |
| `DELETE` | `/api/v1/epics/:id` | Delete an epic (tickets are kept, `epic_id` set to null) |

</details>

<details>
<summary>Tickets</summary>

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/boards/:id/tickets` | List tickets (filter: `status`, `priority`, `epic_id`, `q`, `sort_by`, `sort_order`) |
| `POST` | `/api/v1/boards/:id/tickets` | Create a ticket (`title` required; defaults: `status=todo`, `priority=medium`) |
| `GET` | `/api/v1/tickets/:id` | Get a ticket |
| `PUT` | `/api/v1/tickets/:id` | Update a ticket (any subset of fields; `epic_id: null` clears the epic) |
| `DELETE` | `/api/v1/tickets/:id` | Delete a ticket (cascades to comments and tasks) |

</details>

<details>
<summary>Tasks</summary>

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/tickets/:id/tasks` | List tasks for a ticket |
| `POST` | `/api/v1/tickets/:id/tasks` | Create a task (`title` required) |
| `PUT` | `/api/v1/tasks/:id` | Update a task (`title`, `done`) |
| `DELETE` | `/api/v1/tasks/:id` | Delete a task |

</details>

<details>
<summary>Comments</summary>

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/tickets/:id/comments` | List comments (ordered by `created_at` asc) |
| `POST` | `/api/v1/tickets/:id/comments` | Add a comment (`body` required) |
| `PUT` | `/api/v1/comments/:id` | Update a comment (`body` required) |
| `DELETE` | `/api/v1/comments/:id` | Delete a comment |

</details>

<details>
<summary>Audit history</summary>

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/tickets/:id/events` | List audit events for a ticket (ordered by `created_at` asc) |

</details>

<details>
<summary>Real-time events (SSE)</summary>

`GET /api/v1/boards/:id/events` — Server-Sent Events stream. Emits an event whenever a ticket on the board is created, updated, or deleted. The UI uses this to refresh without polling.

</details>

---

## Project structure

```
kanban-mcp/
├── main.go                  # Entry point: wires DB, store, MCP server, HTTP server
├── frontend/
│   ├── src/                 # Svelte 5 + TypeScript source
│   └── dist/                # Compiled assets (committed; embedded by go:embed)
└── internal/
    ├── api/                 # HTTP handlers, SSE hub, router
    ├── config/              # CLI flag + env var parsing
    ├── db/                  # SQLite open + migrations
    ├── mcp/                 # MCP server and tool registration
    ├── models/              # Shared data types
    └── store/               # SQLite data access layer
```

---

## License

MIT
