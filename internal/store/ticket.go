package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/rgracey/kanban-mcp/internal/models"
)

// ListTickets returns tickets for a board, filtered by the provided filter.
func (s *SQLiteStore) ListTickets(ctx context.Context, boardID string, filter models.TicketFilter) ([]models.Ticket, error) {
	query := `SELECT id, board_id, epic_id, title, description, status, priority, assignee, created_at, updated_at FROM tickets WHERE board_id = ?`
	args := []interface{}{boardID}

	// Add filters
	if filter.Status != nil {
		query += ` AND status = ?`
		args = append(args, *filter.Status)
	}
	if filter.Priority != nil {
		query += ` AND priority = ?`
		args = append(args, *filter.Priority)
	}
	if filter.EpicID != nil {
		query += ` AND epic_id = ?`
		args = append(args, *filter.EpicID)
	}
	if filter.Query != nil {
		query += ` AND (title LIKE ? OR description LIKE ?)`
		searchPattern := `%` + *filter.Query + `%`
		args = append(args, searchPattern, searchPattern)
	}

	// Sorting
	sortOrder := "DESC"
	if filter.SortOrder != nil && (*filter.SortOrder == "asc" || *filter.SortOrder == "ASC") {
		sortOrder = "ASC"
	}
	if filter.SortBy != nil && *filter.SortBy == "priority" {
		// Map priority strings to numeric weight for ordering
		query += ` ORDER BY CASE priority WHEN 'critical' THEN 4 WHEN 'high' THEN 3 WHEN 'medium' THEN 2 WHEN 'low' THEN 1 ELSE 0 END ` + sortOrder
	} else {
		query += ` ORDER BY created_at ` + sortOrder
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := make([]models.Ticket, 0)
	for rows.Next() {
		var t models.Ticket
		var epicID sql.NullString
		var createdAt, updatedAt string

		if err := rows.Scan(&t.ID, &t.BoardID, &epicID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Assignee, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		if epicID.Valid {
			t.EpicID = &epicID.String
		}

		if t.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
			return nil, err
		}
		if t.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
			return nil, err
		}

		tickets = append(tickets, t)
	}

	return tickets, rows.Err()
}

// CreateTicket creates a new ticket.
func (s *SQLiteStore) CreateTicket(ctx context.Context, boardID string, t models.Ticket) (models.Ticket, error) {
	id := newUUID()
	createdAt := timeToRFC3339(time.Now())

	query := `INSERT INTO tickets (id, board_id, epic_id, title, description, status, priority, assignee, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	epicID := sql.NullString{}
	if t.EpicID != nil {
		epicID = sql.NullString{String: *t.EpicID, Valid: true}
	}

	_, err := s.db.ExecContext(ctx, query, id, boardID, epicID, t.Title, t.Description, t.Status, t.Priority, t.Assignee, createdAt, createdAt)
	if err != nil {
		return models.Ticket{}, err
	}

	created := models.Ticket{
		ID:          id,
		BoardID:     boardID,
		EpicID:      t.EpicID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		Assignee:    t.Assignee,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Emit audit event (best-effort; do not fail the create on event error)
	_, _ = s.CreateTicketEvent(ctx, id, models.EventCreated, "", map[string]any{
		"title":    created.Title,
		"status":   string(created.Status),
		"priority": string(created.Priority),
	})

	return created, nil
}

// GetTicket returns a ticket by ID.
func (s *SQLiteStore) GetTicket(ctx context.Context, id string) (models.Ticket, error) {
	query := `SELECT id, board_id, epic_id, title, description, status, priority, assignee, created_at, updated_at FROM tickets WHERE id = ?`
	var t models.Ticket
	var epicID sql.NullString
	var createdAt, updatedAt string

	err := s.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.BoardID, &epicID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Assignee, &createdAt, &updatedAt)
	if err != nil {
		return models.Ticket{}, err
	}

	if epicID.Valid {
		t.EpicID = &epicID.String
	}

	if t.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
		return models.Ticket{}, err
	}
	if t.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
		return models.Ticket{}, err
	}

	return t, nil
}

// UpdateTicket updates a ticket with partial fields.
func (s *SQLiteStore) UpdateTicket(ctx context.Context, id string, fields map[string]any) (models.Ticket, error) {
	// Fetch current state for event diffing
	before, _ := s.GetTicket(ctx, id)

	// Validate known keys
	validKeys := map[string]bool{
		"title":       true,
		"description": true,
		"status":      true,
		"priority":    true,
		"epic_id":     true,
		"assignee":    true,
	}

	for key := range fields {
		if !validKeys[key] {
			return models.Ticket{}, fmt.Errorf("invalid field key: %s", key)
		}
	}

	var setClauses []string
	var args []interface{}

	if val, ok := fields["title"]; ok {
		setClauses = append(setClauses, "title = ?")
		args = append(args, val)
	}
	if val, ok := fields["description"]; ok {
		setClauses = append(setClauses, "description = ?")
		args = append(args, val)
	}
	if val, ok := fields["status"]; ok {
		setClauses = append(setClauses, "status = ?")
		args = append(args, val)
	}
	if val, ok := fields["priority"]; ok {
		setClauses = append(setClauses, "priority = ?")
		args = append(args, val)
	}
	if val, ok := fields["epic_id"]; ok {
		setClauses = append(setClauses, "epic_id = ?")
		if val == nil {
			args = append(args, nil)
		} else {
			args = append(args, val)
		}
	}
	if val, ok := fields["assignee"]; ok {
		setClauses = append(setClauses, "assignee = ?")
		args = append(args, val)
	}

	updatedAt := timeToRFC3339(time.Now())
	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, updatedAt)
	args = append(args, id)

	query := `UPDATE tickets SET ` + joinWithComma(setClauses) + ` WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return models.Ticket{}, err
	}

	updated, err := s.GetTicket(ctx, id)
	if err != nil {
		return models.Ticket{}, err
	}

	// Emit audit event (best-effort)
	if newStatus, ok := fields["status"].(string); ok && string(before.Status) != newStatus {
		_, _ = s.CreateTicketEvent(ctx, id, models.EventMoved, "", map[string]any{
			"from": string(before.Status),
			"to":   newStatus,
		})
	} else {
		payload := map[string]any{}
		for k, v := range fields {
			if v != nil {
				payload[k] = v
			}
		}
		if len(payload) > 0 {
			_, _ = s.CreateTicketEvent(ctx, id, models.EventEdited, "", payload)
		}
	}

	return updated, nil
}

// DeleteTicket deletes a ticket. Cascade deletes comments via FK.
func (s *SQLiteStore) DeleteTicket(ctx context.Context, id string) error {
	query := `DELETE FROM tickets WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}
