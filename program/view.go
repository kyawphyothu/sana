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
)

// expenseColumnWidths holds column widths for the expenses table
type expenseColumnWidths struct {
	Date     int
	Desc     int
	Category int
	Amount   int
}

// summaryColumnWidths holds column widths for the summary table
type summaryColumnWidths struct {
	Category int
	Count    int
	Amount   int
}

// overlayColumnWidths holds column widths for the category overlay table
type overlayColumnWidths struct {
	Date        int
	Description int
	Amount      int
}

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
	content := m.styles.Title.Render(sanaFiglet)

	return m.styles.DrawBorderWithHeightAndTitle(
		content,
		m.width,
		titleBoxHeight,
		DoubleBorderChars(),
		m.styles.Theme.Border,
		"Sana",
	)
}

// renderExpensesBox creates the expenses list table section (second box)
func (m model) renderExpensesBox() string {
	boxHeight := m.calculateExpensesBoxHeight()

	var content strings.Builder

	if len(m.expenses) == 0 {
		content.WriteString(m.styles.Muted.Render("No expenses yet"))
	} else {
		tableWidth := m.width - tableBorderPadding
		content.WriteString(m.buildExpensesTableHeader(tableWidth))

		maxRows := boxHeight - expensesBoxHeaderRows
		if maxRows < 1 {
			maxRows = 1
		}

		visibleExpenses := m.expenses
		if m.expensesScrollOffset < len(visibleExpenses) {
			visibleExpenses = visibleExpenses[m.expensesScrollOffset:]
		}

		widths := m.calculateExpenseColumnWidths(tableWidth)
		rowCount := 0
		for i, expense := range visibleExpenses {
			if rowCount >= maxRows {
				break
			}
			actualRowIndex := m.expensesScrollOffset + i
			isRowSelected := m.isSelected(expensesBox) && actualRowIndex == m.expensesSelectedRow
			content.WriteString(m.renderExpenseRow(expense, widths, isRowSelected) + "\n")
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
	title := m.formatExpensesAndAddBoxTitle(borderColor)

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

// buildExpensesTableHeader returns the table header line and separator for the expenses box
func (m model) buildExpensesTableHeader(tableWidth int) string {
	widths := m.calculateExpenseColumnWidths(tableWidth)
	header := fmt.Sprintf("%-*s  %-*s  %-*s  %*s",
		widths.Date, "Date",
		widths.Desc, "Description",
		widths.Category, "Category",
		widths.Amount, "Amount")
	separator := strings.Repeat("─", tableWidth)
	return m.styles.Header.Render(header) + "\n" + m.styles.Muted.Render(separator) + "\n"
}

// calculateExpenseColumnWidths computes column widths for the expenses table from available table width
func (m model) calculateExpenseColumnWidths(tableWidth int) expenseColumnWidths {
	dateWidth := tableDateWidth
	categoryWidth := tableCategoryWidth
	amountWidth := tableAmountWidth
	spacing := tableColumnSpacing
	totalSpacing := spacing * tableColumnGapsExpenses
	descWidth := tableWidth - dateWidth - categoryWidth - amountWidth - totalSpacing
	if descWidth < tableMinDescWidth {
		descWidth = tableMinDescWidth
	}
	return expenseColumnWidths{
		Date:     dateWidth,
		Desc:     descWidth,
		Category: categoryWidth,
		Amount:   amountWidth,
	}
}

// renderExpenseRow renders a single expense row with optional selection highlight
func (m model) renderExpenseRow(expense types.Expense, widths expenseColumnWidths, isSelected bool) string {
	desc := expense.Description
	if len(desc) > widths.Desc {
		desc = desc[:widths.Desc-descTruncateSuffix] + "..."
	}
	localDate := expense.Date.Local()
	formattedAmount := formatAmountWithCommas(expense.Amount)
	categoryText := expense.Type.String()

	var bgColor, fgColor color.Color
	if isSelected {
		bgColor = m.styles.Theme.Primary
		fgColor = lipgloss.Color("#0F1117")
	} else {
		bgColor = m.styles.Theme.Background
		fgColor = m.styles.Theme.Foreground
	}
	baseStyle := lipgloss.NewStyle().Foreground(fgColor).Background(bgColor)
	if isSelected {
		baseStyle = baseStyle.Bold(true)
	}

	datePart := baseStyle.Width(widths.Date).Align(lipgloss.Left).Render(localDate.Format("2006-01-02 15:04:05"))
	descPart := baseStyle.Width(widths.Desc).Align(lipgloss.Left).Render(desc)
	categoryColor := CategoryColor(categoryText)
	if isSelected {
		categoryColor = InvertColor(categoryColor)
	}
	categoryStyle := lipgloss.NewStyle().Foreground(categoryColor).Background(bgColor).Width(widths.Category).Align(lipgloss.Left)
	if isSelected {
		categoryStyle = categoryStyle.Bold(true)
	}
	categoryPart := categoryStyle.Render(categoryText)
	amountPart := baseStyle.Width(widths.Amount).Align(lipgloss.Right).Render(formattedAmount)
	spacing := baseStyle.Render("  ")
	return datePart + spacing + descPart + spacing + categoryPart + spacing + amountPart
}

// buildSummaryTableHeader returns the table header and separator for the summary box
func (m model) buildSummaryTableHeader(tableWidth int) string {
	widths := m.calculateSummaryColumnWidths(tableWidth)
	header := fmt.Sprintf("%-*s  %*s  %*s", widths.Category, "Category", widths.Count, "Count", widths.Amount, "Amount")
	separator := strings.Repeat("─", tableWidth)
	return m.styles.Header.Render(header) + "\n" + m.styles.Muted.Render(separator) + "\n"
}

// calculateSummaryColumnWidths computes column widths for the summary table
func (m model) calculateSummaryColumnWidths(tableWidth int) summaryColumnWidths {
	amountWidth := tableAmountWidthSummary
	countWidth := tableCountWidth
	spacing := tableColumnSpacing
	totalSpacing := spacing * tableColumnGapsSummary
	categoryWidth := tableWidth - amountWidth - countWidth - totalSpacing
	if categoryWidth < tableMinCategoryWidth {
		categoryWidth = tableMinCategoryWidth
	}
	return summaryColumnWidths{Category: categoryWidth, Count: countWidth, Amount: amountWidth}
}

// renderSummaryRow renders a single summary row (category, count, amount)
func (m model) renderSummaryRow(cat types.CategorySummary, widths summaryColumnWidths, isSelected bool) string {
	formattedAmount := formatAmountWithCommas(cat.Total)
	line := fmt.Sprintf("%-*s  %*d  %*s", widths.Category, cat.Category, widths.Count, cat.Count, widths.Amount, formattedAmount)
	if isSelected {
		return m.styles.Selected.Render(line)
	}
	return m.styles.Line.Render(line)
}

// renderSummaryTotalLine renders the "Total" row at the bottom of the summary table
func (m model) renderSummaryTotalLine(widths summaryColumnWidths, total float64) string {
	formattedTotal := formatAmountWithCommas(total)
	return fmt.Sprintf("%-*s  %*s  %*s", widths.Category, "Total", widths.Count, "", widths.Amount, formattedTotal)
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
	title := m.formatExpensesAndAddBoxTitle(borderColor)

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
	expensesHeight := m.calculateExpensesBoxHeight()
	boxHeight := m.height - titleBoxHeight - expensesHeight

	var content strings.Builder

	if len(m.expenses) == 0 {
		content.WriteString(m.styles.Muted.Render("No expenses to summarize"))
	} else {
		tableWidth := m.width - tableBorderPadding
		content.WriteString(m.buildSummaryTableHeader(tableWidth))

		maxRows := boxHeight - summaryBoxHeaderRows
		if maxRows < 1 {
			maxRows = 1
		}

		visibleSummary := m.summary
		if m.summaryScrollOffset < len(visibleSummary) {
			visibleSummary = visibleSummary[m.summaryScrollOffset:]
		}

		widths := m.calculateSummaryColumnWidths(tableWidth)
		rowCount := 0
		for i, cat := range visibleSummary {
			if rowCount >= maxRows {
				break
			}
			actualRowIndex := m.summaryScrollOffset + i
			isRowSelected := m.isSelected(summaryBox) && actualRowIndex == m.summarySelectedRow
			content.WriteString(m.renderSummaryRow(cat, widths, isRowSelected) + "\n")
			rowCount++
		}

		separator := strings.Repeat("─", tableWidth)
		content.WriteString(m.styles.Muted.Render(separator) + "\n")
		content.WriteString(m.styles.Header.Render(m.renderSummaryTotalLine(widths, m.total)))
	}

	isSelected := m.isSelected(summaryBox)
	borderColor := m.styles.Theme.Border
	if isSelected {
		borderColor = m.styles.Theme.Primary
	}
	title := m.formatSummaryTitle(borderColor, isSelected)

	return m.styles.DrawBorderWithHeightAndTitleBold(
		content.String(),
		m.width,
		boxHeight,
		RoundedBorderChars(),
		borderColor,
		title,
		false,
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

// calculateExpensesBoxHeight calculates height for the stats box (middle box)
func (m model) calculateExpensesBoxHeight() int {
	// Remaining height after title box
	remainingHeight := m.height - titleBoxHeight

	// Give stats box half of remaining space (rounded down)
	// The expenses box will get all remaining space to fill terminal completely
	return remainingHeight / boxHeightDivisor
}

func (m model) calculateAddBoxHeight() int {
	// Remaining height after title box
	remainingHeight := m.height - titleBoxHeight

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

// calculateOverlayColumnWidths computes column widths for the overlay table (Date, Description, Amount)
func (m model) calculateOverlayColumnWidths(tableWidth int) overlayColumnWidths {
	dateWidth := tableDateWidth
	amountWidth := tableAmountWidth
	spacing := tableColumnSpacing
	totalSpacing := spacing * tableColumnGapsOverlay
	descriptionWidth := tableWidth - dateWidth - totalSpacing - amountWidth
	if descriptionWidth < tableMinDescWidth {
		descriptionWidth = tableMinDescWidth
	}
	if dateWidth+descriptionWidth+amountWidth+totalSpacing > tableWidth {
		dateWidth = tableWidth - descriptionWidth - amountWidth - totalSpacing
		if dateWidth < tableMinDescWidth {
			dateWidth = tableMinDescWidth
		}
	}
	return overlayColumnWidths{Date: dateWidth, Description: descriptionWidth, Amount: amountWidth}
}

// buildOverlayTableHeader returns the overlay table header and separator
func (m model) buildOverlayTableHeader(tableWidth int, widths overlayColumnWidths) string {
	header := fmt.Sprintf("%-*s  %-*s  %*s", widths.Date, "Date", widths.Description, "Description", widths.Amount, "Amount")
	separator := strings.Repeat("─", tableWidth)
	return m.styles.Header.Render(header) + "\n" + m.styles.Muted.Render(separator) + "\n"
}

// renderOverlayExpenseRow renders a single expense row in the overlay (no selection highlight)
func (m model) renderOverlayExpenseRow(expense types.Expense, widths overlayColumnWidths) string {
	localDate := expense.Date.Local()
	formattedAmount := formatAmountWithCommas(expense.Amount)
	datePart := m.styles.Line.Width(widths.Date).Align(lipgloss.Left).Render(localDate.Format("2006-01-02"))
	descriptionPart := m.styles.Line.Width(widths.Description).Align(lipgloss.Left).Render(expense.Description)
	amountPart := m.styles.Line.Width(widths.Amount).Align(lipgloss.Right).Render(formattedAmount)
	spacingStr := m.styles.Line.Render("  ")
	return datePart + spacingStr + descriptionPart + spacingStr + amountPart
}

// renderOverlay renders the overlay showing expenses for the selected category
func (m model) renderOverlay() string {
	if m.summarySelectedRow < 0 || m.summarySelectedRow >= len(m.summary) {
		return m.styles.Muted.Render("No category selected")
	}

	selectedCategory := m.summary[m.summarySelectedRow].Category
	filteredExpenses := m.filterExpensesByCategory(selectedCategory)

	if len(filteredExpenses) == 0 {
		content := fmt.Sprintf("No expenses found for category: %s", selectedCategory)
		return m.styles.DrawBorderWithHeightAndTitle(
			content,
			overlayMinWidth,
			overlayMinHeight,
			RoundedBorderChars(),
			m.styles.Theme.Primary,
			selectedCategory,
		)
	}

	overlayWidth := m.width - overlaySideMargin
	if overlayWidth < overlayMinWidth {
		overlayWidth = overlayMinWidth
	}
	if overlayWidth > overlayMaxWidth {
		overlayWidth = overlayMaxWidth
	}
	tableWidth := overlayWidth - tableBorderPadding
	widths := m.calculateOverlayColumnWidths(tableWidth)

	var content strings.Builder
	content.WriteString(m.buildOverlayTableHeader(tableWidth, widths))

	maxRows := overlayMaxRows
	if len(filteredExpenses) < maxRows {
		maxRows = len(filteredExpenses)
	}
	for i := 0; i < maxRows; i++ {
		content.WriteString(m.renderOverlayExpenseRow(filteredExpenses[i], widths) + "\n")
	}

	if len(filteredExpenses) > maxRows {
		remaining := len(filteredExpenses) - maxRows
		content.WriteString("\n")
		content.WriteString(m.styles.Muted.Render(fmt.Sprintf("... and %d more", remaining)))
	}
	content.WriteString("\n")
	content.WriteString(m.styles.Muted.Render("Press Space or Esc to close"))

	overlayHeight := maxRows + overlayHeaderRows
	if overlayHeight < overlayMinHeightFallback {
		overlayHeight = overlayMinHeightFallback
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
