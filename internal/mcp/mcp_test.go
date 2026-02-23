package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rgracey/kanban-mcp/internal/db"
	"github.com/rgracey/kanban-mcp/internal/models"
	"github.com/rgracey/kanban-mcp/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupServer creates a fresh MCP server backed by a temp SQLite DB.
func setupServer(t *testing.T) *server.MCPServer {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	sqlDB, err := db.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { sqlDB.Close() })
	s := store.NewSQLiteStore(sqlDB)
	return NewServer(s)
}

// call invokes a named tool and returns its result, failing the test on error.
func call(t *testing.T, srv *server.MCPServer, toolName string, args map[string]any) *mcpgo.CallToolResult {
	t.Helper()
	tool := srv.GetTool(toolName)
	require.NotNil(t, tool, "tool %q not registered", toolName)
	req := mcpgo.CallToolRequest{}
	req.Params.Arguments = args
	result, err := tool.Handler(context.Background(), req)
	require.NoError(t, err)
	return result
}

// decodeResult unmarshals the first text content of a tool result into v.
func decodeResult(t *testing.T, result *mcpgo.CallToolResult, v any) {
	t.Helper()
	require.False(t, result.IsError, "tool returned error: %v", result.Content)
	require.NotEmpty(t, result.Content)
	text, ok := result.Content[0].(mcpgo.TextContent)
	require.True(t, ok, "expected TextContent, got %T", result.Content[0])
	require.NoError(t, json.Unmarshal([]byte(text.Text), v))
}

// TestAllToolsRegistered verifies all tools are present.
func TestAllToolsRegistered(t *testing.T) {
	srv := setupServer(t)
	expected := []string{
		"board",
		"epic",
		"ticket",
		"task",
		"comment",
		"ticket_history",
	}
	tools := srv.ListTools()
	for _, name := range expected {
		assert.Contains(t, tools, name, "missing tool %q", name)
	}
	assert.Len(t, tools, len(expected), "unexpected number of tools")
}

// TestBoardCRUD exercises board tool actions.
func TestBoardCRUD(t *testing.T) {
	srv := setupServer(t)

	// create
	res := call(t, srv, "board", map[string]any{"action": "create", "name": "My Board", "description": "desc"})
	var created models.Board
	decodeResult(t, res, &created)
	assert.Equal(t, "My Board", created.Name)
	assert.NotEmpty(t, created.ID)

	// list
	listRes := call(t, srv, "board", map[string]any{"action": "list"})
	var boards []models.Board
	decodeResult(t, listRes, &boards)
	require.Len(t, boards, 1)
	assert.Equal(t, created.ID, boards[0].ID)

	// update
	updateRes := call(t, srv, "board", map[string]any{"action": "update", "id": created.ID, "name": "Renamed"})
	var updated models.Board
	decodeResult(t, updateRes, &updated)
	assert.Equal(t, "Renamed", updated.Name)

	// summary
	summRes := call(t, srv, "board", map[string]any{"action": "summary", "id": created.ID})
	var summary models.BoardSummary
	decodeResult(t, summRes, &summary)
	assert.Equal(t, created.ID, summary.BoardID)

	// delete
	delRes := call(t, srv, "board", map[string]any{"action": "delete", "id": created.ID})
	require.False(t, delRes.IsError)

	listRes2 := call(t, srv, "board", map[string]any{"action": "list"})
	var boards2 []models.Board
	decodeResult(t, listRes2, &boards2)
	assert.Empty(t, boards2)
}

// TestEpicCRUD exercises epic tool actions.
func TestEpicCRUD(t *testing.T) {
	srv := setupServer(t)

	boardRes := call(t, srv, "board", map[string]any{"action": "create", "name": "B"})
	var board models.Board
	decodeResult(t, boardRes, &board)

	epicRes := call(t, srv, "epic", map[string]any{"action": "create", "board_id": board.ID, "title": "E1"})
	var epic models.Epic
	decodeResult(t, epicRes, &epic)
	assert.Equal(t, "E1", epic.Title)

	listRes := call(t, srv, "epic", map[string]any{"action": "list", "board_id": board.ID})
	var epics []models.Epic
	decodeResult(t, listRes, &epics)
	require.Len(t, epics, 1)

	updRes := call(t, srv, "epic", map[string]any{"action": "update", "id": epic.ID, "title": "E1-updated"})
	var updEpic models.Epic
	decodeResult(t, updRes, &updEpic)
	assert.Equal(t, "E1-updated", updEpic.Title)

	delRes := call(t, srv, "epic", map[string]any{"action": "delete", "id": epic.ID})
	assert.False(t, delRes.IsError)
}

