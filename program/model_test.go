package program

import (
	"testing"
	"time"

	"github.com/kyawphyothu/sana/types"
)

func TestIsSelected(t *testing.T) {
	m := model{ui: uiState{selected: expensesBox}}
	if !m.isSelected(expensesBox) {
		t.Error("isSelected(expensesBox) should be true")
	}
	if m.isSelected(summaryBox) {
		t.Error("isSelected(summaryBox) should be false")
	}
	m.ui.selected = addBox
	if !m.isSelected(addBox) {
		t.Error("isSelected(addBox) should be true")
	}
}

func TestMoveRowUp(t *testing.T) {
	m := model{
		data: expenseData{
			expenses:      make([]types.Expense, 5),
			summary:       make([]types.CategorySummary, 3),
			monthlyReport: make([]types.MonthlyReport, 4),
		},
		ui: uiState{
			selected: expensesBox,
			expensesList: scrollableList{selectedRow: 2, scrollOffset: 0, length: 5},
		},
	}
	m.moveRowUp()
	if m.ui.expensesList.SelectedRow() != 1 {
		t.Errorf("moveRowUp expenses: selected = %d, want 1", m.ui.expensesList.SelectedRow())
	}
	m.moveRowUp()
	if m.ui.expensesList.SelectedRow() != 0 {
		t.Errorf("moveRowUp expenses: selected = %d, want 0", m.ui.expensesList.SelectedRow())
	}
	m.moveRowUp() // at top, no-op
	if m.ui.expensesList.SelectedRow() != 0 {
		t.Errorf("moveRowUp at top: selected = %d, want 0", m.ui.expensesList.SelectedRow())
	}
}

func TestMoveRowDown(t *testing.T) {
	m := model{
		data: expenseData{
			expenses: make([]types.Expense, 5),
		},
		ui: uiState{
			selected:    expensesBox,
			expensesList: scrollableList{selectedRow: 0, length: 5},
		},
	}
	m.moveRowDown(3)
	if m.ui.expensesList.SelectedRow() != 1 {
		t.Errorf("moveRowDown: selected = %d, want 1", m.ui.expensesList.SelectedRow())
	}
	m.moveRowDown(3)
	m.moveRowDown(3)
	if m.ui.expensesList.SelectedRow() != 3 {
		t.Errorf("moveRowDown: selected = %d, want 3", m.ui.expensesList.SelectedRow())
	}
}

func TestMoveRowToTop(t *testing.T) {
	m := model{
		data: expenseData{expenses: make([]types.Expense, 5)},
		ui: uiState{
			selected:     expensesBox,
			expensesList: scrollableList{selectedRow: 3, scrollOffset: 2, length: 5},
		},
	}
	m.moveRowToTop()
	if m.ui.expensesList.SelectedRow() != 0 || m.ui.expensesList.ScrollOffset() != 0 {
		t.Errorf("moveRowToTop: selected=%d scroll=%d", m.ui.expensesList.SelectedRow(), m.ui.expensesList.ScrollOffset())
	}
}

func TestMoveRowToBottom(t *testing.T) {
	m := model{
		data: expenseData{expenses: make([]types.Expense, 10)},
		ui: uiState{
			selected:     expensesBox,
			expensesList: scrollableList{selectedRow: 0, length: 10},
		},
	}
	m.moveRowToBottom(3)
	if m.ui.expensesList.SelectedRow() != 9 {
		t.Errorf("moveRowToBottom: selected = %d, want 9", m.ui.expensesList.SelectedRow())
	}
}

func TestResetRowSelection(t *testing.T) {
	m := model{
		ui: uiState{
			expensesList:      scrollableList{selectedRow: 2, scrollOffset: 1, length: 5},
			summaryList:       scrollableList{selectedRow: 1, scrollOffset: 0, length: 3},
			monthlyReportList: scrollableList{selectedRow: 4, scrollOffset: 2, length: 10},
		},
	}
	m.resetRowSelection()
	if m.ui.expensesList.SelectedRow() != 0 || m.ui.expensesList.ScrollOffset() != 0 {
		t.Error("expensesList should be reset")
	}
	if m.ui.summaryList.SelectedRow() != 0 || m.ui.summaryList.ScrollOffset() != 0 {
		t.Error("summaryList should be reset")
	}
	if m.ui.monthlyReportList.SelectedRow() != 0 || m.ui.monthlyReportList.ScrollOffset() != 0 {
		t.Error("monthlyReportList should be reset")
	}
}

func TestClampSelections(t *testing.T) {
	now := time.Now()
	m := model{
		data: expenseData{
			expenses:      []types.Expense{{ID: 1, Date: now}, {ID: 2, Date: now}},
			summary:       []types.CategorySummary{{Category: "Food"}, {Category: "Transport"}},
			monthlyReport: []types.MonthlyReport{{Month: now}},
		},
		ui: uiState{
			expensesList:      scrollableList{selectedRow: 5, length: 2},
			summaryList:       scrollableList{selectedRow: 3, length: 2},
			monthlyReportList: scrollableList{selectedRow: 2, length: 1},
		},
	}
	m.clampSelections()
	if m.ui.expensesList.SelectedRow() != 1 {
		t.Errorf("expensesList selected should be clamped to 1, got %d", m.ui.expensesList.SelectedRow())
	}
	if m.ui.summaryList.SelectedRow() != 1 {
		t.Errorf("summaryList selected should be clamped to 1, got %d", m.ui.summaryList.SelectedRow())
	}
	if m.ui.monthlyReportList.SelectedRow() != 0 {
		t.Errorf("monthlyReportList selected should be clamped to 0, got %d", m.ui.monthlyReportList.SelectedRow())
	}
}

func TestClampSelectionsEmptyDataResets(t *testing.T) {
	m := model{
		data: expenseData{
			expenses:      nil,
			summary:       nil,
			monthlyReport: nil,
		},
		ui: uiState{
			expensesList:      scrollableList{selectedRow: 2, scrollOffset: 1, length: 0},
			summaryList:       scrollableList{selectedRow: 1, length: 0},
			monthlyReportList: scrollableList{selectedRow: 1, length: 0},
		},
	}
	m.clampSelections()
	if m.ui.expensesList.SelectedRow() != 0 || m.ui.expensesList.ScrollOffset() != 0 {
		t.Error("empty expenses should reset list")
	}
	if m.ui.summaryList.SelectedRow() != 0 {
		t.Error("empty summary should reset list")
	}
	if m.ui.monthlyReportList.SelectedRow() != 0 {
		t.Error("empty monthlyReport should reset list")
	}
}
