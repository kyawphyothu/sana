package program

import (
	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ui.width = msg.Width
		m.ui.height = msg.Height
		return m, nil

	case dataLoadedMsg:
		if msg.Err != nil {
			m.ui.err = msg.Err
			return m, nil
		}
		m.data.expenses = msg.Expenses
		m.data.summary = msg.Summary
		m.data.total = msg.Total
		return m, nil

	case expenseCreatedMsg:
		if msg.Err != nil {
			m.ui.err = msg.Err
			return m, nil
		}
		m.ui.err = nil
		m.addFormReset()
		m.ui.selected = expensesBox
		return m, loadData(m.db)

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
	if m.ui.showOverlay {
		return m.handleOverlayKeys(msg)
	}

	key := msg.Key()
	isSpace := key.Text == " " || key.Code == ' '
	if isSpace {
		if m.ui.selected == summaryBox && len(m.data.summary) > 0 {
			if m.ui.summaryList.SelectedRow() >= 0 && m.ui.summaryList.SelectedRow() < len(m.data.summary) {
				m.ui.showOverlay = !m.ui.showOverlay
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
		m.ui.selected = expensesBox
		return m, nil
	}

	in := m.addFormInput()
	var cmd tea.Cmd
	*in, cmd = in.Update(msg)
	if msg.String() != "tab" && msg.String() != "enter" && m.form.focused == addFormType {
		m.form.typeCompleted = false
	}
	return m, cmd
}

// handleOverlayKeys handles keys when the overlay is open (close, quit, or ignore).
// Call only when m.ui.showOverlay is true.
func (m model) handleOverlayKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.Key()
	isSpace := key.Text == " " || key.Code == ' '
	if isSpace || msg.String() == "esc" {
		m.ui.showOverlay = false
		return m, nil
	}
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	return m, nil
}

// handleNavigationKeys handles box selection and row navigation for expenses/summary.
func (m model) handleNavigationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "r":
		m.resetRowSelection()
		return m, loadData(m.db)
	case "e":
		m.ui.selected = expensesBox
		return m, nil
	case "a":
		m.ui.selected = addBox
		return m, nil
	case "s":
		m.ui.selected = summaryBox
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
	}
	return 1
}
