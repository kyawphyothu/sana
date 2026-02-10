package program

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	sanaFiglet = `   ____
  / __/__ ____  ___ _
 _\ \/ _ ` + "`" + `/ _ \/ _ ` + "`" + `/
/___/\_,_/_//_/\_,_/`

	minWidth  = 70
	minHeight = 20
)

// View renders the entire UI
func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Check if terminal is too small
	if m.width < minWidth || m.height < minHeight {
		return m.renderTooSmallMessage()
	}

	// Build the three sections
	titleBox := m.renderTitleBox()
	expensesBox := m.renderExpensesBox()
	summaryBox := m.renderSummaryBox()

	// Stack vertically and return directly (no wrapper needed)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleBox,
		expensesBox,
		summaryBox,
	)
}

// renderTitleBox creates the top title section with Sana figlet
func (m model) renderTitleBox() string {
	const titleHeight = 7

	content := m.styles.Title.Render(sanaFiglet)

	return m.styles.DrawBorderWithHeightAndTitle(
		content,
		m.width,
		titleHeight,
		RoundedBorderChars(),
		m.styles.Theme.Primary,
		"Sana",
	)
}

// renderExpensesBox creates the expenses list table section (second box)
func (m model) renderExpensesBox() string {
	boxHeight := m.calculateMiddleBoxHeight()

	var content strings.Builder

	if len(m.expenses) == 0 {
		content.WriteString(m.styles.Muted.Render("No expenses yet"))
	} else {
		// Calculate available width for table
		tableWidth := m.width - 6 // width minus borders and padding

		// Column widths (flexible based on terminal width)
		// Date: 12, Category: 12, Amount: 12, Description: remaining space
		// Spacing between columns: 2 chars each (6 total for 3 gaps)
		dateWidth := 12
		categoryWidth := 12
		amountWidth := 12
		spacing := 2
		totalSpacing := spacing * 3 // 3 gaps between 4 columns

		descWidth := tableWidth - dateWidth - categoryWidth - amountWidth - totalSpacing
		if descWidth < 10 {
			descWidth = 10 // minimum description width
		}

		// Table header
		header := fmt.Sprintf("%-*s  %-*s  %-*s  %*s",
			dateWidth, "Date",
			descWidth, "Description",
			categoryWidth, "Category",
			amountWidth, "Amount")
		content.WriteString(m.styles.Header.Render(header) + "\n")

		// Separator line
		separator := strings.Repeat("─", tableWidth)
		content.WriteString(m.styles.Muted.Render(separator) + "\n")

		// Calculate how many rows can fit
		// boxHeight - 2 (borders) - 2 (header + separator) = available rows
		maxRows := boxHeight - 4
		if maxRows < 1 {
			maxRows = 1
		}

		// Display expenses (already sorted by date desc from DB)
		// Apply scroll offset
		visibleExpenses := m.expenses
		if m.expensesScrollOffset < len(visibleExpenses) {
			visibleExpenses = visibleExpenses[m.expensesScrollOffset:]
		}

		rowCount := 0
		for i, expense := range visibleExpenses {
			if rowCount >= maxRows {
				break // Don't overflow
			}

			// Truncate description if too long
			desc := expense.Description
			if len(desc) > descWidth {
				desc = desc[:descWidth-3] + "..."
			}

			line := fmt.Sprintf("%-*s  %-*s  %-*s  %*.2f",
				dateWidth, expense.Date.Format("2006-01-02"),
				descWidth, desc,
				categoryWidth, expense.Type.String(),
				amountWidth, expense.Amount,
			)

			// Highlight selected row if this box is selected
			actualRowIndex := m.expensesScrollOffset + i
			if m.isSelected(expensesBox) && actualRowIndex == m.expensesSelectedRow {
				content.WriteString(m.styles.Selected.Render(line) + "\n")
			} else {
				content.WriteString(m.styles.Line.Render(line) + "\n")
			}
			rowCount++
		}
	}

	// Show error if any
	if m.err != nil {
		content.WriteString("\n")
		errorStyle := m.styles.Line.Foreground(m.styles.Theme.Error)
		content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	// Choose border color and bold based on selection
	isSelected := m.isSelected(expensesBox)
	borderColor := m.styles.Theme.Primary
	if isSelected {
		borderColor = m.styles.Theme.Selected
	}

	return m.styles.DrawBorderWithHeightAndTitleBold(
		content.String(),
		m.width,
		boxHeight,
		RoundedBorderChars(),
		borderColor,
		"[1]Expenses",
		isSelected,
	)
}

// renderSummaryBox creates the summary section grouped by category (third box)
func (m model) renderSummaryBox() string {
	// Calculate height: use all remaining space after title and expenses box
	const titleHeight = 7
	expensesHeight := m.calculateMiddleBoxHeight()
	boxHeight := m.height - titleHeight - expensesHeight

	var content strings.Builder

	if len(m.expenses) == 0 {
		content.WriteString(m.styles.Muted.Render("No expenses to summarize"))
	} else {
		// Group expenses by category
		categoryTotals := make(map[string]float64)
		for _, expense := range m.expenses {
			categoryTotals[expense.Type.String()] += expense.Amount
		}

		// Calculate available width for table
		tableWidth := m.width - 6 // width minus borders and padding

		// Column widths (flexible based on terminal width)
		// Amount: 15, Category: remaining space
		// Spacing: 2 chars
		amountWidth := 15
		spacing := 2
		categoryWidth := tableWidth - amountWidth - spacing
		if categoryWidth < 10 {
			categoryWidth = 10 // minimum category width
		}

		// Table header
		header := fmt.Sprintf("%-*s  %*s", categoryWidth, "Category", amountWidth, "Amount")
		content.WriteString(m.styles.Header.Render(header) + "\n")

		// Separator line
		separator := strings.Repeat("─", tableWidth)
		content.WriteString(m.styles.Muted.Render(separator) + "\n")

		// Calculate how many rows can fit
		// boxHeight - 2 (borders) - 2 (header + separator) - 2 (separator + total) = available rows
		maxRows := boxHeight - 6
		if maxRows < 1 {
			maxRows = 1
		}

		// Convert map to sorted slice for consistent ordering
		type categoryTotal struct {
			category string
			total    float64
		}
		var categories []categoryTotal
		for category, total := range categoryTotals {
			categories = append(categories, categoryTotal{category, total})
		}

		// Sort categories by total amount in descending order
		sort.Slice(categories, func(i, j int) bool {
			return categories[i].total > categories[j].total
		})

		// Apply scroll offset
		visibleCategories := categories
		if m.summaryScrollOffset < len(visibleCategories) {
			visibleCategories = visibleCategories[m.summaryScrollOffset:]
		}

		// Display category totals
		rowCount := 0
		for i, cat := range visibleCategories {
			if rowCount >= maxRows {
				break // Don't overflow
			}

			line := fmt.Sprintf("%-*s  %*.2f", categoryWidth, cat.category, amountWidth, cat.total)

			// Highlight selected row if this box is selected
			actualRowIndex := m.summaryScrollOffset + i
			if m.isSelected(summaryBox) && actualRowIndex == m.summarySelectedRow {
				content.WriteString(m.styles.Selected.Render(line) + "\n")
			} else {
				content.WriteString(m.styles.Line.Render(line) + "\n")
			}
			rowCount++
		}

		// Separator before total
		content.WriteString(m.styles.Muted.Render(separator) + "\n")

		// Grand total
		grandTotal := m.calculateTotal()
		totalLine := fmt.Sprintf("%-*s  %*.2f", categoryWidth, "Total", amountWidth, grandTotal)
		content.WriteString(m.styles.Header.Render(totalLine))
	}

	// Choose border color and bold based on selection
	isSelected := m.isSelected(summaryBox)
	borderColor := m.styles.Theme.Primary
	if isSelected {
		borderColor = m.styles.Theme.Selected
	}

	return m.styles.DrawBorderWithHeightAndTitleBold(
		content.String(),
		m.width,
		boxHeight,
		RoundedBorderChars(),
		borderColor,
		"[2]Summary",
		isSelected,
	)
}

// renderTooSmallMessage creates small terminal message when terminal is too small
func (m model) renderTooSmallMessage() string {
	message := fmt.Sprintf(
		"Terminal too small!\n\nMinimum size: %dx%d\nCurrent size: %dx%d\n\nPlease resize your terminal.",
		minWidth, minHeight, m.width, m.height,
	)
	return m.styles.Parent.
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(m.styles.Line.Foreground(m.styles.Theme.Error).Render(message))
}

// Helper functions

// calculateTotal calculates the total of all expenses
func (m model) calculateTotal() float64 {
	var total float64
	for _, expense := range m.expenses {
		total += expense.Amount
	}
	return total
}

// calculateMiddleBoxHeight calculates height for the stats box (middle box)
func (m model) calculateMiddleBoxHeight() int {
	const titleHeight = 7

	// Remaining height after title box
	remainingHeight := m.height - titleHeight

	// Give stats box half of remaining space (rounded down)
	// The expenses box will get all remaining space to fill terminal completely
	return remainingHeight / 2
}
