package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rgracey/kanban-mcp/internal/models"
)

// CreateTicketEvent inserts a new event into the ticket_events table.
func (s *SQLiteStore) CreateTicketEvent(ctx context.Context, ticketID string, eventType models.TicketEventType, actor string, payload map[string]any) (models.TicketEvent, error) {
	id := newUUID()
	createdAt := timeToRFC3339(time.Now())

	if payload == nil {
		payload = map[string]any{}
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return models.TicketEvent{}, err
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO ticket_events (id, ticket_id, type, actor, payload, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, ticketID, string(eventType), actor, string(payloadJSON), createdAt,
	)
	if err != nil {
		return models.TicketEvent{}, err
	}

	t, err := rfc3339ToTime(createdAt)
	if err != nil {
		return models.TicketEvent{}, err
	}

	return models.TicketEvent{
		ID:        id,
		TicketID:  ticketID,
		Type:      eventType,
		Actor:     actor,
		Payload:   payload,
		CreatedAt: t,
	}, nil
}

// ListTicketEvents returns all events for a ticket ordered oldest-first.
func (s *SQLiteStore) ListTicketEvents(ctx context.Context, ticketID string) ([]models.TicketEvent, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, ticket_id, type, actor, payload, created_at FROM ticket_events WHERE ticket_id = ? ORDER BY created_at ASC`,
		ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]models.TicketEvent, 0)
	for rows.Next() {
		var e models.TicketEvent
		var payloadStr, createdAt string

		if err := rows.Scan(&e.ID, &e.TicketID, &e.Type, &e.Actor, &payloadStr, &createdAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(payloadStr), &e.Payload); err != nil {
			e.Payload = map[string]any{}
		}

		if e.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
			return nil, err
		}

		events = append(events, e)
	}

	return events, rows.Err()
}
