package store

import (
	"context"
	"time"

	"github.com/rgracey/kanban-mcp/internal/models"
)

func (s *SQLiteStore) ListTasks(ctx context.Context, ticketID string) ([]models.Task, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, ticket_id, title, done, position, created_at, updated_at FROM tasks WHERE ticket_id = ? ORDER BY position ASC, created_at ASC`,
		ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]models.Task, 0)
	for rows.Next() {
		var t models.Task
		var done int
		var createdAt, updatedAt string

		if err := rows.Scan(&t.ID, &t.TicketID, &t.Title, &done, &t.Position, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		t.Done = done == 1
		if t.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
			return nil, err
		}
		if t.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *SQLiteStore) CreateTask(ctx context.Context, ticketID, title string) (models.Task, error) {
	id := newUUID()
	now := timeToRFC3339(time.Now())

	// Position = count of existing tasks so new task goes to end
	var count int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tasks WHERE ticket_id = ?`, ticketID).Scan(&count)

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO tasks (id, ticket_id, title, done, position, created_at, updated_at) VALUES (?, ?, ?, 0, ?, ?, ?)`,
		id, ticketID, title, count, now, now,
	)
	if err != nil {
		return models.Task{}, err
	}

	_, _ = s.CreateTicketEvent(ctx, ticketID, models.EventTaskAdded, "", map[string]any{
		"task_title": title,
	})

	t, _ := rfc3339ToTime(now)
	return models.Task{
		ID:        id,
		TicketID:  ticketID,
		Title:     title,
		Done:      false,
		Position:  count,
		CreatedAt: t,
		UpdatedAt: t,
	}, nil
}

func (s *SQLiteStore) UpdateTask(ctx context.Context, id string, title *string, done *bool) (models.Task, error) {
	now := timeToRFC3339(time.Now())

	if title != nil {
		if _, err := s.db.ExecContext(ctx, `UPDATE tasks SET title = ?, updated_at = ? WHERE id = ?`, *title, now, id); err != nil {
			return models.Task{}, err
		}
	}
	if done != nil {
		doneInt := 0
		if *done {
			doneInt = 1
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE tasks SET done = ?, updated_at = ? WHERE id = ?`, doneInt, now, id); err != nil {
			return models.Task{}, err
		}
	}

	var t models.Task
	var doneInt int
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, ticket_id, title, done, position, created_at, updated_at FROM tasks WHERE id = ?`, id,
	).Scan(&t.ID, &t.TicketID, &t.Title, &doneInt, &t.Position, &createdAt, &updatedAt)
	if err != nil {
		return models.Task{}, err
	}
	t.Done = doneInt == 1
	if t.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
		return models.Task{}, err
	}
	if t.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
		return models.Task{}, err
	}

	payload := map[string]any{"task_title": t.Title}
	if done != nil {
		payload["done"] = *done
	}
	if title != nil {
		payload["title"] = *title
	}
	_, _ = s.CreateTicketEvent(ctx, t.TicketID, models.EventTaskUpdated, "", payload)

	return t, nil
}

func (s *SQLiteStore) DeleteTask(ctx context.Context, id string) error {
	// Fetch task details before deletion so we can record the event
	var ticketID, title string
	_ = s.db.QueryRowContext(ctx, `SELECT ticket_id, title FROM tasks WHERE id = ?`, id).Scan(&ticketID, &title)

	_, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return err
	}

	if ticketID != "" {
		_, _ = s.CreateTicketEvent(ctx, ticketID, models.EventTaskDeleted, "", map[string]any{
			"task_title": title,
		})
	}
	return nil
}
