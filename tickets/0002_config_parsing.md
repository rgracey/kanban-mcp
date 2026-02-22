# 0002 — Config Parsing

## Goal

Implement config loading from CLI flags and environment variables. Flags override env vars.

## Dependencies

- Ticket 0001

## Tasks

1. Add `internal/config/config.go` defining the `Config` struct and a `Load() Config` function.

### Config struct

```go
type Config struct {
    Port         int    // --port / KANBAN_PORT / default 8080
    DBPath       string // --db / KANBAN_DB / default "kanban.db"
    MCPTransport string // --mcp-transport / KANBAN_MCP_TRANSPORT / default "stdio"; values: "stdio","http","both"
    MCPPort      int    // --mcp-port / KANBAN_MCP_PORT / default 8081
    LogLevel     string // --log-level / KANBAN_LOG_LEVEL / default "info"; values: "debug","info","warn","error"
}
```

### `Load()` logic

- Parse flags using stdlib `flag` package.
- For each field: if the flag was not explicitly set, fall back to the corresponding env var; if that is also unset, use the default.
- Validate `MCPTransport` is one of `stdio`, `http`, `both`; validate `LogLevel` is one of `debug`, `info`, `warn`, `error`. Return an error (or `log.Fatal`) on invalid values.

2. Call `config.Load()` from `main.go` and print the resolved config at `debug` log level (logging ticket comes later — a `fmt.Println` placeholder is fine for now).

## Acceptance Criteria

- `go build ./...` passes.
- Running `./kanban-mcp --port 9090` resolves `Port` to `9090`.
- Running `KANBAN_PORT=9191 ./kanban-mcp` resolves `Port` to `9191`.
- A flag value takes precedence over the env var.
- An invalid `--mcp-transport` value causes the process to exit non-zero with an error message.
- Unit tests in `internal/config/config_test.go` cover:
  - Defaults are applied when no flags or env vars are set.
  - Env var is used when flag is absent.
  - Flag overrides env var.
  - Invalid `MCPTransport` returns an error / causes fatal.
- Tests pass: `go test ./internal/config/...`
