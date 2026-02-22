# 0005 — SQLite Store Implementation

## Goal

Implement the `Store` interface against SQLite. This is the only persistence implementation.

## Dependencies

- Ticket 0003 (db open + migrations)
- Ticket 0004 (models + interface)

## Tasks

Create the following files, each implementing the relevant subset of `Store`:

- `internal/store/sqlite.go` — `SQLiteStore` struct holding `*sql.DB`; constructor `NewSQLiteStore(db *sql.DB) *SQLiteStore`; assert it implements `Store` via `var _ Store = (*SQLiteStore)(nil)`.
- `internal/store/board.go` — board methods.
- `internal/store/epic.go` — epic methods.
- `internal/store/ticket.go` — ticket methods.
- `internal/store/comment.go` — comment methods.

### Implementation notes

- All IDs are generated using `crypto/rand` UUID v4 (implement a small `newUUID() string` helper in `sqlite.go`).
- All timestamps stored and retrieved as RFC3339 UTC strings; marshal/unmarshal to/from `time.Time`.
- `UpdateBoard` / `UpdateEpic`: only update fields where the pointer argument is non-nil (partial update).
- `UpdateTicket`: takes `map[string]any` — build the `SET` clause dynamically from the provided keys. Valid keys: `title`, `description`, `status`, `priority`, `epic_id`. Reject unknown keys.
- `ListTickets` filter:
  - `status`, `priority`, `epic_id` — exact match `WHERE` clauses appended only when non-nil.
  - `query` — `WHERE (title LIKE '%?%' OR description LIKE '%?%')` (case-insensitive via SQLite's default `LIKE`).
- `GetBoardSummary`:
  - Ticket counts: `SELECT status, COUNT(*) FROM tickets WHERE board_id=? GROUP BY status`.
  - Epics: `SELECT e.id, e.title, COUNT(t.id) FROM epics e LEFT JOIN tickets t ON t.epic_id=e.id WHERE e.board_id=? GROUP BY e.id`.
- `DeleteBoard` cascades via SQLite `ON DELETE CASCADE` (foreign keys must be on — enforced in ticket 0003).
- `DeleteEpic`: set `epic_id = NULL` on orphaned tickets, then delete the epic. Do this in a transaction.
- `DeleteTicket` cascades via `ON DELETE CASCADE`.

## Acceptance Criteria

- `go build ./...` passes.
- Integration tests in `internal/store/store_test.go` using a temp SQLite file (call `db.Open` from ticket 0003):
  - **Boards**: create, get, list, update (partial), delete (verify cascade deletes tickets and comments).
  - **Epics**: create, get, list, update, delete (verify tickets become epic-less, not deleted).
  - **Tickets**: create, get, list with each filter type, update (partial via map), delete (verify cascade deletes comments).
  - **Comments**: create, get, list (ordered by `created_at` asc), update, delete.
  - **GetBoardSummary**: verify counts are correct after creating tickets with various statuses.
- Tests pass: `go test ./internal/store/...`
