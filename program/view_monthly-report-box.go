package program

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyawphyothu/sana/types"
)

// monthlyReportColumnWidths holds column widths for the monthly report table
type monthlyReportColumnWidths struct {
	Month  int
	Amount int
}

// renderMonthlyReportBox creates the monthly report section (third box)
func (m model) renderMonthlyReportBox() string {
	expensesHeight := m.calculateExpensesBoxHeight()
	boxHeight := m.ui.height - titleBoxHeight - expensesHeight

	boxWidth := m.ui.width - (m.ui.width / 2)

	var content strings.Builder

	if len(m.data.monthlyReport) == 0 {
		content.WriteString(m.styles.Muted.Render("No monthly report data"))
	} else {
		tableWidth := boxWidth - tableBorderPadding
		maxRows := boxHeight - monthlyReportBoxHeaderRows
		if maxRows < 1 {
			maxRows = 1
		}
		widths := m.calculateMonthlyReportColumnWidths(tableWidth)
		config := TableConfig{
			TableWidth:       tableWidth,
			Header:           m.buildMonthlyReportTableHeader(tableWidth),
			MaxRows:          maxRows,
			TotalRows:        len(m.data.monthlyReport),
			ScrollOffset:     m.ui.monthlyReportList.ScrollOffset(),
			SelectedRowIndex: m.ui.monthlyReportList.SelectedRow(),
			HasFocus:         m.isSelected(monthlyReportBox),
		}
		renderRow := func(globalRowIndex int, isSelected bool) string {
			return m.renderMonthlyReportRow(m.data.monthlyReport[globalRowIndex], widths, isSelected)
		}
		content.WriteString(m.renderTableBody(config, renderRow))
	}

	isSelected := m.isSelected(monthlyReportBox)
	borderColor := m.styles.Theme.Border
	if isSelected {
		borderColor = m.styles.Theme.Primary
	}
	title := m.formatMonthlyReportTitle(isSelected)

	return m.styles.DrawBorder(content.String(), BorderOptions{
		Width:       boxWidth,
		Height:      boxHeight,
		Title:       title,
		BorderChars: RoundedBorderChars(),
		Color:       borderColor,
	})
}

// buildMonthlyReportTableHeader returns the table header and separator for the monthly report box
func (m model) buildMonthlyReportTableHeader(tableWidth int) string {
	widths := m.calculateMonthlyReportColumnWidths(tableWidth)
	header := fmt.Sprintf("%-*s  %*s", widths.Month, "Month", widths.Amount, "Total")
	separator := strings.Repeat("─", tableWidth)
	return m.styles.Header.Render(header) + "\n" + m.styles.Muted.Render(separator) + "\n"
}

// renderMonthlyReportRow renders a single monthly report row (month, total)
func (m model) renderMonthlyReportRow(report types.MonthlyReport, widths monthlyReportColumnWidths, isSelected bool) string {
	formattedAmount := formatAmountWithCommas(report.Total)
	starPart := " "
	starStyle := m.styles.Line
	if sameMonth(report.Month, m.ui.activeMonth) {
		starPart = starStyle.Foreground(m.styles.Theme.Success).Render("*")
	} else {
		starPart = starStyle.Render(" ")
	}
	line := fmt.Sprintf("%-*s  %*s", widths.Month, report.Month.Format("2006-01"), widths.Amount, formattedAmount)
	if isSelected {
		return starPart + m.styles.Selected.Render(line)
	}
	return starPart + m.styles.Line.Render(line)
}

// calculateMonthlyReportColumnWidths computes column widths for the monthly report table
func (m model) calculateMonthlyReportColumnWidths(tableWidth int) monthlyReportColumnWidths {
	starWidth := 1
	monthWidth := 10
	amountWidth := tableWidth - starWidth - monthWidth - tableColumnSpacing
	return monthlyReportColumnWidths{Month: monthWidth, Amount: amountWidth}
}

// sameMonth checks if two months are the same
func sameMonth(a, b time.Time) bool {
	return a.Month() == b.Month() && a.Year() == b.Year()
}
