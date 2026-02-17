package program

import (
	"fmt"
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

// overlayColumnWidths holds column widths for the category overlay table
type overlayColumnWidths struct {
	Date        int
	Description int
	Amount      int
}

// TableConfig holds shared table layout and scroll/selection state for the generic table renderer.
// Header and optional Footer are pre-rendered by the caller (table-specific formatting).
// Row rendering is done via the renderRow callback so each table can keep its own row logic.
type TableConfig struct {
	TableWidth       int
	Header           string
	MaxRows          int
	TotalRows        int
	ScrollOffset     int
	SelectedRowIndex int
	HasFocus         bool
	Footer           string // optional, e.g. summary "Total" line
}

// renderTableBody renders header, separator, visible rows (with scroll/selection), and optional footer.
// renderRow(globalRowIndex, isSelected) is called for each visible row; the caller maps index to data.
func (m model) renderTableBody(config TableConfig, renderRow func(globalRowIndex int, isSelected bool) string) string {
	var b strings.Builder
	separator := strings.Repeat("─", config.TableWidth)
	// Header already includes header line + separator + newline from build*TableHeader
	b.WriteString(config.Header)

	start := config.ScrollOffset
	end := min(config.ScrollOffset+config.MaxRows, config.TotalRows)
	for i := start; i < end; i++ {
		isSelected := config.HasFocus && i == config.SelectedRowIndex
		b.WriteString(renderRow(i, isSelected))
		b.WriteString("\n")
	}

	if config.Footer != "" {
		b.WriteString(m.styles.Muted.Render(separator))
		b.WriteString("\n")
		b.WriteString(config.Footer)
	}
	return b.String()
}

// View renders the entire UI
func (m model) View() tea.View {
	if m.ui.width == 0 || m.ui.height == 0 {
		res := tea.NewView("Loading...")
		res.AltScreen = true
		return res
	}

	// Check if terminal is too small
	if m.ui.width < minWidth || m.ui.height < minHeight {
		res := tea.NewView(m.renderTooSmallMessage())
		res.AltScreen = true
		return res
	}

	// Build the three sections
	titleBox := m.renderTitleBox()

	if m.ui.selected == addBox {
		res := tea.NewView(lipgloss.JoinVertical(
			lipgloss.Left,
			titleBox,
			m.renderAddBox(),
		))
		res.AltScreen = true
		res.Cursor = m.fixedCursor()
		return res
	}

	expensesBox := m.renderExpensesBox()

	summaryBox := m.renderSummaryBox()

	monthlyReportBox := m.renderMonthlyReportBox()

	summaryAndMonthlyReportBox := lipgloss.JoinHorizontal(
		lipgloss.Left,
		summaryBox,
		monthlyReportBox,
	)

	// Stack vertically
	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		titleBox,
		expensesBox,
		summaryAndMonthlyReportBox,
	)

	// If overlay is visible, layer it on top of main content using Canvas
	if m.ui.overlay != overlayNone {
		overlay := m.renderOverlay()

		// Create canvas with layers
		mainLayer := lipgloss.NewLayer(mainContent).
			Width(m.ui.width).
			Height(m.ui.height).
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

		overlayX := (m.ui.width - overlayWidth) / 2
		overlayY := (m.ui.height - overlayHeight) / 2
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
	res.Cursor = m.fixedCursor()
	return res
}

// renderTitleBox creates the top title section with Sana figlet
func (m model) renderTitleBox() string {
	content := m.styles.Title.Render(sanaFiglet)

	return m.styles.DrawBorder(content, BorderOptions{
		Width:       m.ui.width,
		Height:      titleBoxHeight,
		Title:       "Sana",
		BorderChars: DoubleBorderChars(),
		Color:       m.styles.Theme.Border,
	})
}

// renderAddBox creates the add expense form section (middle box)
func (m model) renderAddBox() string {
	boxHeight := m.calculateAddBoxHeight()

	// Update prompt styles based on focus before rendering
	m.updatePromptStyles()

	helpStyle := m.styles.Muted

	// Form rows (each textinput has its own prompt)
	rows := []string{
		m.form.typeField.View(),
		m.form.amount.View(),
		m.form.description.View(),
		m.form.date.View(),
	}
	formContent := strings.Join(rows, "\n\n")

	helpText := helpStyle.Render("↑/↓: move • Tab: autocomplete • Enter: submit • Esc: cancel")

	// List available expense types
	typeLabels := types.ExpenseTypeSuggestions()
	typesList := "Available types: " + strings.Join(typeLabels, ", ")
	typesText := helpStyle.Render(typesList)

	content := formContent + "\n\n" + helpText + "\n" + typesText

	// Show validation/creation error below form if any
	if m.ui.err != nil && m.ui.selected == addBox {
		content += "\n\n" + m.styles.Line.Foreground(m.styles.Theme.Error).Render("Error: "+m.ui.err.Error())
	}

	// Choose border color based on selection
	isSelected := m.isSelected(addBox)
	borderColor := m.styles.Theme.Border
	if isSelected {
		borderColor = m.styles.Theme.Primary
	}

	// Format title with bold for selected part
	title := m.formatExpensesAndAddBoxTitle(borderColor)

	return m.styles.DrawBorder(content, BorderOptions{
		Width:       m.ui.width,
		Height:      boxHeight,
		Title:       title,
		BorderChars: RoundedBorderChars(),
		Color:       borderColor,
	})
}

// renderTooSmallMessage creates small terminal message when terminal is too small
func (m model) renderTooSmallMessage() string {
	message := fmt.Sprintf(
		"Terminal too small!\n\nMinimum size: %dx%d\nCurrent size: %dx%d\n\nPlease resize your terminal.",
		minWidth, minHeight, m.ui.width, m.ui.height,
	)
	return m.styles.Parent.
		Width(m.ui.width).
		Height(m.ui.height).
		Align(lipgloss.Center).
		Render(m.styles.Line.Foreground(m.styles.Theme.Error).Render(message))
}

// fixedCursor returns a non-blinking cursor pinned at (0,0) so that
// iTerm's cursor guide stays on the top row instead of jumping to
// whichever cells the v2 renderer last updated.
func (m model) fixedCursor() *tea.Cursor {
	c := tea.NewCursor(0, m.ui.height)
	c.Blink = false
	c.Shape = tea.CursorUnderline
	return c
}

// Helper functions

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
	desc := expense.Description
	if len(desc) > widths.Description {
		desc = desc[:widths.Description-descTruncateSuffix] + "..."
	}

	localDate := expense.Date.Local()
	formattedAmount := formatAmountWithCommas(expense.Amount)
	datePart := m.styles.Line.Width(widths.Date).Align(lipgloss.Left).Render(localDate.Format("2006-01-02"))
	descriptionPart := m.styles.Line.Width(widths.Description).Align(lipgloss.Left).Render(desc)
	amountPart := m.styles.Line.Width(widths.Amount).Align(lipgloss.Right).Render(formattedAmount)
	spacingStr := m.styles.Line.Render("  ")
	return datePart + spacingStr + descriptionPart + spacingStr + amountPart
}

