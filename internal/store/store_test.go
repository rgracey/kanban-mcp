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

	// Test DeleteBoard (cascades to tickets and notes)
	ticket, err := store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Test Ticket",
		Description: "Test Description",
		Status:      models.StatusTodo,
		Priority:    models.PriorityMedium,
	})
	require.NoError(t, err)

	note, err := store.CreateNote(ctx, ticket.ID, "Test Note")
	require.NoError(t, err)

	err = store.DeleteBoard(ctx, board.ID)
	require.NoError(t, err)

	// Verify cascade delete - board should be gone
	_, err = store.GetBoard(ctx, board.ID)
	assert.Error(t, err)

	// Verify cascade delete - ticket should be gone
	_, err = store.GetTicket(ctx, ticket.ID)
	assert.Error(t, err)

	// Verify cascade delete - note should be gone
	_, err = store.GetNote(ctx, note.ID)
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

	// Test DeleteTicket (cascades to notes)
	ticketWithNote, err := store.CreateTicket(ctx, board.ID, models.Ticket{
		Title:       "Ticket with Note",
		Description: "Test",
		Status:      models.StatusTodo,
		Priority:    models.PriorityMedium,
	})
	require.NoError(t, err)

	note, err := store.CreateNote(ctx, ticketWithNote.ID, "Test Note")
	require.NoError(t, err)

	err = store.DeleteTicket(ctx, ticketWithNote.ID)
	require.NoError(t, err)

	// Verify ticket is deleted
	_, err = store.GetTicket(ctx, ticketWithNote.ID)
	assert.Error(t, err)

	// Verify cascade delete - note should be gone
	_, err = store.GetNote(ctx, note.ID)
	assert.Error(t, err)
}

// TestStoreNotes tests note CRUD operations.
func TestStoreNotes(t *testing.T) {
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

	// Test CreateNote
	note, err := store.CreateNote(ctx, ticket.ID, "First Note")
	require.NoError(t, err)
	assert.NotEmpty(t, note.ID)
	assert.Equal(t, ticket.ID, note.TicketID)
	assert.Equal(t, "First Note", note.Body)
	assert.NotZero(t, note.CreatedAt)
	assert.NotZero(t, note.UpdatedAt)

	// Test GetNote
	retrievedNote, err := store.GetNote(ctx, note.ID)
	require.NoError(t, err)
	assert.Equal(t, note.ID, retrievedNote.ID)
	assert.Equal(t, note.Body, retrievedNote.Body)

	// Test ListNotes (single note)
	notes, err := store.ListNotes(ctx, ticket.ID)
	require.NoError(t, err)
	assert.Len(t, notes, 1)
	assert.Equal(t, note.ID, notes[0].ID)

	// Test UpdateNote
	updatedNote, err := store.UpdateNote(ctx, note.ID, "Updated Note Body")
	require.NoError(t, err)
	assert.Equal(t, "Updated Note Body", updatedNote.Body)

	// Test ListNotes (multiple notes, ordered by created_at asc)
	// Wait a bit to ensure different timestamps
	time.Sleep(100 * time.Millisecond)
	note1, err := store.CreateNote(ctx, ticket.ID, "Second Note")
	require.NoError(t, err)

	// Wait a bit more for another distinct timestamp
	time.Sleep(100 * time.Millisecond)
	note2, err := store.CreateNote(ctx, ticket.ID, "Third Note")
	require.NoError(t, err)

	notes, err = store.ListNotes(ctx, ticket.ID)
	require.NoError(t, err)
	assert.Len(t, notes, 3)

	// Notes should be ordered by created_at ascending
	// First note should be the original one
	assert.Equal(t, note.ID, notes[0].ID)
	// Second should be note1
	assert.Equal(t, note1.ID, notes[1].ID)
	// Third should be note2
	assert.Equal(t, note2.ID, notes[2].ID)

	// Test DeleteNote
	err = store.DeleteNote(ctx, note1.ID)
	require.NoError(t, err)

	notes, err = store.ListNotes(ctx, ticket.ID)
	require.NoError(t, err)
	assert.Len(t, notes, 2)
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
