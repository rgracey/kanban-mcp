package db

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpen(t *testing.T) {
	// Create a temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Open database
	dbConn, err := Open(dbPath)
	assert.NoError(t, err)
	defer dbConn.Close()

	// Verify all four tables exist
	expectedTables := []string{"boards", "epics", "tickets", "comments"}

	for _, tableName := range expectedTables {
		var name string
		err := dbConn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&name)
		assert.NoError(t, err, "Table %s should exist", tableName)
	}

	// Verify foreign keys are enabled
	var foreignKeysEnabled int
	err = dbConn.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeysEnabled)
	assert.NoError(t, err)
	assert.Equal(t, 1, foreignKeysEnabled, "Foreign keys should be enabled")

	// Verify WAL mode is enabled
	var journalMode string
	err = dbConn.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	assert.NoError(t, err)
	assert.Equal(t, "wal", journalMode, "WAL mode should be enabled")
}

func TestOpenReRun(t *testing.T) {
	// Create a temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// First run - should create tables and run migrations
	dbConn1, err := Open(dbPath)
	assert.NoError(t, err)
	dbConn1.Close()

	// Second run - should not fail or re-apply migrations
	dbConn2, err := Open(dbPath)
	assert.NoError(t, err)
	defer dbConn2.Close()

	// Verify tables still exist
	expectedTables := []string{"boards", "epics", "tickets", "comments"}
	for _, tableName := range expectedTables {
		var name string
		err := dbConn2.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&name)
		assert.NoError(t, err, "Table %s should still exist after second open", tableName)
	}
}

func TestOpenInvalidPath(t *testing.T) {
	// Try to open with an invalid path
	_, err := Open("/nonexistent/path/invalid.db")
	assert.Error(t, err)
}
