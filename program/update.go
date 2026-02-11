package program

import tea "github.com/charmbracelet/bubbletea"

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

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r": // Refresh data
		m.resetRowSelection()
		return m, loadData(m.db)
	case "e": // Select expenses box
		m.selected = expensesBox
	case "a": // Select add box
		m.selected = addBox
	case "s": // Select summary box
		m.selected = summaryBox
	case "j": // Move selection down
		// Calculate max visible rows based on box height
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowDown(maxRows)
	case "k": // Move selection up
		m.moveRowUp()
	case "g":
		m.moveRowToTop()
	case "G":
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowToBottom(maxRows)
	}

	return m, nil
}

// calculateMaxVisibleRows returns the max number of visible rows in the currently selected box
func (m model) calculateMaxVisibleRows() int {
	const titleHeight = 7
	boxHeight := (m.height - titleHeight) / 2

	switch m.selected {
	case expensesBox:
		// Expenses box: boxHeight - 2 (borders) - 2 (header + separator)
		return boxHeight - 4
	case addBox:
		// Add box doesn't have rows to navigate
		return 1
	case summaryBox:
		// Summary box: boxHeight - 2 (borders) - 2 (header + separator) - 2 (separator + total)
		return boxHeight - 6
	}
	return 1
}
