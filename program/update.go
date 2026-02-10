package program

import tea "github.com/charmbracelet/bubbletea"

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case expensesLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.expenses = msg.Expenses
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
	case "r": // Refresh expenses
		m.resetRowSelection()
		return m, loadExpenses(m.db)
	case "1": // Select expenses box
		m.selected = expensesBox
	case "2": // Select summary box
		m.selected = summaryBox
	case "j": // Move selection down
		// Calculate max visible rows based on box height
		maxRows := m.calculateMaxVisibleRows()
		m.moveRowDown(maxRows)
	case "k": // Move selection up
		m.moveRowUp()
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
	case summaryBox:
		// Summary box: boxHeight - 2 (borders) - 2 (header + separator) - 2 (separator + total)
		return boxHeight - 6
	}
	return 1
}
