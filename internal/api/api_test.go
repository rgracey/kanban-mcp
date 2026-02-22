package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rgracey/kanban-mcp/internal/db"
	"github.com/rgracey/kanban-mcp/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupAPITest creates a test server with the API router
func setupAPITest(t *testing.T) (*httptest.Server, *store.SQLiteStore) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := tempDir + "/test.db"

	sqlDB, err := db.Open(dbPath)
	require.NoError(t, err)

	store := store.NewSQLiteStore(sqlDB)
	router := NewRouter(store, NewHub(), nil)

	server := httptest.NewServer(router)
	t.Cleanup(func() {
		server.Close()
		sqlDB.Close()
	})

	return server, store
}

// makeRequest is a helper to make HTTP requests
func makeRequest(t *testing.T, server *httptest.Server, method, path string, body []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, server.URL+path, bytes.NewBuffer(body))
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// readResponse reads the response body as JSON
func readResponse(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var result map[string]any
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	return result
}

// readResponseList reads the response body as a JSON array
func readResponseList(t *testing.T, resp *http.Response) []map[string]any {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var result []map[string]any
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)
	return result
}

// TestAPIBoards tests all board endpoints
func TestAPIBoards(t *testing.T) {
	server, _ := setupAPITest(t)

	// Test POST /boards - create board
	t.Run("CreateBoard", func(t *testing.T) {
		reqBody := map[string]string{
			"name":        "Test Board",
			"description": "Test Description",
		}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards", body)

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "expected 201 Created")
		data := readResponse(t, resp)

		assert.NotEmpty(t, data["id"])
		assert.Equal(t, "Test Board", data["name"])
		assert.Equal(t, "Test Description", data["description"])
		assert.NotEmpty(t, data["created_at"])
		assert.NotEmpty(t, data["updated_at"])
	})

	// Test POST /boards - validation: missing name
	t.Run("CreateBoard_ValidationError", func(t *testing.T) {
		reqBody := map[string]string{
			"description": "No name provided",
		}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards", body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400 Bad Request")
		data := readResponse(t, resp)
		assert.Contains(t, data["error"], "name is required")
	})

	// Test POST /boards - validation: empty name
	t.Run("CreateBoard_EmptyName", func(t *testing.T) {
		reqBody := map[string]string{
			"name":        "",
			"description": "Empty name",
		}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards", body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400 Bad Request")
		data := readResponse(t, resp)
		assert.Contains(t, data["error"], "name is required")
	})

	// Test GET /boards - list all boards
	t.Run("ListBoards", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/boards", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data := readResponseList(t, resp)
		assert.GreaterOrEqual(t, len(data), 1, "should have at least 1 board")
	})

	// Test GET /boards/{id} - get board by id
	t.Run("GetBoard", func(t *testing.T) {
		// First create a board
		reqBody := map[string]string{"name": "Get Board Test"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards", body)
		data := readResponse(t, resp)
		boardID := data["id"].(string)

		// Now get it
		resp = makeRequest(t, server, "GET", "/api/v1/boards/"+boardID, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data = readResponse(t, resp)
		assert.Equal(t, "Get Board Test", data["name"])
	})

	// Test GET /boards/{id} - not found
	t.Run("GetBoard_NotFound", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/boards/nonexistent-id", nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})

	// Test PUT /boards/{id} - update board
	t.Run("UpdateBoard", func(t *testing.T) {
		// First create a board
		reqBody := map[string]string{"name": "Original Name"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards", body)
		data := readResponse(t, resp)
		boardID := data["id"].(string)

		// Update it
		reqBody = map[string]string{"name": "Updated Name", "description": "Updated Description"}
		body, _ = json.Marshal(reqBody)
		resp = makeRequest(t, server, "PUT", "/api/v1/boards/"+boardID, body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data = readResponse(t, resp)
		assert.Equal(t, "Updated Name", data["name"])
		assert.Equal(t, "Updated Description", data["description"])
	})

	// Test PUT /boards/{id} - partial update (only name)
	t.Run("UpdateBoard_Partial", func(t *testing.T) {
		// First create a board
		reqBody := map[string]string{"name": "Original Name"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards", body)
		data := readResponse(t, resp)
		boardID := data["id"].(string)

		// Update only name
		reqBody = map[string]string{"name": "Partially Updated"}
		body, _ = json.Marshal(reqBody)
		resp = makeRequest(t, server, "PUT", "/api/v1/boards/"+boardID, body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data = readResponse(t, resp)
		assert.Equal(t, "Partially Updated", data["name"])
	})

	// Test PUT /boards/{id} - not found
	t.Run("UpdateBoard_NotFound", func(t *testing.T) {
		reqBody := map[string]string{"name": "Should Not Work"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "PUT", "/api/v1/boards/nonexistent-id", body)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})

	// Test DELETE /boards/{id} - delete board
	t.Run("DeleteBoard", func(t *testing.T) {
		// First create a board with a ticket and comment
		reqBody := map[string]string{"name": "Delete Me"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards", body)
		data := readResponse(t, resp)
		boardID := data["id"].(string)

		// Create ticket
		ticketReq := map[string]string{
			"title":       "Ticket to Delete",
			"description": "Test",
			"status":      "todo",
		}
		body, _ = json.Marshal(ticketReq)
		resp = makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)
		ticketData := readResponse(t, resp)
		ticketID := ticketData["id"].(string)

		// Create comment
		commentReq := map[string]string{"body": "Comment to Delete"}
		body, _ = json.Marshal(commentReq)
		resp = makeRequest(t, server, "POST", "/api/v1/tickets/"+ticketID+"/comments", body)
		commentData := readResponse(t, resp)
		commentID := commentData["id"].(string)

		// Delete the board
		resp = makeRequest(t, server, "DELETE", "/api/v1/boards/"+boardID, nil)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "expected 204 No Content")

		// Verify board is gone
		resp = makeRequest(t, server, "GET", "/api/v1/boards/"+boardID, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify ticket is gone (cascade delete)
		resp = makeRequest(t, server, "GET", "/api/v1/tickets/"+ticketID, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify comment is gone (cascade delete)
		resp = makeRequest(t, server, "GET", "/api/v1/comments/"+commentID, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Test DELETE /boards/{id} - not found
	t.Run("DeleteBoard_NotFound", func(t *testing.T) {
		resp := makeRequest(t, server, "DELETE", "/api/v1/boards/nonexistent-id", nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})

	// Test GET /boards/{id}/summary
	t.Run("GetBoardSummary", func(t *testing.T) {
		// First create a board
		reqBody := map[string]string{"name": "Summary Board"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards", body)
		data := readResponse(t, resp)
		boardID := data["id"].(string)

		// Create tickets with various statuses
		statuses := []string{"todo", "in_progress", "done"}
		for _, status := range statuses {
			ticketReq := map[string]string{
				"title":       fmt.Sprintf("Ticket %s", status),
				"description": "Test",
				"status":      status,
			}
			body, _ = json.Marshal(ticketReq)
			makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)
		}

		// Get summary
		resp = makeRequest(t, server, "GET", "/api/v1/boards/"+boardID+"/summary", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		summary := readResponse(t, resp)
		assert.Equal(t, boardID, summary["board_id"])

		ticketCounts := summary["ticket_counts"].(map[string]any)
		assert.Equal(t, float64(1), ticketCounts["todo"])
		assert.Equal(t, float64(1), ticketCounts["in_progress"])
		assert.Equal(t, float64(1), ticketCounts["done"])
	})
}

// TestAPIEpics tests all epic endpoints
func TestAPIEpics(t *testing.T) {
	server, _ := setupAPITest(t)

	// First create a board
	reqBody := map[string]string{"name": "Epic Board"}
	body, _ := json.Marshal(reqBody)
	resp := makeRequest(t, server, "POST", "/api/v1/boards", body)
	boardData := readResponse(t, resp)
	boardID := boardData["id"].(string)

	// Test POST /boards/{id}/epics - create epic
	t.Run("CreateEpic", func(t *testing.T) {
		reqBody := map[string]string{
			"title":       "Test Epic",
			"description": "Test Description",
		}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/epics", body)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		data := readResponse(t, resp)

		assert.NotEmpty(t, data["id"])
		assert.Equal(t, boardID, data["board_id"])
		assert.Equal(t, "Test Epic", data["title"])
		assert.Equal(t, "Test Description", data["description"])
		assert.NotEmpty(t, data["created_at"])
		assert.NotEmpty(t, data["updated_at"])
	})

	// Test POST /boards/{id}/epics - validation: missing title
	t.Run("CreateEpic_ValidationError", func(t *testing.T) {
		reqBody := map[string]string{"description": "No title"}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/epics", body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Contains(t, data["error"], "title is required")
	})

	// Test GET /boards/{id}/epics - list epics
	t.Run("ListEpics", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/boards/"+boardID+"/epics", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data := readResponseList(t, resp)
		assert.GreaterOrEqual(t, len(data), 1)
	})

	// Test GET /epics/{id} - get epic by id
	t.Run("GetEpic", func(t *testing.T) {
		// Create an epic first
		reqBody := map[string]string{"title": "Get Epic"}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/epics", body)
		data := readResponse(t, resp)
		epicID := data["id"].(string)

		resp = makeRequest(t, server, "GET", "/api/v1/epics/"+epicID, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data = readResponse(t, resp)
		assert.Equal(t, "Get Epic", data["title"])
	})

	// Test GET /epics/{id} - not found
	t.Run("GetEpic_NotFound", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/epics/nonexistent-id", nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})

	// Test PUT /epics/{id} - update epic
	t.Run("UpdateEpic", func(t *testing.T) {
		// Create an epic first
		reqBody := map[string]string{"title": "Original Epic"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/epics", body)
		data := readResponse(t, resp)
		epicID := data["id"].(string)

		// Update it
		reqBody = map[string]string{"title": "Updated Epic", "description": "Updated Description"}
		body, _ = json.Marshal(reqBody)
		resp = makeRequest(t, server, "PUT", "/api/v1/epics/"+epicID, body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data = readResponse(t, resp)
		assert.Equal(t, "Updated Epic", data["title"])
		assert.Equal(t, "Updated Description", data["description"])
	})

	// Test PUT /epics/{id} - not found
	t.Run("UpdateEpic_NotFound", func(t *testing.T) {
		reqBody := map[string]string{"title": "Should Not Work"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "PUT", "/api/v1/epics/nonexistent-id", body)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})

	// Test DELETE /epics/{id} - delete epic
	t.Run("DeleteEpic", func(t *testing.T) {
		// Create an epic
		reqBody := map[string]string{"title": "Delete Me"}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/epics", body)
		data := readResponse(t, resp)
		epicID := data["id"].(string)

		// Create a ticket in that epic
		ticketReq := map[string]string{
			"title":       "Ticket in Epic",
			"description": "Test",
			"status":      "todo",
			"epic_id":     epicID,
		}
		body, _ = json.Marshal(ticketReq)
		resp = makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)
		ticketData := readResponse(t, resp)
		ticketID := ticketData["id"].(string)

		// Delete the epic
		resp = makeRequest(t, server, "DELETE", "/api/v1/epics/"+epicID, nil)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify epic is gone
		resp = makeRequest(t, server, "GET", "/api/v1/epics/"+epicID, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify ticket still exists but epic_id is now null
		resp = makeRequest(t, server, "GET", "/api/v1/tickets/"+ticketID, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		ticketData = readResponse(t, resp)
		assert.Nil(t, ticketData["epic_id"])
	})

	// Test DELETE /epics/{id} - not found
	t.Run("DeleteEpic_NotFound", func(t *testing.T) {
		resp := makeRequest(t, server, "DELETE", "/api/v1/epics/nonexistent-id", nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})
}

// TestAPITickets tests all ticket endpoints
func TestAPITickets(t *testing.T) {
	server, _ := setupAPITest(t)

	// First create a board
	reqBody := map[string]string{"name": "Ticket Board"}
	body, _ := json.Marshal(reqBody)
	resp := makeRequest(t, server, "POST", "/api/v1/boards", body)
	boardData := readResponse(t, resp)
	boardID := boardData["id"].(string)

	// Test POST /boards/{id}/tickets - create ticket
	t.Run("CreateTicket", func(t *testing.T) {
		reqBody := map[string]string{
			"title":       "Test Ticket",
			"description": "Test Description",
			"status":      "todo",
			"priority":    "high",
		}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		data := readResponse(t, resp)

		assert.NotEmpty(t, data["id"])
		assert.Equal(t, boardID, data["board_id"])
		assert.Equal(t, "Test Ticket", data["title"])
		assert.Equal(t, "Test Description", data["description"])
		assert.Equal(t, "todo", data["status"])
		assert.Equal(t, "high", data["priority"])
		assert.NotEmpty(t, data["created_at"])
		assert.NotEmpty(t, data["updated_at"])
	})

	// Test POST /boards/{id}/tickets - validation: missing title
	t.Run("CreateTicket_ValidationError", func(t *testing.T) {
		reqBody := map[string]string{
			"description": "No title",
			"status":      "todo",
		}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Contains(t, data["error"], "title is required")
	})

	// Test POST /boards/{id}/tickets - validation: invalid status
	t.Run("CreateTicket_InvalidStatus", func(t *testing.T) {
		reqBody := map[string]string{
			"title":    "Ticket",
			"status":   "invalid_status",
			"priority": "medium",
		}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Contains(t, data["error"], "status must be one of")
	})

	// Test POST /boards/{id}/tickets - validation: invalid priority
	t.Run("CreateTicket_InvalidPriority", func(t *testing.T) {
		reqBody := map[string]string{
			"title":    "Ticket",
			"status":   "todo",
			"priority": "invalid_priority",
		}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Contains(t, data["error"], "priority must be one of")
	})

	// Test GET /boards/{id}/tickets - list tickets
	t.Run("ListTickets", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/boards/"+boardID+"/tickets", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data := readResponseList(t, resp)
		assert.GreaterOrEqual(t, len(data), 1)
	})

	// Test GET /boards/{id}/tickets with status filter
	t.Run("ListTickets_StatusFilter", func(t *testing.T) {
		// Create a ticket with status "done"
		reqBody := map[string]string{
			"title":    "Done Ticket",
			"status":   "done",
			"priority": "medium",
		}
		body, _ = json.Marshal(reqBody)
		makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)

		// Filter by status=done
		resp := makeRequest(t, server, "GET", "/api/v1/boards/"+boardID+"/tickets?status=done", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data := readResponseList(t, resp)
		assert.Equal(t, 1, len(data))
		assert.Equal(t, "done", data[0]["status"])
	})

	// Test GET /boards/{id}/tickets with priority filter
	t.Run("ListTickets_PriorityFilter", func(t *testing.T) {
		// Filter by priority=high
		resp := makeRequest(t, server, "GET", "/api/v1/boards/"+boardID+"/tickets?priority=high", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data := readResponseList(t, resp)
		for _, ticket := range data {
			assert.Equal(t, "high", ticket["priority"])
		}
	})

	// Test GET /boards/{id}/tickets with keyword search
	t.Run("ListTickets_QuerySearch", func(t *testing.T) {
		// Create a ticket with specific title
		reqBody := map[string]string{
			"title":       "Ticket with keyword test",
			"description": "Testing keyword search functionality",
			"status":      "todo",
			"priority":    "medium",
		}
		body, _ = json.Marshal(reqBody)
		makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)

		// Search for "keyword"
		resp := makeRequest(t, server, "GET", "/api/v1/boards/"+boardID+"/tickets?q=keyword", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data := readResponseList(t, resp)
		assert.GreaterOrEqual(t, len(data), 1)
		// At least one should have "keyword" in title or description
		found := false
		for _, ticket := range data {
			title := ticket["title"].(string)
			desc := ticket["description"].(string)
			if stringContains(title, "keyword") || stringContains(desc, "keyword") {
				found = true
				break
			}
		}
		assert.True(t, found, "should find at least one ticket with keyword in title or description")
	})

	// Test GET /tickets/{id} - get ticket by id
	t.Run("GetTicket", func(t *testing.T) {
		// Create a ticket first
		reqBody := map[string]string{
			"title":    "Get Ticket",
			"status":   "todo",
			"priority": "medium",
		}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)
		data := readResponse(t, resp)
		ticketID := data["id"].(string)

		resp = makeRequest(t, server, "GET", "/api/v1/tickets/"+ticketID, nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data = readResponse(t, resp)
		assert.Equal(t, "Get Ticket", data["title"])
	})

	// Test GET /tickets/{id} - not found
	t.Run("GetTicket_NotFound", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/tickets/nonexistent-id", nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})

	// Test PUT /tickets/{id} - update ticket
	t.Run("UpdateTicket", func(t *testing.T) {
		// Create a ticket first
		reqBody := map[string]string{
			"title":    "Original Ticket",
			"status":   "todo",
			"priority": "medium",
		}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)
		data := readResponse(t, resp)
		ticketID := data["id"].(string)

		// Update it
		reqBody = map[string]string{
			"title":       "Updated Ticket",
			"description": "Updated Description",
			"status":      "in_progress",
			"priority":    "high",
		}
		body, _ = json.Marshal(reqBody)
		resp = makeRequest(t, server, "PUT", "/api/v1/tickets/"+ticketID, body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data = readResponse(t, resp)
		assert.Equal(t, "Updated Ticket", data["title"])
		assert.Equal(t, "Updated Description", data["description"])
		assert.Equal(t, "in_progress", data["status"])
		assert.Equal(t, "high", data["priority"])
	})

	// Test PUT /tickets/{id} - not found
	t.Run("UpdateTicket_NotFound", func(t *testing.T) {
		reqBody := map[string]string{"title": "Should Not Work"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "PUT", "/api/v1/tickets/nonexistent-id", body)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})

	// Test DELETE /tickets/{id} - delete ticket
	t.Run("DeleteTicket", func(t *testing.T) {
		// Create a ticket
		reqBody := map[string]string{
			"title":    "Delete Me",
			"status":   "todo",
			"priority": "medium",
		}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)
		data := readResponse(t, resp)
		ticketID := data["id"].(string)

		// Create a comment on this ticket
		commentReq := map[string]string{"body": "Comment to Delete"}
		body, _ = json.Marshal(commentReq)
		resp = makeRequest(t, server, "POST", "/api/v1/tickets/"+ticketID+"/comments", body)
		commentData := readResponse(t, resp)
		commentID := commentData["id"].(string)

		// Delete the ticket
		resp = makeRequest(t, server, "DELETE", "/api/v1/tickets/"+ticketID, nil)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify ticket is gone
		resp = makeRequest(t, server, "GET", "/api/v1/tickets/"+ticketID, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Verify comment is gone (cascade delete)
		resp = makeRequest(t, server, "GET", "/api/v1/comments/"+commentID, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Test DELETE /tickets/{id} - not found
	t.Run("DeleteTicket_NotFound", func(t *testing.T) {
		resp := makeRequest(t, server, "DELETE", "/api/v1/tickets/nonexistent-id", nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})
}

// TestAPIComments tests all comment endpoints
func TestAPIComments(t *testing.T) {
	server, _ := setupAPITest(t)

	// First create a board and ticket
	reqBody := map[string]string{"name": "Comment Board"}
	body, _ := json.Marshal(reqBody)
	resp := makeRequest(t, server, "POST", "/api/v1/boards", body)
	boardData := readResponse(t, resp)
	boardID := boardData["id"].(string)

	ticketReq := map[string]string{
		"title":    "Comment Test Ticket",
		"status":   "todo",
		"priority": "medium",
	}
	body, _ = json.Marshal(ticketReq)
	resp = makeRequest(t, server, "POST", "/api/v1/boards/"+boardID+"/tickets", body)
	ticketData := readResponse(t, resp)
	ticketID := ticketData["id"].(string)

	// Test POST /tickets/{id}/comments - create comment
	t.Run("CreateComment", func(t *testing.T) {
		reqBody := map[string]string{"body": "Test Comment"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/tickets/"+ticketID+"/comments", body)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		data := readResponse(t, resp)

		assert.NotEmpty(t, data["id"])
		assert.Equal(t, ticketID, data["ticket_id"])
		assert.Equal(t, "Test Comment", data["body"])
		assert.NotEmpty(t, data["created_at"])
		assert.NotEmpty(t, data["updated_at"])
	})

	// Test POST /tickets/{id}/comments - validation: missing body
	t.Run("CreateComment_ValidationError", func(t *testing.T) {
		reqBody := map[string]string{}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/tickets/"+ticketID+"/comments", body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Contains(t, data["error"], "body is required")
	})

	// Test POST /tickets/{id}/comments - validation: empty body
	t.Run("CreateComment_EmptyBody", func(t *testing.T) {
		reqBody := map[string]string{"body": ""}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/tickets/"+ticketID+"/comments", body)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Contains(t, data["error"], "body is required")
	})

	// Test GET /tickets/{id}/comments - list comments
	t.Run("ListComments", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/api/v1/tickets/"+ticketID+"/comments", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		data := readResponseList(t, resp)
		assert.GreaterOrEqual(t, len(data), 1)
	})

	// Test PUT /comments/{id} - update comment
	t.Run("UpdateComment", func(t *testing.T) {
		// Create a comment first
		reqBody := map[string]string{"body": "Original Comment"}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/tickets/"+ticketID+"/comments", body)
		data := readResponse(t, resp)
		commentID := data["id"].(string)

		// Update it
		reqBody = map[string]string{"body": "Updated Comment Body"}
		body, _ = json.Marshal(reqBody)
		resp = makeRequest(t, server, "PUT", "/api/v1/comments/"+commentID, body)

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		data = readResponse(t, resp)
		assert.Equal(t, "Updated Comment Body", data["body"])
	})

	// Test PUT /comments/{id} - not found
	t.Run("UpdateComment_NotFound", func(t *testing.T) {
		reqBody := map[string]string{"body": "Should Not Work"}
		body, _ := json.Marshal(reqBody)
		resp := makeRequest(t, server, "PUT", "/api/v1/comments/nonexistent-id", body)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})

	// Test DELETE /comments/{id} - delete comment
	t.Run("DeleteComment", func(t *testing.T) {
		// Create a comment
		reqBody := map[string]string{"body": "Delete Me"}
		body, _ = json.Marshal(reqBody)
		resp := makeRequest(t, server, "POST", "/api/v1/tickets/"+ticketID+"/comments", body)
		data := readResponse(t, resp)
		commentID := data["id"].(string)

		// Delete the comment
		resp = makeRequest(t, server, "DELETE", "/api/v1/comments/"+commentID, nil)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// Verify comment is gone
		resp = makeRequest(t, server, "GET", "/api/v1/comments/"+commentID, nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Test DELETE /comments/{id} - not found
	t.Run("DeleteComment_NotFound", func(t *testing.T) {
		resp := makeRequest(t, server, "DELETE", "/api/v1/comments/nonexistent-id", nil)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		data := readResponse(t, resp)
		assert.Equal(t, "not found", data["error"])
	})
}

// Helper function for string contains check
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
