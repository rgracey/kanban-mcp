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

// decodeListResult unmarshals a list tool result (wrapped as {"items":[...]}) into a slice pointer.
func decodeListResult[T any](t *testing.T, result *mcpgo.CallToolResult) []T {
	t.Helper()
	var envelope struct {
		Items []T `json:"items"`
	}
	decodeResult(t, result, &envelope)
	if envelope.Items == nil {
		return []T{}
	}
	return envelope.Items
}

// TestAllToolsRegistered verifies all tools are present.
func TestAllToolsRegistered(t *testing.T) {
	srv := setupServer(t)
	expected := []string{
		"board",
		"epic",
		"ticket",
		"task",
		"note",
		"relation",
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
	boards := decodeListResult[models.Board](t, listRes)
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
	boards2 := decodeListResult[models.Board](t, listRes2)
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
	epics := decodeListResult[models.Epic](t, listRes)
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

	// update status → in_progress
	moveRes := call(t, srv, "ticket", map[string]any{
		"action": "update",
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
	tickets := decodeListResult[models.Ticket](t, listRes)
	require.Len(t, tickets, 1)
	assert.Equal(t, ticket.ID, tickets[0].ID)

	// list with filter_status=todo should be empty
	listTodoRes := call(t, srv, "ticket", map[string]any{
		"action":        "list",
		"board_id":      board.ID,
		"filter_status": "todo",
	})
	todoTickets := decodeListResult[models.Ticket](t, listTodoRes)
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

	// get with include_notes and include_history
	getRes := call(t, srv, "ticket", map[string]any{
		"action":          "get",
		"id":              ticket.ID,
		"include_notes":   true,
		"include_history": true,
	})
	require.False(t, getRes.IsError)

	// delete
	delRes := call(t, srv, "ticket", map[string]any{"action": "delete", "id": ticket.ID})
	assert.False(t, delRes.IsError)
}

// TestNoteCRUD exercises note tool actions.
func TestNoteCRUD(t *testing.T) {
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
	addRes := call(t, srv, "note", map[string]any{
		"action":    "add",
		"ticket_id": ticket.ID,
		"body":      "first note",
	})
	var note models.Note
	decodeResult(t, addRes, &note)
	assert.Equal(t, "first note", note.Body)

	// list
	listRes := call(t, srv, "note", map[string]any{"action": "list", "ticket_id": ticket.ID})
	notes := decodeListResult[models.Note](t, listRes)
	require.Len(t, notes, 1)

	// update
	updRes := call(t, srv, "note", map[string]any{
		"action": "update",
		"id":     note.ID,
		"body":   "edited",
	})
	var updNote models.Note
	decodeResult(t, updRes, &updNote)
	assert.Equal(t, "edited", updNote.Body)

	// delete
	delRes := call(t, srv, "note", map[string]any{"action": "delete", "id": note.ID})
	assert.False(t, delRes.IsError)

	listRes2 := call(t, srv, "note", map[string]any{"action": "list", "ticket_id": ticket.ID})
	notes2 := decodeListResult[models.Note](t, listRes2)
	assert.Empty(t, notes2)
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
	tasks := decodeListResult[models.Task](t, listRes)
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
	tasks2 := decodeListResult[models.Task](t, listRes2)
	assert.Empty(t, tasks2)
}

// TestBulkCreate verifies bulk_create returns the created tickets cleanly.
func TestBulkCreate(t *testing.T) {
	srv := setupServer(t)

	boardRes := call(t, srv, "board", map[string]any{"action": "create", "name": "B"})
	var board models.Board
	decodeResult(t, boardRes, &board)

	bulkRes := call(t, srv, "ticket", map[string]any{
		"action":       "bulk_create",
		"board_id":     board.ID,
		"tickets_json": `[{"title":"Alpha","priority":"high"},{"title":"Beta","priority":"low"}]`,
	})
	require.False(t, bulkRes.IsError, "bulk_create should not error")
	created := decodeListResult[models.Ticket](t, bulkRes)
	require.Len(t, created, 2)
	assert.Equal(t, "Alpha", created[0].Title)
	assert.Equal(t, models.PriorityHigh, created[0].Priority)
	assert.Equal(t, "Beta", created[1].Title)
	assert.Equal(t, models.PriorityLow, created[1].Priority)

	// Verify tickets actually exist (no phantom duplicates)
	listRes := call(t, srv, "ticket", map[string]any{"action": "list", "board_id": board.ID})
	all := decodeListResult[models.Ticket](t, listRes)
	assert.Len(t, all, 2, "should be exactly 2 tickets, not duplicates")
}

// TestTicketHistory exercises the ticket history action.
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

	histRes := call(t, srv, "ticket", map[string]any{"action": "history", "id": ticket.ID})
	events := decodeListResult[models.TicketEvent](t, histRes)
	// At minimum a "created" event should exist
	require.NotEmpty(t, events)
}

// TestTicketGetIncludeTasksAndRelations verifies include_tasks and include_relations on ticket get.
func TestTicketGetIncludeTasksAndRelations(t *testing.T) {
	srv := setupServer(t)

	boardRes := call(t, srv, "board", map[string]any{"action": "create", "name": "B"})
	var board models.Board
	decodeResult(t, boardRes, &board)

	t1Res := call(t, srv, "ticket", map[string]any{"action": "create", "board_id": board.ID, "title": "T1"})
	var t1 models.Ticket
	decodeResult(t, t1Res, &t1)

	t2Res := call(t, srv, "ticket", map[string]any{"action": "create", "board_id": board.ID, "title": "T2"})
	var t2 models.Ticket
	decodeResult(t, t2Res, &t2)

	// add a task to t1
	call(t, srv, "task", map[string]any{"action": "create", "ticket_id": t1.ID, "title": "subtask"})

	// add a relation: t1 blocks t2
	call(t, srv, "relation", map[string]any{"action": "add", "ticket_id": t1.ID, "to_ticket_id": t2.ID})

	// get without flags — should have no tasks or relations embedded
	baseRes := call(t, srv, "ticket", map[string]any{"action": "get", "id": t1.ID})
	var baseEnv struct {
		Tasks     []models.Task           `json:"tasks"`
		Relations []models.TicketRelation `json:"relations"`
	}
	decodeResult(t, baseRes, &baseEnv)
	assert.Nil(t, baseEnv.Tasks)
	assert.Nil(t, baseEnv.Relations)

	// get with include_tasks=true and include_relations=true
	fullRes := call(t, srv, "ticket", map[string]any{
		"action":            "get",
		"id":                t1.ID,
		"include_tasks":     true,
		"include_relations": true,
	})
	var fullEnv struct {
		Tasks     []models.Task           `json:"tasks"`
		Relations []models.TicketRelation `json:"relations"`
	}
	decodeResult(t, fullRes, &fullEnv)
	require.Len(t, fullEnv.Tasks, 1)
	assert.Equal(t, "subtask", fullEnv.Tasks[0].Title)
	require.Len(t, fullEnv.Relations, 1)
	assert.Equal(t, t2.ID, fullEnv.Relations[0].ToTicketID)
}
