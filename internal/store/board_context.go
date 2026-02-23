package store

import (
	"context"
	"database/sql"

	"github.com/rgracey/kanban-mcp/internal/models"
)

// ReadyTickets returns unblocked todo tickets ordered by priority descending.
// A ticket is "ready" if it is in status=todo and no other ticket blocks it
// (i.e. its ID does not appear as to_ticket_id in ticket_relations).
func (s *SQLiteStore) ReadyTickets(ctx context.Context, boardID string) ([]models.Ticket, error) {
	query := `
		SELECT id, board_id, epic_id, title, description, status, priority, assignee, created_at, updated_at
		FROM tickets
		WHERE board_id = ?
		  AND status = 'todo'
		  AND id NOT IN (SELECT to_ticket_id FROM ticket_relations)
		ORDER BY
		  CASE priority
		    WHEN 'critical' THEN 4
		    WHEN 'high'     THEN 3
		    WHEN 'medium'   THEN 2
		    WHEN 'low'      THEN 1
		    ELSE 0
		  END DESC,
		  created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, boardID)
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

// BoardContext returns a complete snapshot of a board for LLM agents.
// It fetches board metadata, epics, all tickets, tasks, and relations in
// a small number of queries and assembles them in memory.
func (s *SQLiteStore) BoardContext(ctx context.Context, boardID string) (models.BoardContext, error) {
	// 1. Board metadata
	board, err := s.GetBoard(ctx, boardID)
	if err != nil {
		return models.BoardContext{}, err
	}

	// 2. Epics
	epics, err := s.ListEpics(ctx, boardID)
	if err != nil {
		return models.BoardContext{}, err
	}

	// 3. All tickets
	tickets, err := s.ListTickets(ctx, boardID, models.TicketFilter{})
	if err != nil {
		return models.BoardContext{}, err
	}

	// 4. All tasks for this board in one query
	taskRows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.ticket_id, t.title, t.done, t.position, t.created_at, t.updated_at
		FROM tasks t
		JOIN tickets tk ON tk.id = t.ticket_id
		WHERE tk.board_id = ?
		ORDER BY t.ticket_id, t.position ASC, t.created_at ASC`, boardID)
	if err != nil {
		return models.BoardContext{}, err
	}
	defer taskRows.Close()

	tasksByTicket := map[string][]models.Task{}
	for taskRows.Next() {
		var t models.Task
		var doneInt int
		var createdAt, updatedAt string
		if err := taskRows.Scan(&t.ID, &t.TicketID, &t.Title, &doneInt, &t.Position, &createdAt, &updatedAt); err != nil {
			return models.BoardContext{}, err
		}
		t.Done = doneInt == 1
		if t.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
			return models.BoardContext{}, err
		}
		if t.UpdatedAt, err = rfc3339ToTime(updatedAt); err != nil {
			return models.BoardContext{}, err
		}
		tasksByTicket[t.TicketID] = append(tasksByTicket[t.TicketID], t)
	}
	if err := taskRows.Err(); err != nil {
		return models.BoardContext{}, err
	}

	// 5. All relations for this board in one query (both ends on this board)
	relRows, err := s.db.QueryContext(ctx, `
		SELECT r.from_ticket_id, r.to_ticket_id, r.kind, r.created_at,
		       ft.title AS from_title, tt.title AS to_title
		FROM ticket_relations r
		JOIN tickets ft ON ft.id = r.from_ticket_id
		JOIN tickets tt ON tt.id = r.to_ticket_id
		WHERE ft.board_id = ?
		ORDER BY r.created_at ASC`, boardID)
	if err != nil {
		return models.BoardContext{}, err
	}
	defer relRows.Close()

	// Index relations by both ends so each TicketContext gets its own view.
	relsByTicket := map[string][]models.TicketRelation{}
	for relRows.Next() {
		var rel models.TicketRelation
		var createdAt string
		if err := relRows.Scan(&rel.FromTicketID, &rel.ToTicketID, &rel.Kind, &createdAt, &rel.FromTitle, &rel.ToTitle); err != nil {
			return models.BoardContext{}, err
		}
		if rel.CreatedAt, err = rfc3339ToTime(createdAt); err != nil {
			return models.BoardContext{}, err
		}
		relsByTicket[rel.FromTicketID] = append(relsByTicket[rel.FromTicketID], rel)
		// Also index from the to-side so the blocked ticket sees it too.
		if rel.FromTicketID != rel.ToTicketID {
			relsByTicket[rel.ToTicketID] = append(relsByTicket[rel.ToTicketID], rel)
		}
	}
	if err := relRows.Err(); err != nil {
		return models.BoardContext{}, err
	}

	// 6. Assemble
	ticketContexts := make([]models.TicketContext, 0, len(tickets))
	for _, t := range tickets {
		tc := models.TicketContext{
			Ticket:    t,
			Tasks:     tasksByTicket[t.ID],
			Relations: relsByTicket[t.ID],
		}
		if tc.Tasks == nil {
			tc.Tasks = []models.Task{}
		}
		if tc.Relations == nil {
			tc.Relations = []models.TicketRelation{}
		}
		ticketContexts = append(ticketContexts, tc)
	}

	return models.BoardContext{
		Board:   board,
		Epics:   epics,
		Tickets: ticketContexts,
	}, nil
}
