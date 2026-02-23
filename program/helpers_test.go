package program

import (
	"strings"
	"testing"
	"time"

	"github.com/kyawphyothu/sana/types"
)

func TestFormatAmountWithCommas(t *testing.T) {
	tests := []struct {
		amount float64
		want  string
	}{
		{0, "0.00"},
		{1.5, "1.50"},
		{100, "100.00"},
		{1000, "1,000.00"},
		{1234567.89, "1,234,567.89"},
		{-500.25, "-500.25"},
		{-1234.56, "-1,234.56"},
	}
	for _, tt := range tests {
		got := formatAmountWithCommas(tt.amount)
		if got != tt.want {
			t.Errorf("formatAmountWithCommas(%v) = %q, want %q", tt.amount, got, tt.want)
		}
	}
}

func TestFilterExpensesByCategory(t *testing.T) {
	now := time.Now()
	expenses := []types.Expense{
		{ID: 1, Type: types.ExpenseTypeFood, Amount: 10, Date: now},
		{ID: 2, Type: types.ExpenseTypeTransport, Amount: 20, Date: now},
		{ID: 3, Type: types.ExpenseTypeFood, Amount: 15, Date: now},
	}
	m := model{data: expenseData{expenses: expenses}}
	got := m.filterExpensesByCategory("Food")
	if len(got) != 2 {
		t.Fatalf("filterExpensesByCategory(\"Food\") len = %d, want 2", len(got))
	}
	for _, e := range got {
		if e.Type != types.ExpenseTypeFood {
			t.Errorf("expected type Food, got %v", e.Type)
		}
	}
	gotNone := m.filterExpensesByCategory("Bills")
	if len(gotNone) != 0 {
		t.Errorf("filterExpensesByCategory(\"Bills\") len = %d, want 0", len(gotNone))
	}
}

func TestCalculateExpensesBoxHeight(t *testing.T) {
	m := model{ui: uiState{height: 25}}
	got := m.calculateExpensesBoxHeight()
	// remainingHeight = 25 - 5 = 20, 20/2 = 10
	want := 10
	if got != want {
		t.Errorf("calculateExpensesBoxHeight() = %d, want %d", got, want)
	}
}

func TestCalculateAddBoxHeight(t *testing.T) {
	m := model{ui: uiState{height: 30}}
	got := m.calculateAddBoxHeight()
	want := 30 - titleBoxHeight
	if got != want {
		t.Errorf("calculateAddBoxHeight() = %d, want %d", got, want)
	}
}

func TestFormatSummaryTitle(t *testing.T) {
	m := minimalModelWithStyles()
	// Just ensure it doesn't panic and contains expected text
	got := m.formatSummaryTitle(m.styles.Theme.Border, true)
	if got == "" {
		t.Error("formatSummaryTitle returned empty string")
	}
	if !strings.Contains(got, "[s]") || !strings.Contains(got, "Summary") {
		t.Errorf("formatSummaryTitle should contain [s] and Summary, got %q", got)
	}
}

func TestFormatMonthlyReportTitle(t *testing.T) {
	m := minimalModelWithStyles()
	got := m.formatMonthlyReportTitle(true)
	if got == "" {
		t.Error("formatMonthlyReportTitle returned empty string")
	}
	if !strings.Contains(got, "[m]") || !strings.Contains(got, "Monthly Report") {
		t.Errorf("formatMonthlyReportTitle should contain [m] and Monthly Report, got %q", got)
	}
}

func TestFormatExpensesAndAddBoxTitle(t *testing.T) {
	m := minimalModelWithStyles()
	m.ui.selected = expensesBox
	got := m.formatExpensesAndAddBoxTitle(m.styles.Theme.Border)
	if got == "" {
		t.Error("formatExpensesAndAddBoxTitle returned empty string")
	}
	if !strings.Contains(got, "Expenses") || !strings.Contains(got, "Add Expense") {
		t.Errorf("formatExpensesAndAddBoxTitle should contain Expenses and Add Expense, got %q", got)
	}
}

// minimalModelWithStyles returns a model with only styles set (for helpers that need m.styles).
func minimalModelWithStyles() model {
	return model{styles: NewStyles(DefaultTheme())}
}
