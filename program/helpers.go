package program

import (
	"fmt"
	"strings"

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
