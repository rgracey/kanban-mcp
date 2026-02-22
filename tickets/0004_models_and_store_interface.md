# 0004 — Models and Store Interface

## Goal

Define the Go structs for all domain entities and a `Store` interface that the API and MCP layers depend on. No implementation yet — just the contract.

## Dependencies

- Ticket 0001

## Tasks

1. Create `internal/models/models.go` with the following types:

```go
package models

import "time"

type Status string
const (
    StatusTodo       Status = "todo"
    StatusInProgress Status = "in_progress"
    StatusDone       Status = "done"
)

type Priority string
const (
    PriorityLow      Priority = "low"
    PriorityMedium   Priority = "medium"
    PriorityHigh     Priority = "high"
    PriorityCritical Priority = "critical"
)

type Board struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Epic struct {
    ID          string    `json:"id"`
    BoardID     string    `json:"board_id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Ticket struct {
    ID          string    `json:"id"`
    BoardID     string    `json:"board_id"`
    EpicID      *string   `json:"epic_id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      Status    `json:"status"`
    Priority    Priority  `json:"priority"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Comment struct {
    ID        string    `json:"id"`
    TicketID  string    `json:"ticket_id"`
    Body      string    `json:"body"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type BoardSummary struct {
    BoardID      string         `json:"board_id"`
    TicketCounts map[string]int `json:"ticket_counts"`
    Epics        []EpicSummary  `json:"epics"`
}

type EpicSummary struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    TicketCount int    `json:"ticket_count"`
}

type TicketFilter struct {
    Status   *Status
    Priority *Priority
    EpicID   *string
    Query    *string // keyword search against title + description
}
```

2. Create `internal/store/store.go` defining the `Store` interface:

```go
package store

import (
    "context"
    "github.com/rgracey/kanban-mcp/internal/models"
)

type Store interface {
    // Boards
    ListBoards(ctx context.Context) ([]models.Board, error)
    CreateBoard(ctx context.Context, name, description string) (models.Board, error)
    GetBoard(ctx context.Context, id string) (models.Board, error)
    UpdateBoard(ctx context.Context, id string, name, description *string) (models.Board, error)
    DeleteBoard(ctx context.Context, id string) error
    GetBoardSummary(ctx context.Context, id string) (models.BoardSummary, error)

    // Epics
    ListEpics(ctx context.Context, boardID string) ([]models.Epic, error)
    CreateEpic(ctx context.Context, boardID, title, description string) (models.Epic, error)
    GetEpic(ctx context.Context, id string) (models.Epic, error)
    UpdateEpic(ctx context.Context, id string, title, description *string) (models.Epic, error)
    DeleteEpic(ctx context.Context, id string) error

    // Tickets
    ListTickets(ctx context.Context, boardID string, filter models.TicketFilter) ([]models.Ticket, error)
    CreateTicket(ctx context.Context, boardID string, t models.Ticket) (models.Ticket, error)
    GetTicket(ctx context.Context, id string) (models.Ticket, error)
    UpdateTicket(ctx context.Context, id string, fields map[string]any) (models.Ticket, error)
    DeleteTicket(ctx context.Context, id string) error

    // Comments
    ListComments(ctx context.Context, ticketID string) ([]models.Comment, error)
    CreateComment(ctx context.Context, ticketID, body string) (models.Comment, error)
    GetComment(ctx context.Context, id string) (models.Comment, error)
    UpdateComment(ctx context.Context, id, body string) (models.Comment, error)
    DeleteComment(ctx context.Context, id string) error
}
```

## Acceptance Criteria

- `go build ./...` passes.
- No implementation is required in this ticket — only the types and interface.
- `go vet ./...` passes with no warnings.
