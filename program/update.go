package program

import (
	tea "charm.land/bubbletea/v2"
	"github.com/kyawphyothu/sana/database"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ui.width = msg.Width
		m.ui.height = msg.Height
		return m, nil

	case monthDataLoadedMsg:
		if msg.Err != nil {
			m.ui.err = msg.Err
			return m, nil
		}
		m.ui.err = nil
		m.data.expenses = msg.Expenses
		m.data.summary = msg.Summary
		m.data.total = msg.Total
		m.clampSelections()
		return m, nil

	case monthlyReportLoadedMsg:
		if msg.Err != nil {
			m.ui.err = msg.Err
			return m, nil
		}
		m.ui.err = nil
		m.data.monthlyReport = msg.MonthlyReport
		m.clampSelections()
		return m, nil

	case expenseCreatedMsg:
		if msg.Err != nil {
			m.ui.err = msg.Err
			return m, nil
		}
		m.ui.err = nil
		m.addFormReset()
		m.ui.selected = expensesBox
		return m, m.reloadAllData()

	case expenseDeletedMsg:
		if msg.Err != nil {
			m.ui.err = msg.Err
			m.ui.overlay = overlayConfirmDelete
			return m, nil
		}
		m.ui.err = nil
		m.ui.overlay = overlayNone
		return m, m.reloadAllData()

	case formValidationErrMsg:
		m.ui.err = msg.Err
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	if m.ui.selected == addBox {
		return m.handleAddBoxKeys(msg)
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	}

	if m.ui.selected == monthlyReportBox {
		return m.handleMonthlyReportBoxKeys(msg)
	}
	if m.ui.overlay != overlayNone {
		return m.handleOverlayKeys(msg)
	}

	key := msg.Key()
	isSpace := key.Text == " " || key.Code == ' '
	if isSpace {
		if m.ui.selected == summaryBox && len(m.data.summary) > 0 {
			if m.ui.summaryList.SelectedRow() >= 0 && m.ui.summaryList.SelectedRow() < len(m.data.summary) {
				m.ui.overlay = overlayCategoryDetail
			}
		}
		return m, nil
	}

	return m.handleNavigationKeys(msg)
}

// handleAddBoxKeys handles form navigation and forwards keys to the focused add-form input.
// Call only when m.ui.selected == addBox.
func (m model) handleAddBoxKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.form.typeCompleted {
			m.addFormFocusNext()
			return m, nil
		}
		if m.hasMatchedSuggestions() {
			in := m.addFormInput()
			var cmd tea.Cmd
			*in, cmd = in.Update(msg)
			if m.isValueCompleteSuggestion() {
				m.form.typeCompleted = true
			}
			return m, cmd
		}
		m.addFormFocusNext()
		return m, nil
	case "down":
		m.addFormFocusNext()
		return m, nil
	case "shift+tab", "up":
		m.addFormFocusPrev()
		return m, nil
	case "enter":
		if cmd := m.addFormSubmit(); cmd != nil {
			return m, cmd
		}
		return m, nil
	case "esc":
		m.ui.err = nil
		m.ui.selected = expensesBox
		return m, nil
	}

	in := m.addFormInput()
	var cmd tea.Cmd
	*in, cmd = in.Update(msg)
	if m.form.focused == addFormType {
		m.form.typeCompleted = false
	}
	return m, cmd
}

func (m model) handleMonthlyReportBoxKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		selectedIdx := m.ui.monthlyReportList.SelectedRow()
		if selectedIdx >= 0 && selectedIdx < len(m.data.monthlyReport) {
			m.ui.activeMonth = m.data.monthlyReport[selectedIdx].Month
			return m, loadMonthData(m.db, m.ui.activeMonth)
		}
		return m, nil
	}
	return m.handleNavigationKeys(msg)
}

// handleOverlayKeys dispatches key handling to the active overlay.
func (m model) handleOverlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.ui.overlay {
	case overlayCategoryDetail:
		return m.handleCategoryOverlayKeys(msg)
	case overlayConfirmDelete:
		return m.handleConfirmDeleteOverlayKeys(msg)
	}
	return m, nil
}

// handleCategoryOverlayKeys handles keys for the category detail overlay.
func (m model) handleCategoryOverlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.Key()
	isSpace := key.Text == " " || key.Code == ' '
	if isSpace || msg.String() == "esc" {
		m.ui.overlay = overlayNone
		return m, nil
	}
	return m, nil
}

// handleConfirmDeleteOverlayKeys handles keys for the confirm delete overlay.
func (m model) handleConfirmDeleteOverlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "d", "enter":
		selectedIdx := m.ui.expensesList.SelectedRow()
		if selectedIdx >= 0 && selectedIdx < len(m.data.expenses) {
			expense := m.data.expenses[selectedIdx]
			db := m.db
			return m, func() tea.Msg {
				return expenseDeletedMsg{Err: database.DeleteExpense(db, expense.ID)}
			}
		}
		m.ui.overlay = overlayNone
		return m, nil
	case "esc":
		m.ui.err = nil
		m.ui.overlay = overlayNone
		return m, nil
	}
	return m, nil
}

// handleNavigationKeys handles box selection and row navigation for expenses/summary.
func (m model) handleNavigationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		m.resetRowSelection()
		return m, loadMonthData(m.db, m.ui.activeMonth)
	case "e":
		m.ui.selected = expensesBox
		return m, nil
	case "a":
		m.ui.selected = addBox
		return m, nil
	case "s":
		m.ui.selected = summaryBox
		return m, nil
	case "m":
		m.ui.selected = monthlyReportBox
		return m, nil
	case "j", "down":
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowDown(maxRows)
	case "k", "up":
		m.moveRowUp()
	case "g", "home":
		m.moveRowToTop()
	case "G", "end":
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowToBottom(maxRows)
	case "d":
		if m.ui.selected == expensesBox && len(m.data.expenses) > 0 {
			m.ui.overlay = overlayConfirmDelete
		}
	}
	return m, nil
}

// calculateMaxVisibleRows returns the max number of visible rows in the currently selected box
func (m model) calculateMaxVisibleRows() int {
	boxHeight := (m.ui.height - titleHeightForRows) / boxHeightDivisor

	switch m.ui.selected {
	case expensesBox:
		// Expenses box: boxHeight - borders - header - separator
		return boxHeight - expensesBoxRowOffset
	case addBox:
		// Add box doesn't have rows to navigate
		return 1
	case summaryBox:
		// Summary box: boxHeight - borders - header - separator - separator - total
		return boxHeight - summaryBoxRowOffset
	case monthlyReportBox:
		// Monthly report box: boxHeight - borders - header - separator
		return boxHeight - monthlyReportBoxRowOffset
	}
	return 1
}

// reloadAllData reloads all data for the current month and monthly report
func (m model) reloadAllData() tea.Cmd {
	return tea.Batch(loadMonthData(m.db, m.ui.activeMonth), loadMonthlyReportData(m.db))
}
