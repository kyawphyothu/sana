package program

import (
	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case dataLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.expenses = msg.Expenses
		m.summary = msg.Summary
		m.total = msg.Total
		return m, nil

	case expenseCreatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.err = nil
		m.addFormReset()
		m.selected = expensesBox
		return m, loadData(m.db)

	case formValidationErrMsg:
		m.err = msg.Err
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Get the key from KeyMsg
	key := msg.Key()
	// Check for space key (check both Text and Code)
	isSpace := key.Text == " " || key.Code == ' '

	// Box switching always available
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	// When add box is selected, handle form navigation and forward keys to focused input
	if m.selected == addBox {
		switch msg.String() {
		case "tab":
			// If we just completed a suggestion, move to next field
			if m.typeFieldCompleted {
				m.addFormFocusNext() // This will reset typeFieldCompleted
				return m, nil
			}
			// If Type field is focused and has matched suggestions, accept the suggestion
			if m.hasMatchedSuggestions() {
				in := m.addFormInput()
				var cmd tea.Cmd
				*in, cmd = in.Update(msg)
				// Check if the value is now a complete match (suggestion was accepted)
				if m.isValueCompleteSuggestion() {
					m.typeFieldCompleted = true
				}
				return m, cmd
			}
			// Otherwise, move to next field
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
			m.selected = expensesBox
			return m, nil
		}

		// Forward to focused form input
		in := m.addFormInput()
		var cmd tea.Cmd
		*in, cmd = in.Update(msg)
		// Reset completion flag if user is typing (not Tab or Enter)
		if msg.String() != "tab" && msg.String() != "enter" && m.addFormFocused == addFormType {
			m.typeFieldCompleted = false
		}
		return m, cmd
	}

	// If overlay is open, only handle overlay-specific keys
	if m.showOverlay {
		if isSpace || msg.String() == "esc" {
			m.showOverlay = false
			return m, nil
		}
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Ignore other keys when overlay is open
		return m, nil
	}

	// Handle space key for overlay toggle
	if isSpace {
		if m.selected == summaryBox && len(m.summary) > 0 {
			// Ensure we have a valid selected row
			if m.summarySelectedRow >= 0 && m.summarySelectedRow < len(m.summary) {
				m.showOverlay = !m.showOverlay
			}
		}
		return m, nil
	}

	// Expenses and summary box: row navigation
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "r": // Refresh data
		m.resetRowSelection()
		return m, loadData(m.db)
	case "e": // Select expenses box
		m.selected = expensesBox
		return m, nil
	case "a": // Select add box
		m.selected = addBox
		return m, nil
	case "s": // Select summary box
		m.selected = summaryBox
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
	boxHeight := (m.height - titleHeightForRows) / boxHeightDivisor

	switch m.selected {
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
