package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/rgracey/kanban-mcp/internal/db"
	"github.com/rgracey/kanban-mcp/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupStore creates a new SQLiteStore with a temporary database.
func setupStore(t *testing.T) (*SQLiteStore, string) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	sqlDB, err := db.Open(dbPath)
	require.NoError(t, err)

	return NewSQLiteStore(sqlDB), dbPath
}

// TestStoreBoards tests board CRUD operations.
func TestStoreBoards(t *testing.T) {
	store, _ := setupStore(t)

	ctx := t.Context()

	// Test CreateBoard
	board, err := store.CreateBoard(ctx, "Test Board", "Test Description")
	require.NoError(t, err)
	assert.NotEmpty(t, board.ID)
	assert.Equal(t, "Test Board", board.Name)
	assert.Equal(t, "Test Description", board.Description)
	assert.NotZero(t, board.CreatedAt)
	assert.NotZero(t, board.UpdatedAt)

	// Test GetBoard
	retrievedBoard, err := store.GetBoard(ctx, board.ID)
	require.NoError(t, err)
	assert.Equal(t, board.ID, retrievedBoard.ID)
	assert.Equal(t, board.Name, retrievedBoard.Name)
	assert.Equal(t, board.Description, retrievedBoard.Description)

	// Test ListBoards
	boards, err := store.ListBoards(ctx)
	require.NoError(t, err)
	assert.Len(t, boards, 1)
	assert.Equal(t, board.ID, boards[0].ID)

	// Test UpdateBoard (partial update)
	newName := "Updated Board Name"
	updatedBoard, err := store.UpdateBoard(ctx, board.ID, &newName, nil)
	require.NoError(t, err)
	assert.Equal(t, newName, updatedBoard.Name)
	assert.Equal(t, board.Description, updatedBoard.Description) // unchanged

	// Test UpdateBoard with nil pointer (no update)
	updatedBoard2, err := store.UpdateBoard(ctx, board.ID, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, updatedBoard.ID, updatedBoard2.ID)

	// Test DeleteBoard (cascades to tickets and comments)
	ticket, err := store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Test Ticket",
		Description: "Test Description",
		Status:      models.StatusTodo,
		Priority:    models.PriorityMedium,
	})
	require.NoError(t, err)

	comment, err := store.CreateComment(ctx, ticket.ID, "Test Comment")
	require.NoError(t, err)

	err = store.DeleteBoard(ctx, board.ID)
	require.NoError(t, err)

	// Verify cascade delete - board should be gone
	_, err = store.GetBoard(ctx, board.ID)
	assert.Error(t, err)

	// Verify cascade delete - ticket should be gone
	_, err = store.GetTicket(ctx, ticket.ID)
	assert.Error(t, err)

	// Verify cascade delete - comment should be gone
	_, err = store.GetComment(ctx, comment.ID)
	assert.Error(t, err)
}

// TestStoreEpics tests epic CRUD operations.
func TestStoreEpics(t *testing.T) {
	store, _ := setupStore(t)

	ctx := t.Context()

	// Create a board first
	board, err := store.CreateBoard(ctx, "Test Board", "")
	require.NoError(t, err)

	// Test CreateEpic
	epic, err := store.CreateEpic(ctx, board.ID, "Test Epic", "Test Description")
	require.NoError(t, err)
	assert.NotEmpty(t, epic.ID)
	assert.Equal(t, board.ID, epic.BoardID)
	assert.Equal(t, "Test Epic", epic.Title)
	assert.Equal(t, "Test Description", epic.Description)
	assert.NotZero(t, epic.CreatedAt)
	assert.NotZero(t, epic.UpdatedAt)

	// Test GetEpic
	retrievedEpic, err := store.GetEpic(ctx, epic.ID)
	require.NoError(t, err)
	assert.Equal(t, epic.ID, retrievedEpic.ID)
	assert.Equal(t, epic.BoardID, retrievedEpic.BoardID)
	assert.Equal(t, epic.Title, retrievedEpic.Title)

	// Test ListEpics
	epics, err := store.ListEpics(ctx, board.ID)
	require.NoError(t, err)
	assert.Len(t, epics, 1)
	assert.Equal(t, epic.ID, epics[0].ID)

	// Test UpdateEpic (partial update)
	newTitle := "Updated Epic Title"
	updatedEpic, err := store.UpdateEpic(ctx, epic.ID, &newTitle, nil)
	require.NoError(t, err)
	assert.Equal(t, newTitle, updatedEpic.Title)
	assert.Equal(t, epic.Description, updatedEpic.Description) // unchanged

	// Test DeleteEpic (orphan tickets, don't delete them)
	ticket, err := store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Ticket in Epic",
		Description: "Test",
		Status:      models.StatusTodo,
		Priority:    models.PriorityMedium,
		EpicID:      &epic.ID,
	})
	require.NoError(t, err)

	err = store.DeleteEpic(ctx, epic.ID)
	require.NoError(t, err)

	// Verify epic is deleted
	_, err = store.GetEpic(ctx, epic.ID)
	assert.Error(t, err)

	// Verify ticket still exists but epic_id is now nil
	updatedTicket, err := store.GetTicket(ctx, ticket.ID)
	require.NoError(t, err)
	assert.Nil(t, updatedTicket.EpicID)
}

