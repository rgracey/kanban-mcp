# 0003 — SQLite Connection and Migrations

## Goal

Open a SQLite database, run embedded schema migrations on startup using `golang-migrate`, and expose a ready `*sql.DB` to the rest of the app.

## Dependencies

- Ticket 0001

## Commands

Add dependencies:
```sh
go get modernc.org/sqlite
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/sqlite
go get github.com/golang-migrate/migrate/v4/source/iofs
```

## Tasks

1. Create `internal/db/migrations/0001_init.up.sql`:

```sql
CREATE TABLE boards (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE epics (
    id          TEXT PRIMARY KEY,
    board_id    TEXT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE tickets (
    id          TEXT PRIMARY KEY,
    board_id    TEXT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    epic_id     TEXT REFERENCES epics(id) ON DELETE SET NULL,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'todo',
    priority    TEXT NOT NULL DEFAULT 'medium',
    created_at  TEXT NOT NULL,
    updated_at  TEXT NOT NULL
);

CREATE TABLE comments (
    id         TEXT PRIMARY KEY,
    ticket_id  TEXT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    body       TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
```

2. Create `internal/db/migrations/0001_init.down.sql`:
```sql
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS epics;
DROP TABLE IF EXISTS boards;
```

3. Create `internal/db/db.go`:
   - Embed the migrations directory with `//go:embed migrations/*.sql`.
   - `Open(path string) (*sql.DB, error)` — opens the SQLite file at `path`, enables WAL mode (`PRAGMA journal_mode=WAL`) and foreign key enforcement (`PRAGMA foreign_keys=ON`), then runs all pending up migrations before returning.
   - Use `source/iofs` and `database/sqlite` drivers for `golang-migrate`.

4. Call `db.Open(cfg.DBPath)` in `main.go`; `log.Fatal` on error.

## Acceptance Criteria

- `go build ./...` passes.
- Running the binary creates `kanban.db` (or the configured path) with the four tables present.
- Re-running the binary does not fail or re-apply already-applied migrations.
- Unit test in `internal/db/db_test.go`:
  - Calls `Open` with a temp file path.
  - Asserts all four tables exist via `SELECT name FROM sqlite_master WHERE type='table'`.
  - Asserts foreign keys are on via `PRAGMA foreign_keys`.
- Tests pass: `go test ./internal/db/...`
