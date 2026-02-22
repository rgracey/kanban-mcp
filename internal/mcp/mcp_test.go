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
		"list_boards", "create_board", "update_board", "delete_board", "get_board_summary",
		"list_epics", "create_epic", "update_epic", "delete_epic",
		"list_tickets", "get_ticket", "create_ticket", "update_ticket", "delete_ticket", "move_ticket",
		"list_tasks", "create_task", "update_task", "delete_task",
		"list_ticket_events", "add_ticket_event",
		"list_comments", "add_comment", "update_comment", "delete_comment",
	}
	tools := srv.ListTools()
	for _, name := range expected {
		assert.Contains(t, tools, name, "missing tool %q", name)
	}
	assert.Len(t, tools, len(expected), "unexpected number of tools")
}

// TestBoardCRUD exercises board tools.
func TestBoardCRUD(t *testing.T) {
	srv := setupServer(t)

	// create_board → list_boards returns it
	res := call(t, srv, "create_board", map[string]any{"name": "My Board", "description": "desc"})
	var created models.Board
	decodeResult(t, res, &created)
	assert.Equal(t, "My Board", created.Name)
	assert.NotEmpty(t, created.ID)

	listRes := call(t, srv, "list_boards", nil)
	var boards []models.Board
	decodeResult(t, listRes, &boards)
	require.Len(t, boards, 1)
	assert.Equal(t, created.ID, boards[0].ID)

	// update_board
	updateRes := call(t, srv, "update_board", map[string]any{"id": created.ID, "name": "Renamed"})
	var updated models.Board
	decodeResult(t, updateRes, &updated)
	assert.Equal(t, "Renamed", updated.Name)

	// get_board_summary
	summRes := call(t, srv, "get_board_summary", map[string]any{"id": created.ID})
	var summary models.BoardSummary
	decodeResult(t, summRes, &summary)
	assert.Equal(t, created.ID, summary.BoardID)

	// delete_board
	delRes := call(t, srv, "delete_board", map[string]any{"id": created.ID})
	require.False(t, delRes.IsError)

	listRes2 := call(t, srv, "list_boards", nil)
	var boards2 []models.Board
	decodeResult(t, listRes2, &boards2)
	assert.Empty(t, boards2)
}

// TestEpicCRUD exercises epic tools.
func TestEpicCRUD(t *testing.T) {
	srv := setupServer(t)

	// Need a board first
	boardRes := call(t, srv, "create_board", map[string]any{"name": "B"})
	var board models.Board
	decodeResult(t, boardRes, &board)

	epicRes := call(t, srv, "create_epic", map[string]any{"board_id": board.ID, "title": "E1"})
	var epic models.Epic
	decodeResult(t, epicRes, &epic)
	assert.Equal(t, "E1", epic.Title)

	listRes := call(t, srv, "list_epics", map[string]any{"board_id": board.ID})
	var epics []models.Epic
	decodeResult(t, listRes, &epics)
	require.Len(t, epics, 1)

	updRes := call(t, srv, "update_epic", map[string]any{"id": epic.ID, "title": "E1-updated"})
	var updEpic models.Epic
	decodeResult(t, updRes, &updEpic)
	assert.Equal(t, "E1-updated", updEpic.Title)

	delRes := call(t, srv, "delete_epic", map[string]any{"id": epic.ID})
	assert.False(t, delRes.IsError)
}

// TestTicketWorkflow exercises create_ticket → move_ticket → list_tickets with filter.
func TestTicketWorkflow(t *testing.T) {
	srv := setupServer(t)

	boardRes := call(t, srv, "create_board", map[string]any{"name": "B"})
	var board models.Board
	decodeResult(t, boardRes, &board)

	// create_ticket (defaults to todo)
	ticketRes := call(t, srv, "create_ticket", map[string]any{
		"board_id": board.ID,
		"title":    "Fix bug",
	})
	var ticket models.Ticket
	decodeResult(t, ticketRes, &ticket)
	assert.Equal(t, "Fix bug", ticket.Title)
	assert.Equal(t, models.StatusTodo, ticket.Status)

	// move_ticket → in_progress
	moveRes := call(t, srv, "move_ticket", map[string]any{
		"id":     ticket.ID,
		"status": "in_progress",
	})
	var moved models.Ticket
	decodeResult(t, moveRes, &moved)
	assert.Equal(t, models.StatusInProgress, moved.Status)

	// list_tickets?status=in_progress should return the ticket
	listRes := call(t, srv, "list_tickets", map[string]any{
		"board_id": board.ID,
		"status":   "in_progress",
	})
	var tickets []models.Ticket
	decodeResult(t, listRes, &tickets)
	require.Len(t, tickets, 1)
	assert.Equal(t, ticket.ID, tickets[0].ID)

	// list_tickets?status=todo should be empty
	listTodoRes := call(t, srv, "list_tickets", map[string]any{
		"board_id": board.ID,
		"status":   "todo",
	})
	var todoTickets []models.Ticket
	decodeResult(t, listTodoRes, &todoTickets)
	assert.Empty(t, todoTickets)

	// update_ticket fields
	updRes := call(t, srv, "update_ticket", map[string]any{
		"id":       ticket.ID,
		"priority": "high",
	})
	var updTicket models.Ticket
	decodeResult(t, updRes, &updTicket)
	assert.Equal(t, models.PriorityHigh, updTicket.Priority)

	// delete_ticket
	delRes := call(t, srv, "delete_ticket", map[string]any{"id": ticket.ID})
	assert.False(t, delRes.IsError)
}

// TestCommentCRUD exercises comment tools.
func TestCommentCRUD(t *testing.T) {
	srv := setupServer(t)

	boardRes := call(t, srv, "create_board", map[string]any{"name": "B"})
	var board models.Board
	decodeResult(t, boardRes, &board)

	ticketRes := call(t, srv, "create_ticket", map[string]any{
		"board_id": board.ID,
		"title":    "T",
	})
	var ticket models.Ticket
	decodeResult(t, ticketRes, &ticket)

	// add_comment
	addRes := call(t, srv, "add_comment", map[string]any{
		"ticket_id": ticket.ID,
		"body":      "first comment",
	})
	var comment models.Comment
	decodeResult(t, addRes, &comment)
	assert.Equal(t, "first comment", comment.Body)

	// list_comments
	listRes := call(t, srv, "list_comments", map[string]any{"ticket_id": ticket.ID})
	var comments []models.Comment
	decodeResult(t, listRes, &comments)
	require.Len(t, comments, 1)

	// update_comment
	updRes := call(t, srv, "update_comment", map[string]any{
		"id":   comment.ID,
		"body": "edited",
	})
	var updComment models.Comment
	decodeResult(t, updRes, &updComment)
	assert.Equal(t, "edited", updComment.Body)

	// delete_comment
	delRes := call(t, srv, "delete_comment", map[string]any{"id": comment.ID})
	assert.False(t, delRes.IsError)

	listRes2 := call(t, srv, "list_comments", map[string]any{"ticket_id": ticket.ID})
	var comments2 []models.Comment
	decodeResult(t, listRes2, &comments2)
	assert.Empty(t, comments2)
}
