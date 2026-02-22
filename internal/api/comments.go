package api

import (
	"encoding/json"
	"net/http"
	"strings"

	chi "github.com/go-chi/chi/v5"
	"github.com/rgracey/kanban-mcp/internal/store"
)

// CommentRequest represents the request body for comment creation/update
type CommentRequest struct {
	Body string `json:"body"`
}

// ListComments returns all comments for a ticket
func ListComments(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticketID := chi.URLParam(r, "id")
		if ticketID == "" {
			http.Error(w, `{"error": "ticket id is required"}`, http.StatusBadRequest)
			return
		}

		comments, err := s.ListComments(r.Context(), ticketID)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	}
}

// CreateComment creates a new comment
func CreateComment(s store.Store, hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ticketID := chi.URLParam(r, "id")
		if ticketID == "" {
			http.Error(w, `{"error": "ticket id is required"}`, http.StatusBadRequest)
			return
		}

		ticket, err := s.GetTicket(r.Context(), ticketID)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		var req CommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if strings.TrimSpace(req.Body) == "" {
			http.Error(w, `{"error": "body is required"}`, http.StatusBadRequest)
			return
		}

		comment, err := s.CreateComment(r.Context(), ticketID, req.Body)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		hub.Publish(SSEEvent{Type: "comment.created", BoardID: ticket.BoardID})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(comment)
	}
}

// GetComment retrieves a single comment by ID
func GetComment(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		comment, err := s.GetComment(r.Context(), id)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comment)
	}
}

// UpdateComment updates a comment
func UpdateComment(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		var req CommentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if strings.TrimSpace(req.Body) == "" {
			http.Error(w, `{"error": "body is required"}`, http.StatusBadRequest)
			return
		}

		comment, err := s.UpdateComment(r.Context(), id, req.Body)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comment)
	}
}

// DeleteComment deletes a comment
func DeleteComment(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		// Check if comment exists first
		if _, err := s.GetComment(r.Context(), id); err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		err := s.DeleteComment(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
