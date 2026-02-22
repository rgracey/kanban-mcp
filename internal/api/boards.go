package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rgracey/kanban-mcp/internal/store"
)

// BoardRequest represents the request body for board creation/update
type BoardRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ListBoards returns all boards
func ListBoards(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boards, err := s.ListBoards(r.Context())
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(boards)
	}
}

// CreateBoard creates a new board
func CreateBoard(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req BoardRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if strings.TrimSpace(req.Name) == "" {
			http.Error(w, `{"error": "name is required"}`, http.StatusBadRequest)
			return
		}

		board, err := s.CreateBoard(r.Context(), req.Name, req.Description)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(board)
	}
}

// GetBoard returns a board by ID
func GetBoard(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		board, err := s.GetBoard(r.Context(), id)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(board)
	}
}

// UpdateBoard updates a board
func UpdateBoard(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		var req BoardRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
			return
		}

		// For partial updates, only pass non-empty fields
		var namePtr, descPtr *string
		if req.Name != "" {
			namePtr = &req.Name
		}
		if req.Description != "" {
			descPtr = &req.Description
		}

		board, err := s.UpdateBoard(r.Context(), id, namePtr, descPtr)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(board)
	}
}

// DeleteBoard deletes a board
func DeleteBoard(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		// Check if board exists first
		if _, err := s.GetBoard(r.Context(), id); err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		err := s.DeleteBoard(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// GetBoardSummary returns the summary for a board
func GetBoardSummary(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		summary, err := s.GetBoardSummary(r.Context(), id)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}
