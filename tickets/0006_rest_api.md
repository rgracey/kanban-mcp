# 0006 — REST API

## Goal

Implement all REST API endpoints using `chi`. Handlers call the `Store` interface — no direct DB access.

## Dependencies

- Ticket 0004 (models + Store interface)
- Ticket 0005 (store implementation, needed for integration tests)

## Commands

```sh
go get github.com/go-chi/chi/v5
```

## Tasks

### `internal/api/router.go`

- Create `NewRouter(s store.Store) http.Handler`.
- Mount routes under `/api/v1` using a chi sub-router.
- Middleware: `chi/middleware.Logger`, `chi/middleware.Recoverer`, `chi/middleware.RequestID`.
- Serve the embedded SPA for all non-API routes (handled in ticket 0009 — leave a `// TODO: embed SPA` comment for now).

### Route handlers

Implement handlers in the files below. Each handler must:
- Decode JSON request body where applicable.
- Validate required fields; return `400` with `{"error": "..."}` on failure.
- Return `404` with `{"error": "not found"}` when the store returns no row.
- Return `500` with `{"error": "internal server error"}` on unexpected errors (do not leak internal detail).
- Set `Content-Type: application/json` on all responses.

See `docs/api_contracts.md` for exact request/response shapes.

| File | Routes |
|---|---|
| `internal/api/boards.go` | `GET /boards`, `POST /boards`, `GET /boards/{id}`, `PUT /boards/{id}`, `DELETE /boards/{id}`, `GET /boards/{id}/summary` |
| `internal/api/epics.go` | `GET /boards/{id}/epics`, `POST /boards/{id}/epics`, `GET /epics/{id}`, `PUT /epics/{id}`, `DELETE /epics/{id}` |
| `internal/api/tickets.go` | `GET /boards/{id}/tickets`, `POST /boards/{id}/tickets`, `GET /tickets/{id}`, `PUT /tickets/{id}`, `DELETE /tickets/{id}` |
| `internal/api/comments.go` | `GET /tickets/{id}/comments`, `POST /tickets/{id}/comments`, `PUT /comments/{id}`, `DELETE /comments/{id}` |

### Validation rules

- `Board.Name`: non-empty string.
- `Epic.Title`: non-empty string.
- `Ticket.Title`: non-empty string.
- `Ticket.Status`: must be one of `todo`, `in_progress`, `done` if provided.
- `Ticket.Priority`: must be one of `low`, `medium`, `high`, `critical` if provided.
- `Comment.Body`: non-empty string.

### Wire into `main.go`

```go
store := store.NewSQLiteStore(db)
router := api.NewRouter(store)
http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router)
```

## Acceptance Criteria

- `go build ./...` passes.
- Integration tests in `internal/api/` using `net/http/httptest`:
  - Each endpoint has at minimum: happy path, 404 on unknown ID, 400 on missing required field.
  - `DELETE /boards/:id` — subsequent `GET` returns 404; tickets and comments are gone.
  - `DELETE /epics/:id` — tickets previously in that epic still exist with `epic_id: null`.
  - `GET /boards/:id/tickets?status=todo` — only returns `todo` tickets.
  - `GET /boards/:id/tickets?q=keyword` — returns tickets matching keyword in title or description.
  - `GET /boards/:id/summary` — counts match actual ticket state.
- Tests pass: `go test ./internal/api/...`
