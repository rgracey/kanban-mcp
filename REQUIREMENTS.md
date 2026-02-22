# kanban-mcp Requirements

## Overview

`kanban-mcp` is a lightweight, self-contained kanban board server distributed as a single Go binary. It serves two audiences:

1. **LLMs** — via the Model Context Protocol (MCP), allowing AI agents to read and manipulate boards, epics, tickets, and comments programmatically.
2. **Humans** — via a REST API and an embedded single-page application (SPA) for visual interaction.

---

## Architecture

### Single Binary

- The entire application (HTTP server, MCP server, frontend assets, database) is packaged into a single Go binary.
- Frontend assets are embedded via `go:embed`.
- No external runtime dependencies (no Docker, no separate database server, no Node process).

### Components

| Component | Description |
|---|---|
| MCP Server | Exposes MCP tools over `stdio` and HTTP/SSE |
| REST API | JSON API for all CRUD operations |
| SPA Frontend | Drag-and-drop kanban UI served from the binary |
| SQLite Database | Persistent storage via a local `.db` file |

---

## Tech Stack

### Backend (Go)

| Concern | Library | Notes |
|---|---|---|
| HTTP routing | `github.com/go-chi/chi/v5` | Lightweight, idiomatic, stdlib-compatible |
| MCP server | `github.com/mark3labs/mcp-go` | Supports stdio and HTTP/SSE transports |
| SQLite driver | `modernc.org/sqlite` | Pure Go, no CGo, cross-platform builds |
| DB migrations | `github.com/golang-migrate/migrate/v4` | Embedded SQL migration files |
| Structured logging | `log/slog` (stdlib) | JSON output, no extra dependency |

### Frontend

| Concern | Library | Notes |
|---|---|---|
| Framework | Svelte 5 + Vite | Minimal bundle, no virtual DOM overhead |
| Drag-and-drop | `svelte-dnd-action` | Purpose-built Svelte action, well maintained |
| Styling | Tailwind CSS v4 | Utility-first, zero runtime overhead |
| HTTP client | `fetch` (native) | No extra dependency needed |

### Build

- Frontend is built via `vite build`; the output `dist/` directory is embedded into the Go binary using `go:embed`.
- A `Makefile` (or equivalent) coordinates `vite build` then `go build` into a single step.
- The binary must cross-compile for `linux/amd64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64`.

---

## Configuration

Configuration is via **CLI flags and environment variables**. Flags take precedence over env vars.

| Flag | Env Var | Default | Description |
|---|---|---|---|
| `--port` | `KANBAN_PORT` | `8080` | HTTP port for REST API and SPA |
| `--db` | `KANBAN_DB` | `kanban.db` | Path to the SQLite database file |
| `--mcp-transport` | `KANBAN_MCP_TRANSPORT` | `stdio` | MCP transport: `stdio`, `http`, or `both` |
| `--mcp-port` | `KANBAN_MCP_PORT` | `8081` | Port for MCP HTTP/SSE transport (if enabled) |

---

## Data Model

### Board

Represents a top-level project or workstream.

| Field | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `name` | string | Required |
| `description` | string | Optional |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |

### Epic

A container for related tickets, scoped to a single board.

| Field | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `board_id` | UUID | FK → Board |
| `title` | string | Required |
| `description` | string | Optional |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |

### Ticket

A unit of work belonging to a board, optionally grouped under an epic.

| Field | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `board_id` | UUID | FK → Board |
| `epic_id` | UUID | FK → Epic, nullable |
| `title` | string | Required |
| `description` | string | Optional (markdown) |
| `status` | enum | See Statuses below |
| `priority` | enum | See Priorities below |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |

### Comment

A comment on a ticket.

| Field | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `ticket_id` | UUID | FK → Ticket |
| `body` | string | Required (markdown) |
| `created_at` | timestamp | |
| `updated_at` | timestamp | |

---

## Statuses

Fixed global statuses (applies to all boards):

| Value | Display Label |
|---|---|
| `todo` | To Do |
| `in_progress` | In Progress |
| `done` | Done |

---

## Priorities

| Value | Display Label |
|---|---|
| `low` | Low |
| `medium` | Medium |
| `high` | High |
| `critical` | Critical |

Default priority: `medium`.

---

## REST API

Base path: `/api/v1`

All requests and responses use `Content-Type: application/json`.

