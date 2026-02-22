package db

import (
	"embed"
	"fmt"

	"database/sql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Open opens a SQLite database at the given path, enables WAL mode and
// foreign keys, and runs all pending migrations.
func Open(path string) (*sql.DB, error) {
	// Open a connection for the migration driver
	migrateDB, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create iofs source
	source, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		migrateDB.Close()
		return nil, fmt.Errorf("failed to create migration source: %w", err)
	}

	// Create sqlite driver for migrate - it takes ownership of migrateDB
	driver, err := sqlite.WithInstance(migrateDB, &sqlite.Config{})
	if err != nil {
		source.Close()
		migrateDB.Close()
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithInstance("iofs", source, "sqlite", driver)
	if err != nil {
		source.Close()
		migrateDB.Close()
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		m.Close()
		source.Close()
		migrateDB.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Close migrate instance - this will close migrateDB
	m.Close()

	// Open a new connection for the app
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database for app: %w", err)
	}

	// Enable WAL mode
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return db, nil
}
