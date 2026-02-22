# 0007 — Structured Logging

## Goal

Wire up `log/slog` with JSON output and a configurable log level. Replace any `fmt.Println` placeholders added in earlier tickets.

## Dependencies

- Ticket 0002 (config — provides `LogLevel`)

## Tasks

1. In `main.go`, after loading config, create a `slog.Logger` and set it as the default:

```go
var level slog.Level
switch cfg.LogLevel {
case "debug": level = slog.LevelDebug
case "warn":  level = slog.LevelWarn
case "error": level = slog.LevelError
default:      level = slog.LevelInfo
}
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
slog.SetDefault(logger)
```

2. Replace the `fmt.Println` config debug placeholder (ticket 0002) with:
```go
slog.Debug("config loaded", "config", cfg)
```

3. Log startup events at `info`:
   - `"database opened"` after `db.Open`.
   - `"HTTP server listening"` with `"port"` attr before `ListenAndServe`.

4. Log shutdown events (ticket 0008 covers graceful shutdown — add log lines there).

## Acceptance Criteria

- `go build ./...` passes.
- Running `./kanban-mcp` produces JSON log lines (not plain text).
- Running `./kanban-mcp --log-level debug` includes the config log line.
- Running `./kanban-mcp --log-level warn` suppresses `info` and `debug` lines.
- No `fmt.Println` / `fmt.Printf` calls remain in non-test code (verified by `grep -r "fmt.Print" --include="*.go" .` returning no results outside `_test.go` files).
