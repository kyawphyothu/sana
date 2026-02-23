package program

import (
	"strings"
	"testing"
	"time"

	"github.com/kyawphyothu/sana/types"
)

func TestRenderTableBody(t *testing.T) {
	m := minimalModelWithStyles()
	config := TableConfig{
		TableWidth:       60,
		Header:           "Date  | Desc  | Amount\n",
		MaxRows:          3,
		TotalRows:        5,
		ScrollOffset:     0,
		SelectedRowIndex: 1,
		HasFocus:         true,
		Footer:           "Total: 100",
	}
	rowCount := 0
	got := m.renderTableBody(config, func(globalRowIndex int, isSelected bool) string {
		rowCount++
		if isSelected && globalRowIndex != 1 {
			t.Errorf("expected isSelected only for row 1, got row %d", globalRowIndex)
		}
		return "row" + []string{"0", "1", "2", "3", "4"}[globalRowIndex]
	})
	if rowCount != 3 {
		t.Errorf("renderRow should be called 3 times (MaxRows), got %d", rowCount)
	}
	if !strings.Contains(got, config.Header) {
		t.Error("output should contain header")
	}
	if !strings.Contains(got, "Total: 100") {
		t.Error("output should contain footer")
	}
	if !strings.Contains(got, "row0") || !strings.Contains(got, "row1") || !strings.Contains(got, "row2") {
		t.Errorf("output should contain row0, row1, row2; got %q", got)
	}
}

func TestRenderTableBodyNoFooter(t *testing.T) {
	m := minimalModelWithStyles()
	config := TableConfig{
		TableWidth: 40,
		Header:     "H\n",
		MaxRows:    2,
		TotalRows:  2,
		Footer:     "",
	}
	got := m.renderTableBody(config, func(i int, _ bool) string { return "x" })
	if strings.Contains(got, "Total") {
		t.Error("no footer should be rendered when Footer is empty")
	}
	if !strings.Contains(got, "H") {
		t.Error("output should contain header")
	}
}

func TestCalculateOverlayColumnWidths(t *testing.T) {
	m := minimalModelWithStyles()
	widths := m.calculateOverlayColumnWidths(80)
	if widths.Date <= 0 || widths.Description <= 0 || widths.Amount <= 0 {
		t.Errorf("widths should be positive: %+v", widths)
	}
	total := widths.Date + widths.Description + widths.Amount + (tableColumnSpacing * tableColumnGapsOverlay)
	if total > 80+10 {
		t.Errorf("total width %d should be roughly within tableWidth 80", total)
	}
}

func TestBuildOverlayTableHeader(t *testing.T) {
	m := minimalModelWithStyles()
	widths := m.calculateOverlayColumnWidths(70)
	got := m.buildOverlayTableHeader(70, widths)
	if !strings.Contains(got, "Date") || !strings.Contains(got, "Description") || !strings.Contains(got, "Amount") {
		t.Errorf("header should contain Date, Description, Amount: %q", got)
	}
	if !strings.Contains(got, "─") {
		t.Error("header should contain separator")
	}
}

func TestRenderOverlayExpenseRow(t *testing.T) {
	m := minimalModelWithStyles()
	widths := overlayColumnWidths{Date: 10, Description: 20, Amount: 10}
	exp := types.Expense{
		ID:          1,
		Date:        time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Amount:      1234.56,
		Description: "Lunch",
		Type:        types.ExpenseTypeFood,
	}
	got := m.renderOverlayExpenseRow(exp, widths)
	if got == "" {
		t.Fatal("renderOverlayExpenseRow returned empty string")
	}
	if !strings.Contains(got, "2025-01-15") {
		t.Error("output should contain formatted date")
	}
	if !strings.Contains(got, "1,234.56") {
		t.Error("output should contain formatted amount")
	}
	if !strings.Contains(got, "Lunch") {
		t.Error("output should contain description")
	}
}
