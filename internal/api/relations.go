package api

import (
	"encoding/json"
	"net/http"
	"strings"

	chi "github.com/go-chi/chi/v5"
	"github.com/rgracey/kanban-mcp/internal/models"
	"github.com/rgracey/kanban-mcp/internal/store"
)

// ListRelations returns all relations for a ticket.
func ListRelations(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticketID := chi.URLParam(r, "id")
		relations, err := s.ListRelations(r.Context(), ticketID)
		if err != nil {
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(relations)
	}
}

type relationRequest struct {
	ToTicketID string `json:"to_ticket_id"`
	Kind       string `json:"kind"`
}

// AddRelation creates a relation from this ticket to another.
func AddRelation(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fromID := chi.URLParam(r, "id")

		var req relationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(req.ToTicketID) == "" {
			http.Error(w, `{"error":"to_ticket_id is required"}`, http.StatusBadRequest)
			return
		}
		if fromID == req.ToTicketID {
			http.Error(w, `{"error":"a ticket cannot be related to itself"}`, http.StatusBadRequest)
			return
		}

		kind := models.RelationKind(req.Kind)
		if kind == "" {
			kind = models.RelationBlocks
		}
		if kind != models.RelationBlocks {
			http.Error(w, `{"error":"kind must be 'blocks'"}`, http.StatusBadRequest)
			return
		}

		rel, err := s.AddRelation(r.Context(), fromID, req.ToTicketID, kind)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
				return
			}
			// duplicate primary key = relation already exists
			if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "constraint") {
				http.Error(w, `{"error":"relation already exists"}`, http.StatusConflict)
				return
			}
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rel)
	}
}

// DeleteRelation removes a specific relation from this ticket to another.
func DeleteRelation(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fromID := chi.URLParam(r, "id")
		toID := chi.URLParam(r, "toId")
		kindParam := r.URL.Query().Get("kind")
		if kindParam == "" {
			kindParam = string(models.RelationBlocks)
		}

		if err := s.DeleteRelation(r.Context(), fromID, toID, models.RelationKind(kindParam)); err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error":"relation not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
