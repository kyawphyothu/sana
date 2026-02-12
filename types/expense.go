package types

import (
	"strings"
	"time"
)

// ExpenseType is a category for an expense. Use these constants app-wide.
type ExpenseType string

const (
	ExpenseTypeFood      ExpenseType = "food"
	ExpenseTypeTransport ExpenseType = "transport"
	ExpenseTypeBills     ExpenseType = "bills"
	ExpenseTypeShopping  ExpenseType = "shopping"
	ExpenseTypeHealth    ExpenseType = "health"
	ExpenseTypeOther     ExpenseType = "other"
)

// AllExpenseTypes returns every type for dropdowns, lists, and validation.
func AllExpenseTypes() []ExpenseType {
	return []ExpenseType{
		ExpenseTypeFood,
		ExpenseTypeTransport,
		ExpenseTypeBills,
		ExpenseTypeShopping,
		ExpenseTypeHealth,
		ExpenseTypeOther,
	}
}

// ExpenseTypeSuggestions returns display names for autocomplete (e.g. "Food", "Transport").
func ExpenseTypeSuggestions() []string {
	types := AllExpenseTypes()
	s := make([]string, len(types))
	for i, t := range types {
		s[i] = t.String()
	}
	return s
}

// ParseExpenseType parses a string (display name or raw value like "food") into ExpenseType.
// Returns (type, true) if valid, (other, false) otherwise.
func ParseExpenseType(s string) (ExpenseType, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	for _, t := range AllExpenseTypes() {
		if strings.ToLower(t.String()) == s || string(t) == s {
			return t, true
		}
	}
	return ExpenseTypeOther, false
}

// String returns a display-friendly label for the type.
func (e ExpenseType) String() string {
	switch e {
	case ExpenseTypeFood:
		return "Food"
	case ExpenseTypeTransport:
		return "Transport"
	case ExpenseTypeBills:
		return "Bills"
	case ExpenseTypeShopping:
		return "Shopping"
	case ExpenseTypeHealth:
		return "Health"
	case ExpenseTypeOther:
		return "Other"
	default:
		return string(e)
	}
}

// Expense is the main expense model for the app.
type Expense struct {
	ID          int64
	Date        time.Time
	Amount      float64
	Description string
	Type        ExpenseType
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CategorySummary represents aggregated expense data by category
type CategorySummary struct {
	Category string
	Count    int
	Total    float64
}
