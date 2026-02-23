package store

import (
	"context"
	"fmt"
	"time"

	"github.com/rgracey/kanban-mcp/internal/models"
)

// ListRelations returns all relations involving a ticket (as either end).
// Each row is returned from the perspective of the queried ticket:
//   - rows where from_ticket_id = ticketID are "blocks" relations
//   - rows where to_ticket_id   = ticketID are "blocked_by" relations
//
// The denormalised From/ToTitle fields are populated from the tickets table.
func (s *SQLiteStore) ListRelations(ctx context.Context, ticketID string) ([]models.TicketRelation, error) {
	query := `
		SELECT
			r.from_ticket_id, r.to_ticket_id, r.kind, r.created_at,
			ft.title AS from_title, tt.title AS to_title
		FROM ticket_relations r
		JOIN tickets ft ON ft.id = r.from_ticket_id
		JOIN tickets tt ON tt.id = r.to_ticket_id
		WHERE r.from_ticket_id = ? OR r.to_ticket_id = ?
		ORDER BY r.created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, ticketID, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	relations := make([]models.TicketRelation, 0)
	for rows.Next() {
		var rel models.TicketRelation
		var createdAt string
		if err := rows.Scan(
			&rel.FromTicketID, &rel.ToTicketID, &rel.Kind, &createdAt,
			&rel.FromTitle, &rel.ToTitle,
		); err != nil {
			return nil, err
		}
		var err error
		if rel.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
			return nil, err
		}
		relations = append(relations, rel)
	}
	return relations, rows.Err()
}

// AddRelation creates a directional relation between two tickets.
// Returns an error if the relation already exists or if either ticket ID is invalid.
func (s *SQLiteStore) AddRelation(ctx context.Context, fromTicketID, toTicketID string, kind models.RelationKind) (models.TicketRelation, error) {
	createdAt := timeToRFC3339(time.Now())
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ticket_relations (from_ticket_id, to_ticket_id, kind, created_at) VALUES (?, ?, ?, ?)`,
		fromTicketID, toTicketID, string(kind), createdAt,
	)
	if err != nil {
		return models.TicketRelation{}, err
	}

	// Emit audit events on both tickets (best-effort)
	fromTicket, _ := s.GetTicket(ctx, fromTicketID)
	toTicket, _ := s.GetTicket(ctx, toTicketID)
	_, _ = s.CreateTicketEvent(ctx, fromTicketID, models.EventRelationAdded, "", map[string]any{
		"kind":           string(kind),
		"related_ticket": toTicketID,
		"related_title":  toTicket.Title,
		"direction":      "outgoing",
	})
	_, _ = s.CreateTicketEvent(ctx, toTicketID, models.EventRelationAdded, "", map[string]any{
		"kind":           string(kind),
		"related_ticket": fromTicketID,
		"related_title":  fromTicket.Title,
		"direction":      "incoming",
	})

	rel := models.TicketRelation{
		FromTicketID: fromTicketID,
		ToTicketID:   toTicketID,
		Kind:         kind,
		FromTitle:    fromTicket.Title,
		ToTitle:      toTicket.Title,
	}
	if rel.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
		return models.TicketRelation{}, err
	}
	return rel, nil
}

// DeleteRelation removes a directional relation between two tickets.
func (s *SQLiteStore) DeleteRelation(ctx context.Context, fromTicketID, toTicketID string, kind models.RelationKind) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM ticket_relations WHERE from_ticket_id = ? AND to_ticket_id = ? AND kind = ?`,
		fromTicketID, toTicketID, string(kind),
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("relation not found")
	}

	// Emit audit events on both tickets (best-effort)
	fromTicket, _ := s.GetTicket(ctx, fromTicketID)
	toTicket, _ := s.GetTicket(ctx, toTicketID)
	_, _ = s.CreateTicketEvent(ctx, fromTicketID, models.EventRelationRemoved, "", map[string]any{
		"kind":           string(kind),
		"related_ticket": toTicketID,
		"related_title":  toTicket.Title,
		"direction":      "outgoing",
	})
	_, _ = s.CreateTicketEvent(ctx, toTicketID, models.EventRelationRemoved, "", map[string]any{
		"kind":           string(kind),
		"related_ticket": fromTicketID,
		"related_title":  fromTicket.Title,
		"direction":      "incoming",
	})
	return nil
}
