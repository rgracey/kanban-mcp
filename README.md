# kanban-mcp

A lightweight, self-contained kanban board server distributed as a single Go binary. It serves two audiences:

- **LLMs** — via the Model Context Protocol (MCP), allowing AI agents to read and manipulate boards, epics, tickets, and comments.
- **Humans** — via a REST API and an embedded single-page application for visual drag-and-drop interaction.

---

## Prerequisites

- **Go** 1.22 or later
- **Node.js** 20 or later

---

## Build

The built frontend (`frontend/dist/`) is committed to the repo so `go build` works without any Node tooling. Just:

```sh
go build -o kanban-mcp .
```

If you've changed frontend source files, rebuild the frontend first then commit the updated `dist/`:

```sh
cd frontend && npm install && npm run build && cd ..
git add frontend/dist/
# then go build -o kanban-mcp .
```

---

## Run

```sh
./kanban-mcp
```

With options:

```sh
./kanban-mcp --port 9090 --db /data/kanban.db --mcp-transport both --log-level debug
```

The SPA is served at `http://localhost:8080` (or whichever port you configure).

---

## MCP (stdio)

`kanban-mcp` defaults to stdio transport, making it compatible with Claude Desktop and any MCP client that spawns the binary as a subprocess.

Add it to your Claude Desktop config (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "kanban": {
      "command": "/path/to/kanban-mcp"
    }
  }
}
```

No extra arguments are needed — stdio is the default transport.

To enable the HTTP transport (Streamable HTTP, MCP spec 2025-03-26), the MCP endpoint is mounted at `/mcp` on the same port as the REST API:

```sh
./kanban-mcp --mcp-transport http   # MCP at http://localhost:8080/mcp
./kanban-mcp --mcp-transport both   # MCP HTTP + stdio simultaneously
```

---

## Configuration

Configuration is via CLI flags and environment variables. Flags take precedence over env vars.

| Flag | Env Var | Default | Description |
|---|---|---|---|
| `--port` | `KANBAN_PORT` | `8080` | HTTP port for REST API, SPA, and MCP |
| `--db` | `KANBAN_DB` | `kanban.db` | Path to the SQLite database file |
| `--mcp-transport` | `KANBAN_MCP_TRANSPORT` | `stdio` | MCP transport: `stdio`, `http`, or `both` |
| `--log-level` | `KANBAN_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |

---

## Development

Run the Go server and the Vite dev server in separate terminals. Vite proxies `/api` requests to the Go backend automatically.

```sh
# Terminal 1 — Go backend (REST API + MCP)
go run .

# Terminal 2 — Vite dev server (hot-reload frontend)
cd frontend && npm run dev
```

The frontend dev server runs on `http://localhost:5173` by default and proxies `/api` to `http://localhost:8080`.
