package program

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kyawphyothu/sana/database"
	"github.com/kyawphyothu/sana/types"
)

// selectedBox represents which box is currently selected
type selectedBox int

const (
	expensesBox selectedBox = iota
	summaryBox
)

type model struct {
	db       *sql.DB
	expenses []types.Expense
	styles   Styles

	width  int
	height int

	selected selectedBox

	// Row selection and scrolling
	expensesSelectedRow  int
	expensesScrollOffset int
	summarySelectedRow   int
	summaryScrollOffset  int

	err error
}

// expensesLoadedMsg is sent when ListExpenses finishes (in Init).
type expensesLoadedMsg struct {
	Expenses []types.Expense
	Err      error
}

func InitialModel(db *sql.DB) model {
	theme := DefaultTheme()
	styles := NewStyles(theme)

	return model{
		db:                   db,
		expenses:             []types.Expense{},
		styles:               styles,
		selected:             expensesBox, // Start with expenses box selected
		expensesSelectedRow:  0,
		expensesScrollOffset: 0,
		summarySelectedRow:   0,
		summaryScrollOffset:  0,
	}
}

func (m model) Init() tea.Cmd {
	return loadExpenses(m.db)
}

// loadExpenses returns a command that loads expenses from the database
func loadExpenses(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		expenses, err := database.ListExpenses(db)
		return expensesLoadedMsg{
			Expenses: expenses,
			Err:      err,
		}
	}
}

func (m *model) moveLeft() {
	if m.selected > expensesBox {
		m.selected--
	} else {
		m.selected = summaryBox
	}
}

func (m *model) moveRight() {
	if m.selected < summaryBox {
		m.selected++
	} else {
		m.selected = expensesBox
	}
}

func (m *model) moveRowUp() {
	switch m.selected {
	case expensesBox:
		if m.expensesSelectedRow > 0 {
			m.expensesSelectedRow--
			// Scroll up if needed
			if m.expensesSelectedRow < m.expensesScrollOffset {
				m.expensesScrollOffset = m.expensesSelectedRow
			}
		}
	case summaryBox:
		if m.summarySelectedRow > 0 {
			m.summarySelectedRow--
			// Scroll up if needed
			if m.summarySelectedRow < m.summaryScrollOffset {
				m.summaryScrollOffset = m.summarySelectedRow
			}
		}
	}
}

func (m *model) moveRowDown(maxVisibleRows int) {
	switch m.selected {
	case expensesBox:
		maxRow := len(m.expenses) - 1
		if m.expensesSelectedRow < maxRow {
			m.expensesSelectedRow++
			// Scroll down if needed
			if m.expensesSelectedRow >= m.expensesScrollOffset+maxVisibleRows {
				m.expensesScrollOffset = m.expensesSelectedRow - maxVisibleRows + 1
			}
			// Ensure we don't scroll past the end
			maxScrollOffset := len(m.expenses) - maxVisibleRows
			if maxScrollOffset < 0 {
				maxScrollOffset = 0
			}
			if m.expensesScrollOffset > maxScrollOffset {
				m.expensesScrollOffset = maxScrollOffset
			}
		}
	case summaryBox:
		// Count categories
		categoryCount := 0
		categoryTotals := make(map[string]float64)
		for _, expense := range m.expenses {
			categoryTotals[expense.Type.String()] += expense.Amount
		}
		categoryCount = len(categoryTotals)

		maxRow := categoryCount - 1
		if m.summarySelectedRow < maxRow {
			m.summarySelectedRow++
			// Scroll down if needed
			if m.summarySelectedRow >= m.summaryScrollOffset+maxVisibleRows {
				m.summaryScrollOffset = m.summarySelectedRow - maxVisibleRows + 1
			}
			// Ensure we don't scroll past the end
			maxScrollOffset := categoryCount - maxVisibleRows
			if maxScrollOffset < 0 {
				maxScrollOffset = 0
			}
			if m.summaryScrollOffset > maxScrollOffset {
				m.summaryScrollOffset = maxScrollOffset
			}
		}
	}
}

func (m *model) resetRowSelection() {
	m.expensesSelectedRow = 0
	m.expensesScrollOffset = 0
	m.summarySelectedRow = 0
	m.summaryScrollOffset = 0
}

func (m model) isSelected(box selectedBox) bool {
	return m.selected == box
}
