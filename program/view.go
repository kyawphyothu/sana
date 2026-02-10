package program

import (
	"fmt"
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
	statsBox := m.renderStatsBox()
	expensesBox := m.renderExpensesBox()

	// Stack vertically and return directly (no wrapper needed)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleBox,
		statsBox,
		expensesBox,
	)
}

// renderTitleBox creates the top title section with Sana figlet
func (m model) renderTitleBox() string {
	const titleHeight = 7

	content := m.styles.Title.Render(sanaFiglet)

	return m.styles.DrawBorderWithHeight(
		content,
		m.width,
		titleHeight,
		RoundedBorderChars(),
		m.styles.Theme.Primary,
	)
}

// renderStatsBox creates the middle stats section
func (m model) renderStatsBox() string {
	boxHeight := m.calculateMiddleBoxHeight()

	var content strings.Builder

	// Calculate total balance
	total := m.calculateTotal()

	content.WriteString(m.styles.Header.Render("Balance") + "\n")
	content.WriteString(m.styles.Line.Render(fmt.Sprintf("$%.2f", total)) + "\n\n")
	content.WriteString(m.styles.Muted.Render(fmt.Sprintf("Total expenses: %d", len(m.expenses))))

	// Show error if any
	if m.err != nil {
		content.WriteString("\n\n")
		errorStyle := m.styles.Line.Foreground(m.styles.Theme.Error)
		content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	// Choose border color based on selection
	borderColor := m.styles.Theme.Primary
	if m.isSelected(statsBox) {
		borderColor = m.styles.Theme.Selected
	}

	return m.styles.DrawBorderWithHeight(
		content.String(),
		m.width,
		boxHeight,
		RoundedBorderChars(),
		borderColor,
	)
}

// renderExpensesBox creates the bottom expenses list section
func (m model) renderExpensesBox() string {
	// Calculate height: use all remaining space after title and stats box
	const titleHeight = 7
	statsHeight := m.calculateMiddleBoxHeight()
	boxHeight := m.height - titleHeight - statsHeight

	var content strings.Builder

	content.WriteString(m.styles.Header.Render("Recent Expenses:") + "\n")

	if len(m.expenses) == 0 {
		content.WriteString(m.styles.Muted.Render("No expenses yet"))
	} else {
		for _, expense := range m.expenses {
			line := fmt.Sprintf("- %s: $%.2f",
				expense.Date.Format("2006-01-02"),
				expense.Amount,
			)
			content.WriteString(m.styles.Line.Render(line) + "\n")
		}
	}

	// Choose border color based on selection
	borderColor := m.styles.Theme.Primary
	if m.isSelected(expensesBox) {
		borderColor = m.styles.Theme.Selected
	}

	return m.styles.DrawBorderWithHeight(
		content.String(),
		m.width,
		boxHeight,
		RoundedBorderChars(),
		borderColor,
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
