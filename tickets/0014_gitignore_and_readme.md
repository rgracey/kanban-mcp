# 0014 — .gitignore and README

## Goal

Add a `.gitignore` appropriate for a Go + Node project and a minimal `README.md` covering build and run instructions.

## Dependencies

None (can be done at any point).

## Tasks

### `.gitignore`

```
# Go
kanban-mcp
kanban-mcp.exe
*.db

# Frontend build output
frontend/dist/
frontend/node_modules/

# Editor
.DS_Store
.idea/
.vscode/
```

### `README.md`

Must cover:

1. **Prerequisites** — Go 1.22+, Node 20+.
2. **Build**:
   ```sh
   cd frontend && npm install && npm run build && cd ..
   go build -o kanban-mcp ./...
   ```
3. **Run**:
   ```sh
   ./kanban-mcp
   ```
   With options:
   ```sh
   ./kanban-mcp --port 9090 --db /data/kanban.db --mcp-transport both --log-level debug
   ```
4. **MCP (stdio)** — how to add the binary as an MCP server in Claude Desktop (`command: /path/to/kanban-mcp`, no args needed for stdio default).
5. **Configuration table** — copy the table from `REQUIREMENTS.md § Configuration`.
6. **Development** — run the Go server, then in a separate terminal run the Vite dev server:
   ```sh
   # Terminal 1
   go run ./...

   # Terminal 2
   cd frontend && npm run dev
   ```
   The Vite dev server proxies `/api` to `http://localhost:8080`.

## Acceptance Criteria

- `git status` does not show `frontend/dist/`, `frontend/node_modules/`, or any `.db` file as untracked after build.
- README contains all six sections listed above.