// renderOverlay dispatches to the appropriate overlay renderer based on the active overlay kind.
func (m model) renderOverlay() string {
	switch m.ui.overlay {
	case overlayCategoryDetail:
		return m.renderCategoryDetailOverlay()
	case overlayConfirmDelete:
		return m.renderConfirmDeleteOverlay()
	}
	return ""
}

// renderCategoryDetailOverlay renders the overlay showing expenses for the selected category.
func (m model) renderCategoryDetailOverlay() string {
	if m.ui.summaryList.SelectedRow() < 0 || m.ui.summaryList.SelectedRow() >= len(m.data.summary) {
		return m.styles.Muted.Render("No category selected")
	}

	selectedCategory := m.data.summary[m.ui.summaryList.SelectedRow()].Category
	filteredExpenses := m.filterExpensesByCategory(selectedCategory)

	categoryColor := CategoryColor(selectedCategory)
	categoryStyle := lipgloss.NewStyle().Foreground(categoryColor).Bold(true).Background(m.styles.Theme.Background)

	if len(filteredExpenses) == 0 {
		content := fmt.Sprintf("No expenses found for category: %s", selectedCategory)
		return m.styles.DrawBorder(content, BorderOptions{
			Width:       overlayMinWidth,
			Height:      overlayMinHeight,
			Title:       categoryStyle.Render(selectedCategory),
			BorderChars: RoundedBorderChars(),
			Color:       m.styles.Theme.Primary,
		})
	}

	overlayWidth := m.ui.width - overlaySideMargin
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

	return m.styles.DrawBorder(content.String(), BorderOptions{
		Width:       overlayWidth,
		Height:      overlayHeight,
		Title:       categoryStyle.Render(selectedCategory),
		BorderChars: RoundedBorderChars(),
		Color:       m.styles.Theme.Primary,
	})
}

// renderConfirmDeleteOverlay renders the confirmation overlay for deleting an expense.
func (m model) renderConfirmDeleteOverlay() string {
	selectedIdx := m.ui.expensesList.SelectedRow()
	if selectedIdx < 0 || selectedIdx >= len(m.data.expenses) {
		return m.styles.Muted.Render("No expense selected")
	}

	expense := m.data.expenses[selectedIdx]
	localDate := expense.Date.Local()
	formattedAmount := formatAmountWithCommas(expense.Amount)

	var content strings.Builder
	content.WriteString(m.styles.Line.Render(fmt.Sprintf("Date:     %s", localDate.Format("2006-01-02 15:04:05"))))
	content.WriteString("\n")
	content.WriteString(m.styles.Line.Render(fmt.Sprintf("Type:     %s", expense.Type.String())))
	content.WriteString("\n")
	content.WriteString(m.styles.Line.Render(fmt.Sprintf("Amount:   %s", formattedAmount)))
	content.WriteString("\n")
	desc := expense.Description
	if len(desc) > confirmDeleteOverlayWidth-tableBorderPadding-10 {
		desc = desc[:confirmDeleteOverlayWidth-tableBorderPadding-10-descTruncateSuffix] + "..."
	}
	content.WriteString(m.styles.Line.Render(fmt.Sprintf("Desc:     %s", desc)))
	content.WriteString("\n\n")

	warningStyle := m.styles.Line.Foreground(m.styles.Theme.Error).Bold(true)
	content.WriteString(warningStyle.Render("Delete this expense?"))
	content.WriteString("\n\n")

	content.WriteString(m.styles.Muted.Render("d/Enter: delete • Esc: cancel"))

	return m.styles.DrawBorder(content.String(), BorderOptions{
		Width:       confirmDeleteOverlayWidth,
		Height:      confirmDeleteOverlayHeight,
		Title:       "Confirm Delete",
		BorderChars: RoundedBorderChars(),
		Color:       m.styles.Theme.Error,
	})
}
