package database

import (
	"database/sql"
	"fmt"
)

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS _migrations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

// migrations run in order. Add new migrations to the end of the slice.
var migrations = []struct {
	name string
	sql  string
}{
	{
		name: "001_initial",
		sql: `
CREATE TABLE IF NOT EXISTS expenses (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date DATETIME NOT NULL,
	amount REAL NOT NULL,
	description TEXT DEFAULT NULL,
	expense_type TEXT NOT NULL DEFAULT 'other',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`,
	},
}

// Migrate runs all pending migrations on db.
func Migrate(db *sql.DB) error {
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	for _, m := range migrations {
		applied, err := migrationApplied(db, m.name)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", m.name, err)
		}
		if applied {
			continue
		}

		if _, err := db.Exec(m.sql); err != nil {
			return fmt.Errorf("migration %s: %w", m.name, err)
		}
		if _, err := db.Exec("INSERT INTO _migrations (name) VALUES (?)", m.name); err != nil {
			return fmt.Errorf("record migration %s: %w", m.name, err)
		}
	}

	return nil
}

func migrationApplied(db *sql.DB, name string) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM _migrations WHERE name = ?", name).Scan(&count)
	return count > 0, err
}
