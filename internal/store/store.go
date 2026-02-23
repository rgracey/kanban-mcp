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
	GetBoardByName(ctx context.Context, name string) (models.Board, error)
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

	// Notes (agent scratchpad)
	ListNotes(ctx context.Context, ticketID string) ([]models.Note, error)
	CreateNote(ctx context.Context, ticketID, body string) (models.Note, error)
	GetNote(ctx context.Context, id string) (models.Note, error)
	UpdateNote(ctx context.Context, id, body string) (models.Note, error)
	DeleteNote(ctx context.Context, id string) error

	// Events
	ListTicketEvents(ctx context.Context, ticketID string) ([]models.TicketEvent, error)
	CreateTicketEvent(ctx context.Context, ticketID string, eventType models.TicketEventType, actor string, payload map[string]any) (models.TicketEvent, error)

	// Tasks
	ListTasks(ctx context.Context, ticketID string) ([]models.Task, error)
	CreateTask(ctx context.Context, ticketID, title string) (models.Task, error)
	UpdateTask(ctx context.Context, id string, title *string, done *bool) (models.Task, error)
	DeleteTask(ctx context.Context, id string) error

	// Ticket relations
	ListRelations(ctx context.Context, ticketID string) ([]models.TicketRelation, error)
	AddRelation(ctx context.Context, fromTicketID, toTicketID string, kind models.RelationKind) (models.TicketRelation, error)
	DeleteRelation(ctx context.Context, fromTicketID, toTicketID string, kind models.RelationKind) error

	// Bulk operations
	BulkCreateTickets(ctx context.Context, boardID string, tickets []models.Ticket) ([]models.Ticket, error)

	// Agent-oriented helpers
	// ReadyTickets returns unblocked todo tickets for a board ordered by priority descending.
	// A ticket is "ready" if it has no incoming blocking relations (i.e. nothing blocks it).
	ReadyTickets(ctx context.Context, boardID string) ([]models.Ticket, error)
	// BoardContext returns a full snapshot of a board: board metadata, epics, and all tickets
	// with their tasks and relations embedded. Designed for LLM agents needing full context.
	BoardContext(ctx context.Context, boardID string) (models.BoardContext, error)
}
