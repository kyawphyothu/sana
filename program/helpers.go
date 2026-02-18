package program

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/kyawphyothu/sana/types"
)

// formatAmountWithCommas formats a float64 amount with comma separators for thousands
func formatAmountWithCommas(amount float64) string {
	// Format to 2 decimal places
	formatted := fmt.Sprintf("%.2f", amount)

	// Split into integer and decimal parts
	parts := strings.Split(formatted, ".")
	if len(parts) != 2 {
		return formatted
	}

	integerPart := parts[0]
	decimalPart := parts[1]

	// Handle negative numbers
	negative := false
	if len(integerPart) > 0 && integerPart[0] == '-' {
		negative = true
		integerPart = integerPart[1:]
	}

	// Add commas every 3 digits from right to left
	var result strings.Builder
	if negative {
		result.WriteString("-")
	}

	length := len(integerPart)
	for i, digit := range integerPart {
		if i > 0 && (length-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}

	// Add decimal part
	result.WriteString(".")
	result.WriteString(decimalPart)

	return result.String()
}

// filterExpensesByCategory filters expenses by category name (display name)
// Expenses are already sorted by date desc from DB, so we maintain that order
func (m model) filterExpensesByCategory(categoryName string) []types.Expense {
	// Convert category display name back to ExpenseType
	var targetType types.ExpenseType
	for _, expType := range types.AllExpenseTypes() {
		if expType.String() == categoryName {
			targetType = expType
			break
		}
	}

	// Filter expenses (maintains date desc order from DB)
	var filtered []types.Expense
	for _, expense := range m.data.expenses {
		if expense.Type == targetType {
			filtered = append(filtered, expense)
		}
	}

	return filtered
}

// calculateExpensesBoxHeight calculates height for the stats box (middle box)
func (m model) calculateExpensesBoxHeight() int {
	// Remaining height after title box
	remainingHeight := m.ui.height - titleBoxHeight

	// Give stats box half of remaining space (rounded down)
	// The expenses box will get all remaining space to fill terminal completely
	return remainingHeight / boxHeightDivisor
}

func (m model) calculateAddBoxHeight() int {
	// Remaining height after title box
	remainingHeight := m.ui.height - titleBoxHeight

	// Give add box all remaining space
	return remainingHeight
}

// formatSummaryTitle formats the title for the summary box with bold if selected
func (m model) formatSummaryTitle(borderColor color.Color, isSelected bool) string {
	// Shortcut key style - stands out
	shortcutStyle := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Success).
		Background(m.styles.Theme.Background)

	var textStyle lipgloss.Style
	if isSelected {
		// Selected: bright color + bold for maximum visibility
		textStyle = lipgloss.NewStyle().
			Foreground(m.styles.Theme.Selected).
			Background(m.styles.Theme.Background).
			Bold(true)
	} else {
		// Unselected: muted color for readability without distraction
		textStyle = lipgloss.NewStyle().
			Foreground(m.styles.Theme.Muted).
			Background(m.styles.Theme.Background)
	}

	return shortcutStyle.Render("[s]") + textStyle.Render("Summary")
}

// formatExpensesAndAddBoxTitle formats the title for the middle box with bold for selected section
func (m model) formatExpensesAndAddBoxTitle(borderColor color.Color) string {
	// Create styles for selected (bright + bold for visibility) and unselected (muted for readability)
	selectedStyle := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Selected).
		Background(m.styles.Theme.Background).
		Bold(true)

	unselectedStyle := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Muted).
		Background(m.styles.Theme.Background)

	// Shortcut key style - stands out
	shortcutStyle := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Success).
		Background(m.styles.Theme.Background)

	separatorStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Background(m.styles.Theme.Background)

	// Build title with appropriate bold sections based on m.ui.selected
	var expensesText, addText string

	shortcutExpensesText := "[e]"

	// Only bold and brighten if the corresponding box is actually selected
	switch m.ui.selected {
	case expensesBox:
		expensesText = selectedStyle.Render("Expenses")
		addText = unselectedStyle.Render("Add Expense")
	case addBox:
		shortcutExpensesText = "[esc]"
		expensesText = unselectedStyle.Render("Expenses")
		addText = selectedStyle.Render("Add Expense")
	default:
		// summaryBox or other is selected - show both as unselected
		expensesText = unselectedStyle.Render("Expenses")
		addText = unselectedStyle.Render("Add Expense")
	}

	title := shortcutStyle.Render(shortcutExpensesText) +
		expensesText +
		separatorStyle.Render(" - ") +
		shortcutStyle.Render("[a]") +
		addText

	return title
}

// formatMonthlyReportTitle formats the title for the monthly report box with bold if selected
func (m model) formatMonthlyReportTitle(isSelected bool) string {
	// Shortcut key style - stands out
	shortcutStyle := lipgloss.NewStyle().
		Foreground(m.styles.Theme.Success).
		Background(m.styles.Theme.Background)

	var textStyle lipgloss.Style
	if isSelected {
		// Selected: bright color + bold for maximum visibility
		textStyle = lipgloss.NewStyle().
			Foreground(m.styles.Theme.Selected).
			Background(m.styles.Theme.Background).
			Bold(true)
	} else {
		// Unselected: muted color for readability without distraction
		textStyle = lipgloss.NewStyle().
			Foreground(m.styles.Theme.Muted).
			Background(m.styles.Theme.Background)
	}

	return shortcutStyle.Render("[m]") + textStyle.Render("Monthly Report")
}
