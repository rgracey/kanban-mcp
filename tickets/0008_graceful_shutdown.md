# 0008 — Graceful Shutdown

## Goal

On `SIGINT` or `SIGTERM`, stop accepting new connections, wait for in-flight requests to complete, then exit cleanly.

## Dependencies

- Ticket 0006 (HTTP server running)
- Ticket 0007 (logging)

## Tasks

Replace the bare `http.ListenAndServe` call in `main.go` with:

```go
srv := &http.Server{
    Addr:    fmt.Sprintf(":%d", cfg.Port),
    Handler: router,
}

go func() {
    slog.Info("HTTP server listening", "port", cfg.Port)
    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        slog.Error("server error", "err", err)
        os.Exit(1)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

slog.Info("shutting down")
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    slog.Error("shutdown error", "err", err)
}
slog.Info("shutdown complete")
```

Close the database connection after server shutdown:
```go
db.Close()
```

## Acceptance Criteria

- `go build ./...` passes.
- Sending `SIGINT` (`Ctrl+C`) to the running binary logs `"shutting down"` then `"shutdown complete"` and exits 0.
- An in-flight request that started before the signal is allowed to complete (verify manually: add a `time.Sleep` in a handler temporarily, send signal, confirm response still arrives).
- Process exits within 11 seconds even if a request hangs (the 10s context timeout forces it).
