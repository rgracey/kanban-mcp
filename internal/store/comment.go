package store

import (
	"context"
	"time"

	"github.com/rgracey/kanban-mcp/internal/models"
)

// ListComments returns all comments for a ticket, ordered by created_at ascending.
func (s *SQLiteStore) ListComments(ctx context.Context, ticketID string) ([]models.Comment, error) {
	query := `SELECT id, ticket_id, body, created_at, updated_at FROM comments WHERE ticket_id = ? ORDER BY created_at ASC`
	rows, err := s.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]models.Comment, 0)
	for rows.Next() {
		var c models.Comment
		var createdAt, updatedAt string
		if err := rows.Scan(&c.ID, &c.TicketID, &c.Body, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if c.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
			return nil, err
		}
		if c.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// CreateComment creates a new comment.
func (s *SQLiteStore) CreateComment(ctx context.Context, ticketID, body string) (models.Comment, error) {
	id := newUUID()
	createdAt := timeToRFC3339(time.Now())
	query := `INSERT INTO comments (id, ticket_id, body, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, id, ticketID, body, createdAt, createdAt)
	if err != nil {
		return models.Comment{}, err
	}
	return models.Comment{
		ID:        id,
		TicketID:  ticketID,
		Body:      body,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// GetComment returns a comment by ID.
func (s *SQLiteStore) GetComment(ctx context.Context, id string) (models.Comment, error) {
	query := `SELECT id, ticket_id, body, created_at, updated_at FROM comments WHERE id = ?`
	var c models.Comment
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.TicketID, &c.Body, &createdAt, &updatedAt)
	if err != nil {
		return models.Comment{}, err
	}
	if c.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
		return models.Comment{}, err
	}
	if c.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
		return models.Comment{}, err
	}
	return c, nil
}

// UpdateComment updates a comment's body.
func (s *SQLiteStore) UpdateComment(ctx context.Context, id, body string) (models.Comment, error) {
	updatedAt := timeToRFC3339(time.Now())
	query := `UPDATE comments SET body = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, body, updatedAt, id)
	if err != nil {
		return models.Comment{}, err
	}
	return s.GetComment(ctx, id)
}

// DeleteComment deletes a comment.
func (s *SQLiteStore) DeleteComment(ctx context.Context, id string) error {
	query := `DELETE FROM comments WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}
