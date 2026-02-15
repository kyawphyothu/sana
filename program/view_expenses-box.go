package program

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/kyawphyothu/sana/types"
)

// renderExpensesBox creates the expenses list table section (second box)
func (m model) renderExpensesBox() string {
	boxHeight := m.calculateExpensesBoxHeight()

	var content strings.Builder

	if len(m.data.expenses) == 0 {
		content.WriteString(m.styles.Muted.Render("No expenses yet"))
	} else {
		tableWidth := m.ui.width - tableBorderPadding
		maxRows := boxHeight - expensesBoxHeaderRows
		if maxRows < 1 {
			maxRows = 1
		}
		widths := m.calculateExpenseColumnWidths(tableWidth)
		config := TableConfig{
			TableWidth:       tableWidth,
			Header:           m.buildExpensesTableHeader(tableWidth),
			MaxRows:          maxRows,
			TotalRows:        len(m.data.expenses),
			ScrollOffset:     m.ui.expensesList.ScrollOffset(),
			SelectedRowIndex: m.ui.expensesList.SelectedRow(),
			HasFocus:         m.isSelected(expensesBox),
		}
		renderRow := func(globalRowIndex int, isSelected bool) string {
			return m.renderExpenseRow(m.data.expenses[globalRowIndex], widths, isSelected)
		}
		content.WriteString(m.renderTableBody(config, renderRow))
	}

	// Show error if any
	if m.ui.err != nil {
		content.WriteString("\n")
		errorStyle := m.styles.Line.Foreground(m.styles.Theme.Error)
		content.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.ui.err)))
	}

	// Choose border color based on selection
	isSelected := m.isSelected(expensesBox)
	borderColor := m.styles.Theme.Border
	if isSelected {
		borderColor = m.styles.Theme.Primary
	}

	// Format title with bold for selected part
	title := m.formatExpensesAndAddBoxTitle(borderColor)

	return m.styles.DrawBorder(content.String(), BorderOptions{
		Width:       m.ui.width,
		Height:      boxHeight,
		Title:       title,
		BorderChars: RoundedBorderChars(),
		Color:       borderColor,
	})
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
		categoryColor = CategoryColorSelected(categoryText)
	}
	categoryStyle := baseStyle.Foreground(categoryColor).Width(widths.Category).Align(lipgloss.Left)
	categoryPart := categoryStyle.Render(categoryText)
	amountPart := baseStyle.Width(widths.Amount).Align(lipgloss.Right).Render(formattedAmount)
	spacing := baseStyle.Render("  ")
	return datePart + spacing + descPart + spacing + categoryPart + spacing + amountPart
}
