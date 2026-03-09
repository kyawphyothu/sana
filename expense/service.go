package expense

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kyawphyothu/sana/database"
	"github.com/kyawphyothu/sana/types"
)

// ParseDate parses a date string into a local time.Time.
// Accepts: "" or "today" (now), "YYYY-MM-DD", or "YYYY-MM-DD HH:MM:SS".
// For date-only input, the current time's clock is applied.
func ParseDate(dateStr string) (time.Time, error) {
	s := strings.TrimSpace(dateStr)
	if s == "" || strings.ToLower(s) == "today" {
		return time.Now(), nil
	}
	// Try "YYYY-MM-DD HH:MM:SS" first
	t, err := time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
	if err == nil {
		return t, nil
	}
	// Then "YYYY-MM-DD"
	t, err = time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("date must be YYYY-MM-DD or YYYY-MM-DD HH:MM:SS or 'today': %w", err)
	}
	now := time.Now()
	return time.Date(t.Year(), t.Month(), t.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.Local), nil
}

// ParseMonth parses a month string "YYYY-MM" into the first day of that month in local time.
// Empty string returns the current month (first day).
func ParseMonth(monthStr string) (time.Time, error) {
	s := strings.TrimSpace(monthStr)
	if s == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local), nil
	}
	t, err := time.ParseInLocation("2006-01", s, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("month must be YYYY-MM: %w", err)
	}
	return t, nil
}

// AddExpense validates and parses add-expense input, then creates the expense.
// All parsing and validation live here so CLI and TUI share one implementation.
// Returns the new expense ID or an error (e.g. invalid amount, date, or DB error).
func AddExpense(db *sql.DB, amountStr, description, typeStr, dateStr string) (int64, error) {
	amount, err := strconv.ParseFloat(strings.TrimSpace(amountStr), 64)
	if err != nil || amount <= 0 {
		return 0, fmt.Errorf("amount must be a positive number")
	}
	date, err := ParseDate(dateStr)
	if err != nil {
		return 0, err
	}
	expType, _ := types.ParseExpenseType(typeStr)
	if expType == "" {
		expType = types.ExpenseTypeOther
	}
	desc := strings.TrimSpace(description)
	return database.CreateExpense(db, date, amount, desc, expType)
}