// TestTicketWorkflow exercises ticket tool actions.
func TestTicketWorkflow(t *testing.T) {
	srv := setupServer(t)

	boardRes := call(t, srv, "board", map[string]any{"action": "create", "name": "B"})
	var board models.Board
	decodeResult(t, boardRes, &board)

	// create (defaults to todo)
	ticketRes := call(t, srv, "ticket", map[string]any{
		"action":   "create",
		"board_id": board.ID,
		"title":    "Fix bug",
	})
	var ticket models.Ticket
	decodeResult(t, ticketRes, &ticket)
	assert.Equal(t, "Fix bug", ticket.Title)
	assert.Equal(t, models.StatusTodo, ticket.Status)

	// move → in_progress
	moveRes := call(t, srv, "ticket", map[string]any{
		"action": "move",
		"id":     ticket.ID,
		"status": "in_progress",
	})
	var moved models.Ticket
	decodeResult(t, moveRes, &moved)
	assert.Equal(t, models.StatusInProgress, moved.Status)

	// list with filter_status=in_progress
	listRes := call(t, srv, "ticket", map[string]any{
		"action":        "list",
		"board_id":      board.ID,
		"filter_status": "in_progress",
	})
	var tickets []models.Ticket
	decodeResult(t, listRes, &tickets)
	require.Len(t, tickets, 1)
	assert.Equal(t, ticket.ID, tickets[0].ID)

	// list with filter_status=todo should be empty
	listTodoRes := call(t, srv, "ticket", map[string]any{
		"action":        "list",
		"board_id":      board.ID,
		"filter_status": "todo",
	})
	var todoTickets []models.Ticket
	decodeResult(t, listTodoRes, &todoTickets)
	assert.Empty(t, todoTickets)

	// update fields
	updRes := call(t, srv, "ticket", map[string]any{
		"action":   "update",
		"id":       ticket.ID,
		"priority": "high",
	})
	var updTicket models.Ticket
	decodeResult(t, updRes, &updTicket)
	assert.Equal(t, models.PriorityHigh, updTicket.Priority)

	// get with include_comments and include_history
	getRes := call(t, srv, "ticket", map[string]any{
		"action":           "get",
		"id":               ticket.ID,
		"include_comments": true,
		"include_history":  true,
	})
	require.False(t, getRes.IsError)

	// delete
	delRes := call(t, srv, "ticket", map[string]any{"action": "delete", "id": ticket.ID})
	assert.False(t, delRes.IsError)
}

// TestCommentCRUD exercises comment tool actions.
func TestCommentCRUD(t *testing.T) {
	srv := setupServer(t)

	boardRes := call(t, srv, "board", map[string]any{"action": "create", "name": "B"})
	var board models.Board
	decodeResult(t, boardRes, &board)

	ticketRes := call(t, srv, "ticket", map[string]any{
		"action":   "create",
		"board_id": board.ID,
		"title":    "T",
	})
	var ticket models.Ticket
	decodeResult(t, ticketRes, &ticket)

	// add
	addRes := call(t, srv, "comment", map[string]any{
		"action":    "add",
		"ticket_id": ticket.ID,
		"body":      "first comment",
	})
	var comment models.Comment
	decodeResult(t, addRes, &comment)
	assert.Equal(t, "first comment", comment.Body)

	// list
	listRes := call(t, srv, "comment", map[string]any{"action": "list", "ticket_id": ticket.ID})
	var comments []models.Comment
	decodeResult(t, listRes, &comments)
	require.Len(t, comments, 1)

	// update
	updRes := call(t, srv, "comment", map[string]any{
		"action": "update",
		"id":     comment.ID,
		"body":   "edited",
	})
	var updComment models.Comment
	decodeResult(t, updRes, &updComment)
	assert.Equal(t, "edited", updComment.Body)

	// delete
	delRes := call(t, srv, "comment", map[string]any{"action": "delete", "id": comment.ID})
	assert.False(t, delRes.IsError)

	listRes2 := call(t, srv, "comment", map[string]any{"action": "list", "ticket_id": ticket.ID})
	var comments2 []models.Comment
	decodeResult(t, listRes2, &comments2)
	assert.Empty(t, comments2)
}

// TestTaskCRUD exercises task tool actions.
func TestTaskCRUD(t *testing.T) {
	srv := setupServer(t)

	boardRes := call(t, srv, "board", map[string]any{"action": "create", "name": "B"})
	var board models.Board
	decodeResult(t, boardRes, &board)

	ticketRes := call(t, srv, "ticket", map[string]any{
		"action":   "create",
		"board_id": board.ID,
		"title":    "T",
	})
	var ticket models.Ticket
	decodeResult(t, ticketRes, &ticket)

	// create task
	createRes := call(t, srv, "task", map[string]any{
		"action":    "create",
		"ticket_id": ticket.ID,
		"title":     "Do the thing",
	})
	var task models.Task
	decodeResult(t, createRes, &task)
	assert.Equal(t, "Do the thing", task.Title)
	assert.False(t, task.Done)

	// list
	listRes := call(t, srv, "task", map[string]any{"action": "list", "ticket_id": ticket.ID})
	var tasks []models.Task
	decodeResult(t, listRes, &tasks)
	require.Len(t, tasks, 1)

	// update (mark done)
	updRes := call(t, srv, "task", map[string]any{
		"action": "update",
		"id":     task.ID,
		"done":   true,
	})
	var updTask models.Task
	decodeResult(t, updRes, &updTask)
	assert.True(t, updTask.Done)

	// delete
	delRes := call(t, srv, "task", map[string]any{"action": "delete", "id": task.ID})
	assert.False(t, delRes.IsError)

	listRes2 := call(t, srv, "task", map[string]any{"action": "list", "ticket_id": ticket.ID})
	var tasks2 []models.Task
	decodeResult(t, listRes2, &tasks2)
	assert.Empty(t, tasks2)
}

// TestTicketHistory exercises the ticket_history tool.
func TestTicketHistory(t *testing.T) {
	srv := setupServer(t)

	boardRes := call(t, srv, "board", map[string]any{"action": "create", "name": "B"})
	var board models.Board
	decodeResult(t, boardRes, &board)

	ticketRes := call(t, srv, "ticket", map[string]any{
		"action":   "create",
		"board_id": board.ID,
		"title":    "T",
	})
	var ticket models.Ticket
	decodeResult(t, ticketRes, &ticket)

	histRes := call(t, srv, "ticket_history", map[string]any{"ticket_id": ticket.ID})
	var events []models.TicketEvent
	decodeResult(t, histRes, &events)
	// At minimum a "created" event should exist
	require.NotEmpty(t, events)
}