### Boards

| Method | Path | Description |
|---|---|---|
| `GET` | `/boards` | List all boards |
| `POST` | `/boards` | Create a board |
| `GET` | `/boards/:id` | Get a board |
| `PUT` | `/boards/:id` | Update a board |
| `DELETE` | `/boards/:id` | Delete a board (cascades to epics, tickets, comments) |
| `GET` | `/boards/:id/summary` | Board overview: ticket counts by status, open epics |

### Epics

| Method | Path | Description |
|---|---|---|
| `GET` | `/boards/:id/epics` | List epics for a board |
| `POST` | `/boards/:id/epics` | Create an epic |
| `GET` | `/epics/:id` | Get an epic |
| `PUT` | `/epics/:id` | Update an epic |
| `DELETE` | `/epics/:id` | Delete an epic (tickets become epic-less, not deleted) |

### Tickets

| Method | Path | Description |
|---|---|---|
| `GET` | `/boards/:id/tickets` | List tickets for a board (supports filters: `status`, `priority`, `epic_id`, `q` for keyword search) |
| `POST` | `/boards/:id/tickets` | Create a ticket |
| `GET` | `/tickets/:id` | Get a ticket |
| `PUT` | `/tickets/:id` | Update a ticket (including status/priority changes) |
| `DELETE` | `/tickets/:id` | Delete a ticket (cascades to comments) |

### Comments

| Method | Path | Description |
|---|---|---|
| `GET` | `/tickets/:id/comments` | List comments on a ticket |
| `POST` | `/tickets/:id/comments` | Add a comment |
| `PUT` | `/comments/:id` | Update a comment |
| `DELETE` | `/comments/:id` | Delete a comment |

---

## MCP Server

### Transports

- **stdio** — standard MCP transport; compatible with Claude Desktop and other MCP clients that spawn the binary as a subprocess.
- **HTTP/SSE** — MCP over HTTP using Server-Sent Events for streaming responses; useful for remote or networked MCP clients.
- Both transports can be active simultaneously.

### MCP Tools

The following tools are exposed to LLMs:

| Tool Name | Description |
|---|---|
| `list_boards` | List all boards |
| `create_board` | Create a new board |
| `update_board` | Update board name/description |
| `delete_board` | Delete a board |
| `get_board_summary` | Get ticket counts by status and open epics for a board |
| `list_epics` | List epics for a board |
| `create_epic` | Create a new epic on a board |
| `update_epic` | Update epic title/description |
| `delete_epic` | Delete an epic |
| `list_tickets` | List tickets for a board; supports filtering by `status`, `priority`, `epic_id`, and keyword search |
| `create_ticket` | Create a ticket on a board |
| `update_ticket` | Update ticket fields (title, description, status, priority, epic) |
| `delete_ticket` | Delete a ticket |
| `move_ticket` | Change the status/column of a ticket |
| `list_comments` | List comments on a ticket |
| `add_comment` | Add a comment to a ticket |
| `update_comment` | Update a comment's body |
| `delete_comment` | Delete a comment |

---

## Frontend (SPA)

- Embedded in the binary via `go:embed`; served at `/`
- **Drag-and-drop kanban board view**: columns represent statuses (`To Do`, `In Progress`, `Done`); tickets are cards that can be dragged between columns to change status
- Board switcher to navigate between multiple boards
- Epic grouping or filtering within a board view
- Ticket detail panel: view/edit title, description, priority, epic, and comments
- Create/edit/delete boards, epics, tickets, and comments from the UI
- Communicates exclusively with the REST API (`/api/v1`)

---

## Non-Functional Requirements

- **Single binary distribution**: `go build` produces one self-contained executable
- **No authentication**: designed for trusted local use; no login, tokens, or user accounts
- **Cross-platform**: must build and run on Linux, macOS, and Windows
- **Graceful shutdown**: in-flight requests are completed before the process exits
- **Structured logging**: JSON logs to stdout, with log level configurable via `--log-level` / `KANBAN_LOG_LEVEL` (`debug`, `info`, `warn`, `error`)

---

## Out of Scope (v1)

- User accounts, authentication, or authorisation
- Custom per-board columns/statuses
- File attachments
- Assignees
- Due dates
- Story points / estimation
- Notifications or webhooks
- Multi-node / distributed deployment
