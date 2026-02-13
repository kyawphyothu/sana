package program

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/kyawphyothu/sana/types"
)

const (
	sanaFiglet = `░█▀▀░█▀█░█▀█░█▀█
░▀▀█░█▀█░█░█░█▀█
░▀▀▀░▀░▀░▀░▀░▀░▀`

	minWidth  = 70
	minHeight = 20
)

// View renders the entire UI
func (m model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		res := tea.NewView("Loading...")
		res.AltScreen = true
		return res
	}

	// Check if terminal is too small
	if m.width < minWidth || m.height < minHeight {
		res := tea.NewView(m.renderTooSmallMessage())
		res.AltScreen = true
		return res
	}

	// Build the three sections
	titleBox := m.renderTitleBox()

	if m.selected == addBox {
		res := tea.NewView(lipgloss.JoinVertical(
			lipgloss.Left,
			titleBox,
			m.renderAddBox(),
		))
		res.AltScreen = true
		return res
	}

	expensesBox := m.renderExpensesBox()

	summaryBox := m.renderSummaryBox()

	// Stack vertically
	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleBox,
		expensesBox,
		summaryBox,
	)

	// If overlay is visible, layer it on top of main content using Canvas
	if m.showOverlay {
		overlay := m.renderOverlay()

		// Create canvas with layers
		mainLayer := lipgloss.NewLayer(mainContent).
			Width(m.width).
			Height(m.height).
			X(0).
			Y(0).
			Z(0) // Background layer

		// Calculate overlay position (centered)
		overlayHeight := len(strings.Split(overlay, "\n"))
		overlayWidth := 0
		for _, line := range strings.Split(overlay, "\n") {
			width := lipgloss.Width(line)
			if width > overlayWidth {
				overlayWidth = width
			}
		}

		overlayX := (m.width - overlayWidth) / 2
		overlayY := (m.height - overlayHeight) / 2
		if overlayX < 0 {
			overlayX = 0
		}
		if overlayY < 0 {
			overlayY = 0
		}

		overlayLayer := lipgloss.NewLayer(overlay).
			X(overlayX).
			Y(overlayY).
			Z(1) // Foreground layer (on top)

		canvas := lipgloss.NewCanvas(mainLayer, overlayLayer)
		finalContent := canvas.Render()

		res := tea.NewView(finalContent)
		res.AltScreen = true
		return res
	}

	// Stack vertically and return directly (no wrapper needed)
	res := tea.NewView(mainContent)
	res.AltScreen = true
	return res
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
		dateWidth := 21
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

			// Convert UTC time to local timezone for display
			localDate := expense.Date.Local()
			formattedAmount := formatAmountWithCommas(expense.Amount)
			categoryText := expense.Type.String()

			// Check if this row is selected
			actualRowIndex := m.expensesScrollOffset + i
			isRowSelected := m.isSelected(expensesBox) && actualRowIndex == m.expensesSelectedRow

			// Determine background and foreground colors based on selection
			var bgColor color.Color
			var fgColor color.Color
			if isRowSelected {
				bgColor = m.styles.Theme.Primary
				fgColor = lipgloss.Color("#0F1117")
			} else {
				bgColor = m.styles.Theme.Background
				fgColor = m.styles.Theme.Foreground
			}

			// Create a base style for the row
			baseStyle := lipgloss.NewStyle().
				Foreground(fgColor).
				Background(bgColor)

			// Add bold styling for selected rows
			if isRowSelected {
				baseStyle = baseStyle.Bold(true)
			}

			// Style each column part with proper width and alignment
			datePart := baseStyle.
				Width(dateWidth).
				Align(lipgloss.Left).
				Render(localDate.Format("2006-01-02 15:04:05"))

			descPart := baseStyle.
				Width(descWidth).
				Align(lipgloss.Left).
				Render(desc)

			// Category gets its own color but same background and bold if selected
			categoryColor := CategoryColor(categoryText)
			// Invert category color when row is selected
			if isRowSelected {
				categoryColor = InvertColor(categoryColor)
			}
			categoryStyle := lipgloss.NewStyle().
				Foreground(categoryColor).
				Background(bgColor).
				Width(categoryWidth).
				Align(lipgloss.Left)
			if isRowSelected {
				categoryStyle = categoryStyle.Bold(true)
			}
			categoryPart := categoryStyle.Render(categoryText)

			amountPart := baseStyle.
				Width(amountWidth).
				Align(lipgloss.Right).
				Render(formattedAmount)

			// Spacing between columns (2 spaces) with background
			spacing := baseStyle.Render("  ")

			// Combine all parts - the background should flow continuously
			line := datePart + spacing + descPart + spacing + categoryPart + spacing + amountPart

			content.WriteString(line + "\n")
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
	boxHeight := m.calculateAddBoxHeight()

	// Update prompt styles based on focus before rendering
	m.updatePromptStyles()

	helpStyle := m.styles.Muted

	// Form rows (each textinput has its own prompt)
	rows := []string{
		m.addType.View(),
		m.addAmount.View(),
		m.addDescription.View(),
		m.addDate.View(),
	}
	formContent := strings.Join(rows, "\n\n")

	helpText := helpStyle.Render("↑/↓: move • Tab: autocomplete • Enter: submit • Esc: cancel")

	// List available expense types
	typeLabels := types.ExpenseTypeSuggestions()
	typesList := "Available types: " + strings.Join(typeLabels, ", ")
	typesText := helpStyle.Render(typesList)

	content := formContent + "\n\n" + helpText + "\n" + typesText

	// Show validation/creation error below form if any
	if m.err != nil && m.selected == addBox {
		content += "\n\n" + m.styles.Line.Foreground(m.styles.Theme.Error).Render("Error: "+m.err.Error())
	}

	// Choose border color based on selection
	isSelected := m.isSelected(addBox)
	borderColor := m.styles.Theme.Border
	if isSelected {
		borderColor = m.styles.Theme.Primary
	}

	// Format title with bold for selected part
	title := m.formatMiddleBoxTitle(borderColor)

	return m.styles.DrawBorderWithHeightAndTitleBold(
		content,
		m.width,
		boxHeight,
		RoundedBorderChars(),
		borderColor,
		title,
		false,
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

			formattedAmount := formatAmountWithCommas(cat.Total)
			line := fmt.Sprintf("%-*s  %*d  %*s", categoryWidth, cat.Category, countWidth, cat.Count, amountWidth, formattedAmount)

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
		formattedTotal := formatAmountWithCommas(m.total)
		totalLine := fmt.Sprintf("%-*s  %*s  %*s", categoryWidth, "Total", countWidth, "", amountWidth, formattedTotal)
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
		Align(lipgloss.Center).
		Render(m.styles.Line.Foreground(m.styles.Theme.Error).Render(message))
}

