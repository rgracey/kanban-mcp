package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

// SQLiteStore implements the Store interface using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLiteStore with the given database connection.
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// newUUID generates a new UUID v4 string using crypto/rand.
func newUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		panic(err)
	}
	// Set version 4 (0100xxxx)
	uuid[6] = (uuid[6] & 0x0F) | 0x40
	// Set variant (10xxxxxx)
	uuid[8] = (uuid[8] & 0x3F) | 0x80
	return hex.EncodeToString(uuid)
}

// timeToRFC3339 converts time.Time to RFC3339 string.
func timeToRFC3339(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// rfc3339ToTime parses RFC3339 string to time.Time.
func rfc3339ToTime(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// assert at compile-time that SQLiteStore implements Store
var _ Store = (*SQLiteStore)(nil)
