package database

import (
	"database/sql"
	"fmt"
	"time"
)

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS _migrations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);`

// migrations run in order. Add new migrations to the end of the slice.
var migrations = []struct {
	name   string
	sql    string
	goFunc func(db *sql.DB) error
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
	{
		name:   "002_utc_to_local_timezone",
		goFunc: migrateUTCToLocal,
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

		if m.sql != "" {
			if _, err := db.Exec(m.sql); err != nil {
				return fmt.Errorf("migration %s: %w", m.name, err)
			}
		}
		if m.goFunc != nil {
			if err := m.goFunc(db); err != nil {
				return fmt.Errorf("migration %s: %w", m.name, err)
			}
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

// migrateUTCToLocal converts all expense date values from UTC to local timezone.
// Previously dates were stored with UTC offset (e.g. "2026-02-28 23:30:00+00:00").
// Now we store local time without offset so SQLite strftime works on local dates.
func migrateUTCToLocal(db *sql.DB) error {
	rows, err := db.Query("SELECT id, date FROM expenses")
	if err != nil {
		return err
	}
	defer rows.Close()

	type record struct {
		id   int64
		date string
	}
	var records []record
	for rows.Next() {
		var r record
		if err := rows.Scan(&r.id, &r.date); err != nil {
			return err
		}
		records = append(records, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	utcFormats := []string{
		"2006-01-02 15:04:05.999999-07:00",
		"2006-01-02T15:04:05.999999Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05-07:00",
		"2006-01-02 15:04:05",
	}

	for _, r := range records {
		var t time.Time
		var parsed bool
		for _, format := range utcFormats {
			t, err = time.Parse(format, r.date)
			if err == nil {
				parsed = true
				break
			}
		}
		if !parsed {
			continue
		}
		localStr := t.Local().Format(DateTimeStorageFormat)
		if _, err := db.Exec("UPDATE expenses SET date = ? WHERE id = ?", localStr, r.id); err != nil {
			return err
		}
	}
	return nil
}
