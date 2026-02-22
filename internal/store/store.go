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
