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
	statsBox selectedBox = iota
	expensesBox
)

type model struct {
	db       *sql.DB
	expenses []types.Expense
	styles   Styles

	width  int
	height int

	selected selectedBox

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
		db:       db,
		expenses: []types.Expense{},
		styles:   styles,
		selected: statsBox, // Start with stats box selected
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
	if m.selected > statsBox {
		m.selected--
	} else {
		m.selected = expensesBox
	}
}

func (m *model) moveRight() {
	if m.selected < expensesBox {
		m.selected++
	} else {
		m.selected = statsBox
	}
}

func (m model) isSelected(box selectedBox) bool {
	return m.selected == box
}
