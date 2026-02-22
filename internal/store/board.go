package store

import (
	"context"
	"time"

	"github.com/rgracey/kanban-mcp/internal/models"
)

// ListBoards returns all boards.
func (s *SQLiteStore) ListBoards(ctx context.Context) ([]models.Board, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM boards ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	boards := make([]models.Board, 0)
	for rows.Next() {
		var b models.Board
		var createdAt, updatedAt string
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if b.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
			return nil, err
		}
		if b.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
			return nil, err
		}
		boards = append(boards, b)
	}
	return boards, rows.Err()
}

// CreateBoard creates a new board.
func (s *SQLiteStore) CreateBoard(ctx context.Context, name, description string) (models.Board, error) {
	id := newUUID()
	createdAt := timeToRFC3339(time.Now())
	query := `INSERT INTO boards (id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, id, name, description, createdAt, createdAt)
	if err != nil {
		return models.Board{}, err
	}
	return models.Board{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetBoard returns a board by ID.
func (s *SQLiteStore) GetBoard(ctx context.Context, id string) (models.Board, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM boards WHERE id = ?`
	var b models.Board
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx, query, id).Scan(&b.ID, &b.Name, &b.Description, &createdAt, &updatedAt)
	if err != nil {
		return models.Board{}, err
	}
	if b.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
		return models.Board{}, err
	}
	if b.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
		return models.Board{}, err
	}
	return b, nil
}

// UpdateBoard updates a board's name and description (partial update).
func (s *SQLiteStore) UpdateBoard(ctx context.Context, id string, name, description *string) (models.Board, error) {
	// Build dynamic SET clause
	var setClauses []string
	var args []interface{}

	if name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *name)
	}
	if description != nil {
		setClauses = append(setClauses, "description = ?")
		args = append(args, *description)
	}

	if len(setClauses) == 0 {
		// No fields to update, just get the current board
		return s.GetBoard(ctx, id)
	}

	updatedAt := timeToRFC3339(time.Now())
	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, updatedAt)
	args = append(args, id)

	query := `UPDATE boards SET ` + joinWithComma(setClauses) + ` WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return models.Board{}, err
	}
	return s.GetBoard(ctx, id)
}

// DeleteBoard deletes a board. Cascade deletes tickets and comments via FK.
func (s *SQLiteStore) DeleteBoard(ctx context.Context, id string) error {
	query := `DELETE FROM boards WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// GetBoardSummary returns ticket counts by status and epics with ticket counts.
func (s *SQLiteStore) GetBoardSummary(ctx context.Context, id string) (models.BoardSummary, error) {
	summary := models.BoardSummary{
		BoardID:      id,
		TicketCounts: make(map[string]int),
	}

	// Get ticket counts by status
	query := `SELECT status, COUNT(*) FROM tickets WHERE board_id = ? GROUP BY status`
	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return summary, err
		}
		summary.TicketCounts[status] = count
	}
	if err := rows.Err(); err != nil {
		return summary, err
	}

	// Get epics with ticket counts
	query = `SELECT e.id, e.title, COUNT(t.id) FROM epics e LEFT JOIN tickets t ON t.epic_id = e.id WHERE e.board_id = ? GROUP BY e.id`
	rows, err = s.db.QueryContext(ctx, query, id)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.EpicSummary
		if err := rows.Scan(&e.ID, &e.Title, &e.TicketCount); err != nil {
			return summary, err
		}
		summary.Epics = append(summary.Epics, e)
	}
	if err := rows.Err(); err != nil {
		return summary, err
	}

	return summary, nil
}

// joinWithComma joins string slices with commas.
func joinWithComma(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += ", " + parts[i]
	}
	return result
}
