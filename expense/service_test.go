package expense

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/kyawphyothu/sana/database"
)

// testDB returns an in-memory SQLite DB with migrations applied.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestParseDate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		in      string
		wantErr bool
		check   func(t *testing.T, got time.Time)
	}{
		{
			name:    "empty string returns now",
			in:      "",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				diff := now.Sub(got)
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Second {
					t.Errorf("expected ~now, got %v (diff %v)", got, diff)
				}
			},
		},
		{
			name:    "today returns now",
			in:      "today",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				diff := now.Sub(got)
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Second {
					t.Errorf("expected ~now, got %v (diff %v)", got, diff)
				}
			},
		},
		{
			name:    "TODAY case insensitive",
			in:      "TODAY",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				diff := now.Sub(got)
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Second {
					t.Errorf("expected ~now, got %v", got)
				}
			},
		},
		{
			name:    "date only YYYY-MM-DD",
			in:      "2025-03-15",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				if got.Year() != 2025 || got.Month() != 3 || got.Day() != 15 {
					t.Errorf("expected 2025-03-15, got %v", got.Format("2006-01-02"))
				}
			},
		},
		{
			name:    "datetime YYYY-MM-DD HH:MM:SS",
			in:      "2025-06-10 14:30:00",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				if got.Year() != 2025 || got.Month() != 6 || got.Day() != 10 {
					t.Errorf("expected date 2025-06-10, got %v", got.Format("2006-01-02"))
				}
				if got.Hour() != 14 || got.Minute() != 30 || got.Second() != 0 {
					t.Errorf("expected time 14:30:00, got %v", got.Format("15:04:05"))
				}
			},
		},
		{
			name:    "whitespace trimmed",
			in:      "  2025-03-15  ",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				if got.Year() != 2025 || got.Month() != 3 || got.Day() != 15 {
					t.Errorf("expected 2025-03-15, got %v", got.Format("2006-01-02"))
				}
			},
		},
		{
			name:    "invalid date",
			in:      "not-a-date",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid format",
			in:      "03-15-2025",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid month",
			in:      "2025-13-01",
			wantErr: true,
			check:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDate(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestParseMonth(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		in      string
		wantErr bool
		check   func(t *testing.T, got time.Time)
	}{
		{
			name:    "empty string returns first of current month",
			in:      "",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				if got.Year() != now.Year() || got.Month() != now.Month() || got.Day() != 1 {
					t.Errorf("expected first of current month, got %v", got.Format("2006-01-02"))
				}
				if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 {
					t.Errorf("expected midnight, got %v", got.Format("15:04:05"))
				}
			},
		},
		{
			name:    "YYYY-MM",
			in:      "2025-03",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				if got.Year() != 2025 || got.Month() != 3 || got.Day() != 1 {
					t.Errorf("expected 2025-03-01, got %v", got.Format("2006-01-02"))
				}
			},
		},
		{
			name:    "whitespace trimmed",
			in:      "  2025-12  ",
			wantErr: false,
			check: func(t *testing.T, got time.Time) {
				if got.Year() != 2025 || got.Month() != 12 || got.Day() != 1 {
					t.Errorf("expected 2025-12-01, got %v", got.Format("2006-01-02"))
				}
			},
		},
		{
			name:    "invalid month string",
			in:      "2025",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid format",
			in:      "03-2025",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid month number",
			in:      "2025-13",
			wantErr: true,
			check:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMonth(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMonth(%q) err = %v, wantErr %v", tt.in, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

func TestAddExpense_Validation(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	tests := []struct {
		name     string
		amount   string
		desc     string
		typeStr  string
		dateStr  string
		wantErr  bool
		errContains string
	}{
		{
			name:     "zero amount",
			amount:   "0",
			desc:     "x",
			typeStr:  "food",
			dateStr:  "2025-03-15",
			wantErr:  true,
			errContains: "amount must be a positive number",
		},
		{
			name:     "negative amount",
			amount:   "-10",
			desc:     "x",
			typeStr:  "food",
			dateStr:  "2025-03-15",
			wantErr:  true,
			errContains: "amount must be a positive number",
		},
		{
			name:     "invalid amount",
			amount:   "abc",
			desc:     "x",
			typeStr:  "food",
			dateStr:  "2025-03-15",
			wantErr:  true,
			errContains: "amount must be a positive number",
		},
		{
			name:     "invalid date",
			amount:   "10",
			desc:     "x",
			typeStr:  "food",
			dateStr:  "invalid",
			wantErr:  true,
			errContains: "date must be",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := AddExpense(db, tt.amount, tt.desc, tt.typeStr, tt.dateStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddExpense() err = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("AddExpense() err = %v, want containing %q", err, tt.errContains)
				}
			}
			if !tt.wantErr && id <= 0 {
				t.Errorf("AddExpense() id = %d, want positive", id)
			}
		})
	}
}

func TestAddExpense_Success(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	id, err := AddExpense(db, "42.50", "coffee", "food", "2025-03-15")
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}
	if id <= 0 {
		t.Errorf("AddExpense id = %d, want positive", id)
	}

	// Verify row exists
	expenses, err := database.ListExpenses(db, time.Date(2025, 3, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("ListExpenses: %v", err)
	}
	if len(expenses) != 1 {
		t.Fatalf("ListExpenses: got %d rows, want 1", len(expenses))
	}
	e := expenses[0]
	if e.ID != id || e.Amount != 42.50 || e.Description != "coffee" || e.Type.String() != "Food" {
		t.Errorf("got expense %+v", e)
	}
}

func TestAddExpense_DefaultTypeAndDate(t *testing.T) {
	db := testDB(t)
	defer db.Close()

	// Empty type -> other; empty date -> today
	id, err := AddExpense(db, "1", "misc", "", "")
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}
	if id <= 0 {
		t.Errorf("AddExpense id = %d, want positive", id)
	}

	expenses, err := database.ListExpenses(db, time.Now())
	if err != nil {
		t.Fatalf("ListExpenses: %v", err)
	}
	var found bool
	for _, e := range expenses {
		if e.ID == id {
			found = true
			if e.Type.String() != "Other" {
				t.Errorf("expected type Other, got %s", e.Type.String())
			}
			break
		}
	}
	if !found {
		t.Error("new expense not found in current month list")
	}
}

