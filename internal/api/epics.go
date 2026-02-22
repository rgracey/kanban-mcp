package api

import (
	"encoding/json"
	"net/http"
	"strings"

	chi "github.com/go-chi/chi/v5"
	"github.com/rgracey/kanban-mcp/internal/store"
)

// EpicRequest represents the request body for epic creation/update
type EpicRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ListEpics returns all epics for a board
func ListEpics(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := chi.URLParam(r, "id")
		if boardID == "" {
			http.Error(w, `{"error": "board id is required"}`, http.StatusBadRequest)
			return
		}

		epics, err := s.ListEpics(r.Context(), boardID)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(epics)
	}
}

// CreateEpic creates a new epic
func CreateEpic(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		boardID := chi.URLParam(r, "id")
		if boardID == "" {
			http.Error(w, `{"error": "board id is required"}`, http.StatusBadRequest)
			return
		}

		var req EpicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
			return
		}

		// Validate required fields
		if strings.TrimSpace(req.Title) == "" {
			http.Error(w, `{"error": "title is required"}`, http.StatusBadRequest)
			return
		}

		epic, err := s.CreateEpic(r.Context(), boardID, req.Title, req.Description)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(epic)
	}
}

// GetEpic returns an epic by ID
func GetEpic(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		epic, err := s.GetEpic(r.Context(), id)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(epic)
	}
}

// UpdateEpic updates an epic
func UpdateEpic(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		var req EpicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)
			return
		}

		// For partial updates, only pass non-empty fields
		var titlePtr, descPtr *string
		if req.Title != "" {
			titlePtr = &req.Title
		}
		if req.Description != "" {
			descPtr = &req.Description
		}

		epic, err := s.UpdateEpic(r.Context(), id, titlePtr, descPtr)
		if err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(epic)
	}
}

// DeleteEpic deletes an epic
func DeleteEpic(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, `{"error": "id is required"}`, http.StatusBadRequest)
			return
		}

		// Check if epic exists first
		if _, err := s.GetEpic(r.Context(), id); err != nil {
			if isNotFoundError(err) {
				http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
				return
			}
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		err := s.DeleteEpic(r.Context(), id)
		if err != nil {
			http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