// TestStoreTickets tests ticket CRUD operations.
func TestStoreTickets(t *testing.T) {
	store, _ := setupStore(t)

	ctx := t.Context()

	// Create a board first
	board, err := store.CreateBoard(ctx, "Test Board", "")
	require.NoError(t, err)

	// Test CreateTicket
	ticket, err := store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Test Ticket",
		Description: "Test Description",
		Status:      models.StatusTodo,
		Priority:    models.PriorityHigh,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, ticket.ID)
	assert.Equal(t, board.ID, ticket.BoardID)
	assert.Equal(t, "Test Ticket", ticket.Title)
	assert.Equal(t, models.StatusTodo, ticket.Status)
	assert.Equal(t, models.PriorityHigh, ticket.Priority)
	assert.NotZero(t, ticket.CreatedAt)
	assert.NotZero(t, ticket.UpdatedAt)

	// Test GetTicket
	retrievedTicket, err := store.GetTicket(ctx, ticket.ID)
	require.NoError(t, err)
	assert.Equal(t, ticket.ID, retrievedTicket.ID)
	assert.Equal(t, ticket.Title, retrievedTicket.Title)

	// Test UpdateTicket (partial update via map)
	updatedTicket, err := store.UpdateTicket(ctx, ticket.ID, map[string]any{
		"title":       "Updated Title",
		"description": "Updated Description",
		"status":      models.StatusInProgress,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updatedTicket.Title)
	assert.Equal(t, "Updated Description", updatedTicket.Description)
	assert.Equal(t, models.StatusInProgress, updatedTicket.Status)

	// Test UpdateTicket with epic_id
	ticketEpic, err := store.CreateEpic(ctx, board.ID, "Test Epic", "")
	require.NoError(t, err)

	updatedTicket2, err := store.UpdateTicket(ctx, ticket.ID, map[string]any{
		"epic_id": ticketEpic.ID,
	})
	require.NoError(t, err)
	assert.NotNil(t, updatedTicket2.EpicID)
	assert.Equal(t, ticketEpic.ID, *updatedTicket2.EpicID)

	// Test UpdateTicket with nil epic_id
	updatedTicket3, err := store.UpdateTicket(ctx, ticket.ID, map[string]any{
		"epic_id": nil,
	})
	require.NoError(t, err)
	assert.Nil(t, updatedTicket3.EpicID)

	// Test UpdateTicket with invalid field key
	_, err = store.UpdateTicket(ctx, ticket.ID, map[string]any{
		"invalid_field": "value",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid field key")

	// Test ListTickets (no filter) - only 1 ticket exists (the one we created and updated)
	tickets, err := store.ListTickets(ctx, board.ID, models.TicketFilter{})
	require.NoError(t, err)
	assert.Len(t, tickets, 1)
	assert.Equal(t, ticket.ID, tickets[0].ID)

	// Test ListTickets with status filter
	statusFilter := models.StatusInProgress
	tickets, err = store.ListTickets(ctx, board.ID, models.TicketFilter{
		Status: &statusFilter,
	})
	require.NoError(t, err)
	assert.Len(t, tickets, 1)
	assert.Equal(t, models.StatusInProgress, tickets[0].Status)

	// Test DeleteTicket (cascades to comments)
	ticketWithComment, err := store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Ticket with Comment",
		Description: "Test",
		Status:      models.StatusTodo,
		Priority:    models.PriorityMedium,
	})
	require.NoError(t, err)

	comment, err := store.CreateComment(ctx, ticketWithComment.ID, "Test Comment")
	require.NoError(t, err)

	err = store.DeleteTicket(ctx, ticketWithComment.ID)
	require.NoError(t, err)

	// Verify ticket is deleted
	_, err = store.GetTicket(ctx, ticketWithComment.ID)
	assert.Error(t, err)

	// Verify cascade delete - comment should be gone
	_, err = store.GetComment(ctx, comment.ID)
	assert.Error(t, err)
}