// Helper functions

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

// calculateMiddleBoxHeight calculates height for the stats box (middle box)
func (m model) calculateMiddleBoxHeight() int {
	const titleHeight = 5

	// Remaining height after title box
	remainingHeight := m.height - titleHeight

	// Give stats box half of remaining space (rounded down)
	// The expenses box will get all remaining space to fill terminal completely
	return remainingHeight / 2
}

func (m model) calculateAddBoxHeight() int {
	const titleHeight = 5

	// Remaining height after title box
	remainingHeight := m.height - titleHeight

	// Give add box half of remaining space (rounded down)
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

// renderOverlay renders the overlay showing expenses for the selected category
func (m model) renderOverlay() string {
	// Get selected category
	if m.summarySelectedRow < 0 || m.summarySelectedRow >= len(m.summary) {
		return m.styles.Muted.Render("No category selected")
	}

	selectedCategory := m.summary[m.summarySelectedRow].Category

	// Filter expenses by category
	filteredExpenses := m.filterExpensesByCategory(selectedCategory)

	if len(filteredExpenses) == 0 {
		content := fmt.Sprintf("No expenses found for category: %s", selectedCategory)
		overlayWidth := 60
		overlayHeight := 5
		return m.styles.DrawBorderWithHeightAndTitle(
			content,
			overlayWidth,
			overlayHeight,
			RoundedBorderChars(),
			m.styles.Theme.Primary,
			selectedCategory,
		)
	}

	// Calculate overlay dimensions (centered, reasonable size)
	overlayWidth := m.width - 20
	if overlayWidth < 60 {
		overlayWidth = 60
	}
	if overlayWidth > 100 {
		overlayWidth = 100
	}

	// Calculate available width for table
	tableWidth := overlayWidth - 6 // width minus borders and padding

	// Column widths: Date and Amount
	dateWidth := 12
	amountWidth := 12
	spacing := 2
	totalSpacing := spacing * 2 // 2 gaps between 3 columns
	descriptionWidth := tableWidth - dateWidth - totalSpacing - amountWidth
	if descriptionWidth < 5 {
		descriptionWidth = 5
	}

	// Ensure we have enough space
	if dateWidth+descriptionWidth+amountWidth+totalSpacing > tableWidth {
		dateWidth = tableWidth - descriptionWidth - amountWidth - totalSpacing
		if dateWidth < 10 {
			dateWidth = 10
		}
	}

	var content strings.Builder

	// Table header
	header := fmt.Sprintf("%-*s  %-*s  %*s", dateWidth, "Date", descriptionWidth, "Description", amountWidth, "Amount")
	content.WriteString(m.styles.Header.Render(header) + "\n")

	// Separator line
	separator := strings.Repeat("─", tableWidth)
	content.WriteString(m.styles.Muted.Render(separator) + "\n")

	// Calculate how many rows can fit (leave some space for header and separator)
	maxRows := 15 // Reasonable max for overlay
	if len(filteredExpenses) < maxRows {
		maxRows = len(filteredExpenses)
	}

	// Display expenses (already sorted by date desc)
	for i := 0; i < maxRows; i++ {
		expense := filteredExpenses[i]

		// Convert UTC time to local timezone for display
		localDate := expense.Date.Local()
		formattedAmount := formatAmountWithCommas(expense.Amount)

		datePart := m.styles.Line.
			Width(dateWidth).
			Align(lipgloss.Left).
			Render(localDate.Format("2006-01-02"))

		descriptionPart := m.styles.Line.
			Width(descriptionWidth).
			Align(lipgloss.Left).
			Render(expense.Description)

		amountPart := m.styles.Line.
			Width(amountWidth).
			Align(lipgloss.Right).
			Render(formattedAmount)

		// Spacing between columns
		spacingStr := m.styles.Line.Render("  ")

		line := datePart + spacingStr + descriptionPart + spacingStr + amountPart
		content.WriteString(line + "\n")
	}

	// If there are more expenses, show a message
	if len(filteredExpenses) > maxRows {
		remaining := len(filteredExpenses) - maxRows
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render(fmt.Sprintf("... and %d more", remaining)))
	}

	// Add help text
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("Press Space or Esc to close"))

	overlayHeight := maxRows + 6 // header + separator + rows + before help + help + border * 2
	if overlayHeight < 8 {
		overlayHeight = 7
	}

	return m.styles.DrawBorderWithHeightAndTitle(
		content.String(),
		overlayWidth,
		overlayHeight,
		RoundedBorderChars(),
		m.styles.Theme.Primary,
		selectedCategory,
	)
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
	for _, expense := range m.expenses {
		if expense.Type == targetType {
			filtered = append(filtered, expense)
		}
	}

	return filtered
}

// formatMiddleBoxTitle formats the title for the middle box with bold for selected section
func (m model) formatMiddleBoxTitle(borderColor color.Color) string {
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

	shortcutExpensesText := "[e]"

	// Only bold and brighten if the corresponding box is actually selected
	switch m.selected {
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
