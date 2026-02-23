package store

import (
	"context"
	"time"

	"github.com/rgracey/kanban-mcp/internal/models"
)

// ListNotes returns all notes for a ticket, ordered by created_at ascending.
func (s *SQLiteStore) ListNotes(ctx context.Context, ticketID string) ([]models.Note, error) {
	query := `SELECT id, ticket_id, body, created_at, updated_at FROM notes WHERE ticket_id = ? ORDER BY created_at ASC`
	rows, err := s.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]models.Note, 0)
	for rows.Next() {
		var n models.Note
		var createdAt, updatedAt string
		if err := rows.Scan(&n.ID, &n.TicketID, &n.Body, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if n.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
			return nil, err
		}
		if n.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// CreateNote creates a new note.
func (s *SQLiteStore) CreateNote(ctx context.Context, ticketID, body string) (models.Note, error) {
	id := newUUID()
	createdAt := timeToRFC3339(time.Now())
	query := `INSERT INTO notes (id, ticket_id, body, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, id, ticketID, body, createdAt, createdAt)
	if err != nil {
		return models.Note{}, err
	}
	n := models.Note{
		ID:        id,
		TicketID:  ticketID,
		Body:      body,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, _ = s.CreateTicketEvent(ctx, ticketID, models.EventCommented, "", nil)
	return n, nil
}

// GetNote returns a note by ID.
func (s *SQLiteStore) GetNote(ctx context.Context, id string) (models.Note, error) {
	query := `SELECT id, ticket_id, body, created_at, updated_at FROM notes WHERE id = ?`
	var n models.Note
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx, query, id).Scan(&n.ID, &n.TicketID, &n.Body, &createdAt, &updatedAt)
	if err != nil {
		return models.Note{}, err
	}
	if n.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
		return models.Note{}, err
	}
	if n.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
		return models.Note{}, err
	}
	return n, nil
}

// UpdateNote updates a note's body.
func (s *SQLiteStore) UpdateNote(ctx context.Context, id, body string) (models.Note, error) {
	updatedAt := timeToRFC3339(time.Now())
	query := `UPDATE notes SET body = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, body, updatedAt, id)
	if err != nil {
		return models.Note{}, err
	}
	n, err := s.GetNote(ctx, id)
	if err != nil {
		return models.Note{}, err
	}
	_, _ = s.CreateTicketEvent(ctx, n.TicketID, models.EventCommentEdited, "", nil)
	return n, nil
}

// DeleteNote deletes a note.
func (s *SQLiteStore) DeleteNote(ctx context.Context, id string) error {
	query := `DELETE FROM notes WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}