// TestStoreComments tests comment CRUD operations.
func TestStoreComments(t *testing.T) {
	store, _ := setupStore(t)

	ctx := t.Context()

	// Create a board and ticket first
	board, err := store.CreateBoard(ctx, "Test Board", "")
	require.NoError(t, err)

	ticket, err := store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Test Ticket",
		Description: "Test",
		Status:      models.StatusTodo,
		Priority:    models.PriorityMedium,
	})
	require.NoError(t, err)

	// Test CreateComment
	comment, err := store.CreateComment(ctx, ticket.ID, "First Comment")
	require.NoError(t, err)
	assert.NotEmpty(t, comment.ID)
	assert.Equal(t, ticket.ID, comment.TicketID)
	assert.Equal(t, "First Comment", comment.Body)
	assert.NotZero(t, comment.CreatedAt)
	assert.NotZero(t, comment.UpdatedAt)

	// Test GetComment
	retrievedComment, err := store.GetComment(ctx, comment.ID)
	require.NoError(t, err)
	assert.Equal(t, comment.ID, retrievedComment.ID)
	assert.Equal(t, comment.Body, retrievedComment.Body)

	// Test ListComments (single comment)
	comments, err := store.ListComments(ctx, ticket.ID)
	require.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, comment.ID, comments[0].ID)

	// Test UpdateComment
	updatedComment, err := store.UpdateComment(ctx, comment.ID, "Updated Comment Body")
	require.NoError(t, err)
	assert.Equal(t, "Updated Comment Body", updatedComment.Body)

	// Test ListComments (multiple comments, ordered by created_at asc)
	// Wait a bit to ensure different timestamps
	time.Sleep(100 * time.Millisecond)
	comment1, err := store.CreateComment(ctx, ticket.ID, "Second Comment")
	require.NoError(t, err)

	// Wait a bit more for another distinct timestamp
	time.Sleep(100 * time.Millisecond)
	comment2, err := store.CreateComment(ctx, ticket.ID, "Third Comment")
	require.NoError(t, err)

	comments, err = store.ListComments(ctx, ticket.ID)
	require.NoError(t, err)
	assert.Len(t, comments, 3)

	// Comments should be ordered by created_at ascending
	// First comment should be the original one
	assert.Equal(t, comment.ID, comments[0].ID)
	// Second should be comment1
	assert.Equal(t, comment1.ID, comments[1].ID)
	// Third should be comment2
	assert.Equal(t, comment2.ID, comments[2].ID)

	// Test DeleteComment
	err = store.DeleteComment(ctx, comment1.ID)
	require.NoError(t, err)

	comments, err = store.ListComments(ctx, ticket.ID)
	require.NoError(t, err)
	assert.Len(t, comments, 2)
}

// TestGetBoardSummary tests the GetBoardSummary function.
func TestGetBoardSummary(t *testing.T) {
	store, _ := setupStore(t)

	ctx := t.Context()

	// Create a board
	board, err := store.CreateBoard(ctx, "Test Board", "")
	require.NoError(t, err)

	// Create epics
	epic1, err := store.CreateEpic(ctx, board.ID, "Epic 1", "")
	require.NoError(t, err)

	epic2, err := store.CreateEpic(ctx, board.ID, "Epic 2", "")
	require.NoError(t, err)

	// Create tickets with various statuses
	_, err = store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Ticket 1",
		Description: "Test",
		Status:      models.StatusTodo,
		Priority:    models.PriorityMedium,
		EpicID:      nil,
	})
	require.NoError(t, err)

	_, err = store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Ticket 2",
		Description: "Test",
		Status:      models.StatusInProgress,
		Priority:    models.PriorityMedium,
		EpicID:      nil,
	})
	require.NoError(t, err)

	_, err = store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Ticket 3",
		Description: "Test",
		Status:      models.StatusDone,
		Priority:    models.PriorityMedium,
		EpicID:      nil,
	})
	require.NoError(t, err)

	_, err = store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Ticket 4",
		Description: "Test",
		Status:      models.StatusTodo,
		Priority:    models.PriorityMedium,
		EpicID:      &epic1.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Ticket 5",
		Description: "Test",
		Status:      models.StatusInProgress,
		Priority:    models.PriorityMedium,
		EpicID:      &epic1.ID,
	})
	require.NoError(t, err)

	_, err = store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Ticket 6",
		Description: "Test",
		Status:      models.StatusDone,
		Priority:    models.PriorityMedium,
		EpicID:      &epic2.ID,
	})
	require.NoError(t, err)

	// Get board summary
	summary, err := store.GetBoardSummary(ctx, board.ID)
	require.NoError(t, err)

	// Verify ticket counts
	assert.Equal(t, board.ID, summary.BoardID)
	assert.Equal(t, 2, summary.TicketCounts[string(models.StatusTodo)])
	assert.Equal(t, 2, summary.TicketCounts[string(models.StatusInProgress)])
	assert.Equal(t, 2, summary.TicketCounts[string(models.StatusDone)])

	// Verify epics
	assert.Len(t, summary.Epics, 2)

	// Find epic summaries by ID
	var epic1Summary, epic2Summary *models.EpicSummary
	for i := range summary.Epics {
		if summary.Epics[i].ID == epic1.ID {
			epic1Summary = &summary.Epics[i]
		}
		if summary.Epics[i].ID == epic2.ID {
			epic2Summary = &summary.Epics[i]
		}
	}

	assert.NotNil(t, epic1Summary)
	assert.NotNil(t, epic2Summary)
	assert.Equal(t, 2, epic1Summary.TicketCount) // Ticket 4 and 5
	assert.Equal(t, 1, epic2Summary.TicketCount) // Ticket 6
}

// TestStoreInterfaceImplementation verifies that SQLiteStore implements Store.
func TestStoreInterfaceImplementation(t *testing.T) {
	// This test will fail to compile if SQLiteStore doesn't implement Store
	var _ Store = (*SQLiteStore)(nil)
}
