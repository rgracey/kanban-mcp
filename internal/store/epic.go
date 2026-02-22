package store

import (
	"context"
	"time"

	"github.com/rgracey/kanban-mcp/internal/models"
)

// ListEpics returns all epics for a board.
func (s *SQLiteStore) ListEpics(ctx context.Context, boardID string) ([]models.Epic, error) {
	query := `SELECT id, board_id, title, description, created_at, updated_at FROM epics WHERE board_id = ? ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query, boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	epics := make([]models.Epic, 0)
	for rows.Next() {
		var e models.Epic
		var createdAt, updatedAt string
		if err := rows.Scan(&e.ID, &e.BoardID, &e.Title, &e.Description, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if e.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
			return nil, err
		}
		if e.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
			return nil, err
		}
		epics = append(epics, e)
	}
	return epics, rows.Err()
}

// CreateEpic creates a new epic.
func (s *SQLiteStore) CreateEpic(ctx context.Context, boardID, title, description string) (models.Epic, error) {
	id := newUUID()
	createdAt := timeToRFC3339(time.Now())
	query := `INSERT INTO epics (id, board_id, title, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, id, boardID, title, description, createdAt, createdAt)
	if err != nil {
		return models.Epic{}, err
	}
	return models.Epic{
		ID:          id,
		BoardID:     boardID,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetEpic returns an epic by ID.
func (s *SQLiteStore) GetEpic(ctx context.Context, id string) (models.Epic, error) {
	query := `SELECT id, board_id, title, description, created_at, updated_at FROM epics WHERE id = ?`
	var e models.Epic
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx, query, id).Scan(&e.ID, &e.BoardID, &e.Title, &e.Description, &createdAt, &updatedAt)
	if err != nil {
		return models.Epic{}, err
	}
	if e.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
		return models.Epic{}, err
	}
	if e.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
		return models.Epic{}, err
	}
	return e, nil
}

// UpdateEpic updates an epic's title and description (partial update).
func (s *SQLiteStore) UpdateEpic(ctx context.Context, id string, title, description *string) (models.Epic, error) {
	var setClauses []string
	var args []interface{}

	if title != nil {
		setClauses = append(setClauses, "title = ?")
		args = append(args, *title)
	}
	if description != nil {
		setClauses = append(setClauses, "description = ?")
		args = append(args, *description)
	}

	if len(setClauses) == 0 {
		// No fields to update, just get the current epic
		return s.GetEpic(ctx, id)
	}

	updatedAt := timeToRFC3339(time.Now())
	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, updatedAt)
	args = append(args, id)

	query := `UPDATE epics SET ` + joinWithComma(setClauses) + ` WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return models.Epic{}, err
	}
	return s.GetEpic(ctx, id)
}

// DeleteEpic deletes an epic. Orphaned tickets have epic_id set to NULL.
func (s *SQLiteStore) DeleteEpic(ctx context.Context, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Set epic_id to NULL on all tickets that reference this epic
	query := `UPDATE tickets SET epic_id = NULL WHERE epic_id = ?`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Delete the epic
	query = `DELETE FROM epics WHERE id = ?`
	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}
