package program

import (
	"fmt"
	"strings"

	"github.com/kyawphyothu/sana/types"
)

// summaryColumnWidths holds column widths for the summary table
type summaryColumnWidths struct {
	Category int
	Count    int
	Amount   int
}

// renderSummaryBox creates the summary section grouped by category (third box)
func (m model) renderSummaryBox() string {
	expensesHeight := m.calculateExpensesBoxHeight()
	boxHeight := m.ui.height - titleBoxHeight - expensesHeight

	boxWidth := m.ui.width / 2

	var content strings.Builder

	if len(m.data.expenses) == 0 {
		content.WriteString(m.styles.Muted.Render("No expenses to summarize"))
	} else {
		tableWidth := boxWidth - tableBorderPadding
		maxRows := boxHeight - summaryBoxHeaderRows
		if maxRows < 1 {
			maxRows = 1
		}
		widths := m.calculateSummaryColumnWidths(tableWidth)
		footer := m.styles.Header.Render(m.renderSummaryTotalLine(widths, m.data.total))
		config := TableConfig{
			TableWidth:       tableWidth,
			Header:           m.buildSummaryTableHeader(tableWidth),
			MaxRows:          maxRows,
			TotalRows:        len(m.data.summary),
			ScrollOffset:     m.ui.summaryList.ScrollOffset(),
			SelectedRowIndex: m.ui.summaryList.SelectedRow(),
			HasFocus:         m.isSelected(summaryBox),
			Footer:           footer,
		}
		renderRow := func(globalRowIndex int, isSelected bool) string {
			return m.renderSummaryRow(m.data.summary[globalRowIndex], widths, isSelected)
		}
		content.WriteString(m.renderTableBody(config, renderRow))
	}

	isSelected := m.isSelected(summaryBox)
	borderColor := m.styles.Theme.Border
	if isSelected {
		borderColor = m.styles.Theme.Primary
	}
	title := m.formatSummaryTitle(borderColor, isSelected)

	return m.styles.DrawBorder(content.String(), BorderOptions{
		Width:       boxWidth,
		Height:      boxHeight,
		Title:       title,
		BorderChars: RoundedBorderChars(),
		Color:       borderColor,
	})
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
