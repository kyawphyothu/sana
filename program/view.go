package program

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	sanaFiglet = `░█▀▀░█▀█░█▀█░█▀█
░▀▀█░█▀█░█░█░█▀█
░▀▀▀░▀░▀░▀░▀░▀░▀`

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

	// Conditionally render middle box based on selection
	var middleBox string
	if m.selected == addBox {
		middleBox = m.renderAddBox()
	} else {
		middleBox = m.renderExpensesBox()
	}

	summaryBox := m.renderSummaryBox()

	// Stack vertically and return directly (no wrapper needed)
	return lipgloss.JoinVertical(
		lipgloss.Left,
		titleBox,
		middleBox,
		summaryBox,
	)
}

// renderTitleBox creates the top title section with Sana figlet
func (m model) renderTitleBox() string {
	const titleHeight = 5

	content := m.styles.Title.Render(sanaFiglet)

	return m.styles.DrawBorderWithHeightAndTitle(
		content,
		m.width,
		titleHeight,
		DoubleBorderChars(),
		m.styles.Theme.Border,
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

	// Choose border color based on selection
	isSelected := m.isSelected(expensesBox)
	borderColor := m.styles.Theme.Border
	if isSelected {
		borderColor = m.styles.Theme.Primary
	}

	// Format title with bold for selected part
	title := m.formatMiddleBoxTitle(borderColor)

	return m.styles.DrawBorderWithHeightAndTitleBold(
		content.String(),
		m.width,
		boxHeight,
		RoundedBorderChars(),
		borderColor,
		title,
		false, // We handle bold in the title itself
	)
}

// renderAddBox creates the add expense form section (middle box)
func (m model) renderAddBox() string {
	boxHeight := m.calculateMiddleBoxHeight()

	var content strings.Builder
	content.WriteString(m.styles.Muted.Render("Add expense form will go here..."))

	// Choose border color based on selection
	isSelected := m.isSelected(addBox)
	borderColor := m.styles.Theme.Border
	if isSelected {
		borderColor = m.styles.Theme.Primary
	}

	// Format title with bold for selected part
	title := m.formatMiddleBoxTitle(borderColor)

	return m.styles.DrawBorderWithHeightAndTitleBold(
		content.String(),
		m.width,
		boxHeight,
		RoundedBorderChars(),
		borderColor,
		title,
		false, // We handle bold in the title itself
	)
}

// renderSummaryBox creates the summary section grouped by category (third box)
func (m model) renderSummaryBox() string {
	// Calculate height: use all remaining space after title and expenses box
	const titleHeight = 5
	expensesHeight := m.calculateMiddleBoxHeight()
	boxHeight := m.height - titleHeight - expensesHeight

	var content strings.Builder

	if len(m.expenses) == 0 {
		content.WriteString(m.styles.Muted.Render("No expenses to summarize"))
	} else {
		// Calculate available width for table
		tableWidth := m.width - 6 // width minus borders and padding

		// Column widths (flexible based on terminal width)
		// Amount: 15, Category: remaining space
		// Spacing: 2 chars
		amountWidth := 15
		countWidth := 5
		spacing := 2
		totalSpacing := spacing * 2 // 2 gaps between 3 columns
		categoryWidth := tableWidth - amountWidth - countWidth - totalSpacing
		if categoryWidth < 10 {
			categoryWidth = 10 // minimum category width
		}

		// Table header
		header := fmt.Sprintf("%-*s  %*s  %*s", categoryWidth, "Category", countWidth, "Count", amountWidth, "Amount")
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

		// Apply scroll offset (using database summary)
		visibleSummary := m.summary
		if m.summaryScrollOffset < len(visibleSummary) {
			visibleSummary = visibleSummary[m.summaryScrollOffset:]
		}

		// Display category totals
		rowCount := 0
		for i, cat := range visibleSummary {
			if rowCount >= maxRows {
				break // Don't overflow
			}

			line := fmt.Sprintf("%-*s  %*d  %*.2f", categoryWidth, cat.Category, countWidth, cat.Count, amountWidth, cat.Total)

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
		totalLine := fmt.Sprintf("%-*s  %*s  %*.2f", categoryWidth, "Total", countWidth, "", amountWidth, m.total)
		content.WriteString(m.styles.Header.Render(totalLine))
	}

	// Choose border color based on selection
	isSelected := m.isSelected(summaryBox)
	borderColor := m.styles.Theme.Border
	if isSelected {
		borderColor = m.styles.Theme.Primary
	}

	// Format title with bold if selected
	title := m.formatSummaryTitle(borderColor, isSelected)

	return m.styles.DrawBorderWithHeightAndTitleBold(
		content.String(),
		m.width,
		boxHeight,
		RoundedBorderChars(),
		borderColor,
		title,
		false, // We handle bold in the title itself
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

// calculateMiddleBoxHeight calculates height for the stats box (middle box)
func (m model) calculateMiddleBoxHeight() int {
	const titleHeight = 5

	// Remaining height after title box
	remainingHeight := m.height - titleHeight

	// Give stats box half of remaining space (rounded down)
	// The expenses box will get all remaining space to fill terminal completely
	return remainingHeight / 2
}

// formatSummaryTitle formats the title for the summary box with bold if selected
func (m model) formatSummaryTitle(borderColor lipgloss.Color, isSelected bool) string {
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

// formatMiddleBoxTitle formats the title for the middle box with bold for selected section
func (m model) formatMiddleBoxTitle(borderColor lipgloss.Color) string {
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

	// Build title with appropriate bold sections based on m.selected
	var expensesText, addText string

	// Only bold and brighten if the corresponding box is actually selected
	switch m.selected {
	case expensesBox:
		expensesText = selectedStyle.Render("Expenses")
		addText = unselectedStyle.Render("Add Expense")
	case addBox:
		expensesText = unselectedStyle.Render("Expenses")
		addText = selectedStyle.Render("Add Expense")
	default:
		// summaryBox or other is selected - show both as unselected
		expensesText = unselectedStyle.Render("Expenses")
		addText = unselectedStyle.Render("Add Expense")
	}

	title := shortcutStyle.Render("[e]") +
		expensesText +
		separatorStyle.Render(" - ") +
		shortcutStyle.Render("[a]") +
		addText

	return title
}
