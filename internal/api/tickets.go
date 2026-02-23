package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	chi "github.com/go-chi/chi/v5"
	"github.com/rgracey/kanban-mcp/internal/models"
	"github.com/rgracey/kanban-mcp/internal/store"
)

// TicketRequest represents the request body for ticket creation/update
type TicketRequest struct {
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Status      *string                  `json:"status"`
	Priority    *string                  `json:"priority"`
	EpicID      *string                  `json:"epic_id"`
	Assignee    *string                  `json:"assignee"`
	References  []models.TicketReference `json:"references"`
	Resolution  *models.TicketResolution `json:"resolution"`
}

// ListTickets returns tickets for a board
func ListTickets(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := chi.URLParam(r, "id")
		if boardID == "" {
			http.Error(w, `{"error": "board id is required"}`, http.StatusBadRequest)
			return
		}

		// Parse query parameters
		query := r.URL.Query()
		filter := models.TicketFilter{}

		if status := query.Get("status"); status != "" {
			s := models.Status(status)
			filter.Status = &s
		}

		if priority := query.Get("priority"); priority != "" {
			p := models.Priority(priority)
			filter.Priority = &p
		}

		if epicID := query.Get("epic_id"); epicID != "" {
			filter.EpicID = &epicID
		}

		if q := query.Get("q"); q != "" {
			filter.Query = &q
		}
		if sortBy := query.Get("sort_by"); sortBy != "" {
			filter.SortBy = &sortBy
		}
		if sortOrder := query.Get("sort_order"); sortOrder != "" {
			filter.SortOrder = &sortOrder
		}

		tickets, err := s.ListTickets(r.Context(), boardID, filter)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tickets)
	}
}

// CreateTicket creates a new ticket
func CreateTicket(s store.Store, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := chi.URLParam(r, "id")
		if boardID == "" {
			http.Error(w, `{"error": "board id is required"}`, http.StatusBadRequest)
			return
		}

		var req TicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if strings.TrimSpace(req.Title) == "" {
			http.Error(w, `{"error": "title is required"}`, http.StatusBadRequest)
			return
		}

		// Set defaults and validate
		var status models.Status
		if req.Status != nil {
			status = models.Status(*req.Status)
			if status != models.StatusTodo && status != models.StatusInProgress && status != models.StatusDone {
				http.Error(w, `{"error": "status must be one of: todo, in_progress, done"}`, http.StatusBadRequest)
				return
			}
		} else {
			status = models.StatusTodo
		}

		var priority models.Priority
		if req.Priority != nil {
			priority = models.Priority(*req.Priority)
			if priority != models.PriorityLow && priority != models.PriorityMedium &&
				priority != models.PriorityHigh && priority != models.PriorityCritical {
				http.Error(w, `{"error": "priority must be one of: low, medium, high, critical"}`, http.StatusBadRequest)
				return
			}
		} else {
			priority = models.PriorityMedium
		}

		assignee := ""
		if req.Assignee != nil {
			assignee = *req.Assignee
		}

		ticket := models.Ticket{
			BoardID:     boardID,
			EpicID:      req.EpicID,
			Title:       req.Title,
			Description: req.Description,
			Status:      status,
			Priority:    priority,
			Assignee:    assignee,
			References:  req.References,
			Resolution:  req.Resolution,
		}

		ticket, err := s.CreateTicket(r.Context(), boardID, ticket)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		hub.Publish(SSEEvent{Type: "ticket.created", BoardID: boardID})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ticket)
	}
}

// GetTicket returns a ticket by ID
func GetTicket(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		ticket, err := s.GetTicket(r.Context(), id)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ticket)
	}
}

