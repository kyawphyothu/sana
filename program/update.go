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
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = expensesBox
		return m, m.reloadAllData()

	case expenseDeletedMsg:
		if msg.Err != nil {
			m.ui.err = msg.Err
			m.ui.overlay = overlayConfirmDelete
			return m, nil
		}
		m.ui.err = nil
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = expensesBox
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
	if m.ui.selected == expensesBox {
		return m.handleExpensesBoxKeys(msg)
	}
	if m.ui.selected == monthlyReportBox {
		return m.handleMonthlyReportBoxKeys(msg)
	}
	if m.ui.selected == summaryBox {
		return m.handleSummaryBoxKeys(msg)
	}
	if m.ui.overlay != overlayNone {
		return m.handleOverlayKeys(msg)
	}

	return m, nil
}

// handleExpensesBoxKeys handles keys for the expenses box.
func (m model) handleExpensesBoxKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "d":
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = confirmDeleteOverlay
		m.ui.overlay = overlayConfirmDelete
		return m, nil
	case "?":
		return m.help()
	case "q":
		return m, tea.Quit
	case "r":
		m.resetRowSelection()
		return m, loadMonthData(m.db, m.ui.activeMonth)
	case "a":
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = addBox
		return m, nil
	case "s":
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = summaryBox
		return m, nil
	case "m":
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = monthlyReportBox
		return m, nil
	case "j", "down":
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowDown(maxRows)
		return m, nil
	case "k", "up":
		m.moveRowUp()
		return m, nil
	case "g", "home":
		m.moveRowToTop()
		return m, nil
	case "G", "end":
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowToBottom(maxRows)
		return m, nil
	}
	return m, nil
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
		m.ui.selected = m.ui.previousSelected
		m.ui.err = nil
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

// handleSummaryBoxKeys handles keys for the summary box.
func (m model) handleSummaryBoxKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "space":
		if m.ui.summaryList.SelectedRow() >= 0 && m.ui.summaryList.SelectedRow() < len(m.data.summary) {
			m.ui.selected = categoryDetailOverlay
			m.ui.overlay = overlayCategoryDetail
		}
		return m, nil
	case "?":
		return m.help()
	case "q":
		return m, tea.Quit
	case "r":
		m.resetRowSelection()
		return m, loadMonthData(m.db, m.ui.activeMonth)
	case "a":
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = addBox
		return m, nil
	case "e":
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = expensesBox
		return m, nil
	case "m":
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = monthlyReportBox
		return m, nil
	case "j", "down":
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowDown(maxRows)
		return m, nil
	case "k", "up":
		m.moveRowUp()
		return m, nil
	case "g", "home":
		m.moveRowToTop()
		return m, nil
	case "G", "end":
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowToBottom(maxRows)
		return m, nil
	}

	return m, nil
}

// handleMonthlyReportBoxKeys handles keys for the monthly report box.
func (m model) handleMonthlyReportBoxKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		selectedIdx := m.ui.monthlyReportList.SelectedRow()
		if selectedIdx >= 0 && selectedIdx < len(m.data.monthlyReport) {
			m.ui.activeMonth = m.data.monthlyReport[selectedIdx].Month
			return m, loadMonthData(m.db, m.ui.activeMonth)
		}
		return m, nil
	case "?":
		return m.help()
	case "q":
		return m, tea.Quit
	case "r":
		m.resetRowSelection()
		return m, loadMonthData(m.db, m.ui.activeMonth)
	case "a":
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = addBox
		return m, nil
	case "e":
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = expensesBox
		return m, nil
	case "s":
		m.ui.previousSelected = m.ui.selected
		m.ui.selected = summaryBox
		return m, nil
	case "j", "down":
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowDown(maxRows)
		return m, nil
	case "k", "up":
		m.moveRowUp()
		return m, nil
	case "g", "home":
		m.moveRowToTop()
		return m, nil
	case "G", "end":
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowToBottom(maxRows)
		return m, nil
	}

	return m, nil
}

// handleOverlayKeys dispatches key handling to the active overlay.
func (m model) handleOverlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.ui.overlay {
	case overlayCategoryDetail:
		return m.handleCategoryOverlayKeys(msg)
	case overlayConfirmDelete:
		return m.handleConfirmDeleteOverlayKeys(msg)
	case overlayHelp:
		return m.handleHelpOverlayKeys(msg)
	}
	return m, nil
}

// handleCategoryOverlayKeys handles keys for the category detail overlay.
func (m model) handleCategoryOverlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "space":
		m.ui.overlay = overlayNone
		m.ui.selected = summaryBox
		return m, nil
	case "esc":
		m.ui.selected = summaryBox
		m.ui.overlay = overlayNone
		m.ui.err = nil
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
		m.ui.selected = expensesBox
		m.ui.overlay = overlayNone
		return m, nil
	case "esc":
		m.ui.selected = expensesBox
		m.ui.overlay = overlayNone
		m.ui.err = nil
		return m, nil
	}
	return m, nil
}

// handleHelpOverlayKeys handles keys for the help overlay.
func (m model) handleHelpOverlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "?":
		m.ui.selected = m.ui.previousSelected
		m.ui.overlay = overlayNone
		return m, nil
	case "q":
		return m, tea.Quit
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
