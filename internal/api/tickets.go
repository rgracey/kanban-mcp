package api

import (
	"encoding/json"
	"net/http"
	"strings"

	chi "github.com/go-chi/chi/v5"
	"github.com/rgracey/kanban-mcp/internal/models"
	"github.com/rgracey/kanban-mcp/internal/store"
)

// TicketRequest represents the request body for ticket creation
type TicketRequest struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	EpicID      *string `json:"epic_id"`
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
func CreateTicket(s store.Store) http.HandlerFunc {
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

		ticket := models.Ticket{
			BoardID:     boardID,
			EpicID:      req.EpicID,
			Title:       req.Title,
			Description: req.Description,
			Status:      status,
			Priority:    priority,
		}

		ticket, err := s.CreateTicket(r.Context(), boardID, ticket)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

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
func UpdateTicket(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		var req TicketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
			return
		}

		// Validate status if provided
		if req.Status != nil {
			status := models.Status(*req.Status)
			if status != models.StatusTodo && status != models.StatusInProgress && status != models.StatusDone {
				http.Error(w, `{"error": "status must be one of: todo, in_progress, done"}`, http.StatusBadRequest)
				return
			}
		}

		// Validate priority if provided
		if req.Priority != nil {
			priority := models.Priority(*req.Priority)
			if priority != models.PriorityLow && priority != models.PriorityMedium &&
				priority != models.PriorityHigh && priority != models.PriorityCritical {
				http.Error(w, `{"error": "priority must be one of: low, medium, high, critical"}`, http.StatusBadRequest)
				return
			}
		}

		// Build update fields map
		fields := make(map[string]interface{})
		if req.Title != "" {
			fields["title"] = req.Title
		}
		if req.Description != "" {
			fields["description"] = req.Description
		}
		if req.Status != nil {
			fields["status"] = *req.Status
		}
		if req.Priority != nil {
			fields["priority"] = *req.Priority
		}
		if req.EpicID != nil {
			fields["epic_id"] = *req.EpicID
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ticket)
	}
}

// DeleteTicket deletes a ticket
func DeleteTicket(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		// Check if ticket exists first
		if _, err := s.GetTicket(r.Context(), id); err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		err := s.DeleteTicket(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