// UpdateTicket updates a ticket
func UpdateTicket(s store.Store, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		// Decode into raw map so we can distinguish "field absent" from "field: null"
		var raw map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
			return
		}

		// Build update fields map
		fields := make(map[string]interface{})

		if v, ok := raw["title"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				http.Error(w, `{"error": "invalid title"}`, http.StatusBadRequest)
				return
			}
			if s != "" {
				fields["title"] = s
			}
		}
		if v, ok := raw["description"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				http.Error(w, `{"error": "invalid description"}`, http.StatusBadRequest)
				return
			}
			fields["description"] = s
		}
		if v, ok := raw["status"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				http.Error(w, `{"error": "invalid status"}`, http.StatusBadRequest)
				return
			}
			status := models.Status(s)
			if status != models.StatusTodo && status != models.StatusInProgress && status != models.StatusDone {
				http.Error(w, `{"error": "status must be one of: todo, in_progress, done"}`, http.StatusBadRequest)
				return
			}
			fields["status"] = s
		}
		if v, ok := raw["priority"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				http.Error(w, `{"error": "invalid priority"}`, http.StatusBadRequest)
				return
			}
			priority := models.Priority(s)
			if priority != models.PriorityLow && priority != models.PriorityMedium &&
				priority != models.PriorityHigh && priority != models.PriorityCritical {
				http.Error(w, `{"error": "priority must be one of: low, medium, high, critical"}`, http.StatusBadRequest)
				return
			}
			fields["priority"] = s
		}
		if v, ok := raw["epic_id"]; ok {
			// null clears the epic; a string value sets it
			if string(v) == "null" {
				fields["epic_id"] = nil
			} else {
				var s string
				if err := json.Unmarshal(v, &s); err != nil {
					http.Error(w, `{"error": "invalid epic_id"}`, http.StatusBadRequest)
					return
				}
				fields["epic_id"] = s
			}
		}
		if v, ok := raw["assignee"]; ok {
			var s string
			if err := json.Unmarshal(v, &s); err != nil {
				http.Error(w, `{"error": "invalid assignee"}`, http.StatusBadRequest)
				return
			}
			fields["assignee"] = s
		}
		if v, ok := raw["references"]; ok {
			// Pass the raw JSON string; the store will unmarshal/validate it
			fields["references"] = string(v)
		}
		if v, ok := raw["resolution"]; ok {
			if string(v) == "null" {
				fields["resolution"] = nil
			} else {
				fields["resolution"] = string(v)
			}
		}

		ticket, err := s.UpdateTicket(r.Context(), id, fields)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		hub.Publish(SSEEvent{Type: "ticket.updated", BoardID: ticket.BoardID})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ticket)
	}
}

// BulkCreateTickets creates multiple tickets in a single transaction.
func BulkCreateTickets(s store.Store, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := chi.URLParam(r, "id")
		if boardID == "" {
			http.Error(w, `{"error": "board id is required"}`, http.StatusBadRequest)
			return
		}

		var reqs []TicketRequest
		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			http.Error(w, `{"error": "invalid JSON — expected an array of ticket objects"}`, http.StatusBadRequest)
			return
		}
		if len(reqs) == 0 {
			http.Error(w, `{"error": "at least one ticket is required"}`, http.StatusBadRequest)
			return
		}

		tickets := make([]models.Ticket, 0, len(reqs))
		for i, req := range reqs {
			if strings.TrimSpace(req.Title) == "" {
				http.Error(w, fmt.Sprintf(`{"error": "ticket %d: title is required"}`, i), http.StatusBadRequest)
				return
			}
			var status models.Status
			if req.Status != nil {
				status = models.Status(*req.Status)
				if status != models.StatusTodo && status != models.StatusInProgress && status != models.StatusDone {
					http.Error(w, fmt.Sprintf(`{"error": "ticket %d: invalid status"}`, i), http.StatusBadRequest)
					return
				}
			} else {
				status = models.StatusTodo
			}
			var priority models.Priority
			if req.Priority != nil {
				priority = models.Priority(*req.Priority)
				if priority != models.PriorityLow && priority != models.PriorityMedium &&
					priority != models.PriorityHigh && priority != models.PriorityCritical {
					http.Error(w, fmt.Sprintf(`{"error": "ticket %d: invalid priority"}`, i), http.StatusBadRequest)
					return
				}
			} else {
				priority = models.PriorityMedium
			}
			assignee := ""
			if req.Assignee != nil {
				assignee = *req.Assignee
			}
			tickets = append(tickets, models.Ticket{
				EpicID:      req.EpicID,
				Title:       req.Title,
				Description: req.Description,
				Status:      status,
				Priority:    priority,
				Assignee:    assignee,
			})
		}

		created, err := s.BulkCreateTickets(r.Context(), boardID, tickets)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		hub.Publish(SSEEvent{Type: "ticket.created", BoardID: boardID})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	}
}

// DeleteTicket deletes a ticket
func DeleteTicket(s store.Store, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		// Check if ticket exists first (also gives us boardID for the event)
		ticket, err := s.GetTicket(r.Context(), id)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		if err := s.DeleteTicket(r.Context(), id); err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		hub.Publish(SSEEvent{Type: "ticket.deleted", BoardID: ticket.BoardID})
		w.WriteHeader(http.StatusNoContent)
	}
}

// ReadyTickets returns unblocked todo tickets ordered by priority.
// GET /api/v1/boards/{id}/ready
func ReadyTickets(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := chi.URLParam(r, "id")
		tickets, err := s.ReadyTickets(r.Context(), boardID)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tickets)
	}
}

// GetBoardContext returns a full board snapshot for LLM agents.
// GET /api/v1/boards/{id}/context
func GetBoardContext(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := chi.URLParam(r, "id")
		bctx, err := s.BoardContext(r.Context(), boardID)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bctx)
	}
}
