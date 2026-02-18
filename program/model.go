package program

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/kyawphyothu/sana/database"
	"github.com/kyawphyothu/sana/types"
)

// selectedBox represents which box is currently selected
type selectedBox int

const (
	expensesBox selectedBox = iota
	addBox
	summaryBox
	monthlyReportBox
)

// overlayKind identifies which overlay is currently visible.
type overlayKind int

const (
	overlayNone           overlayKind = iota
	overlayCategoryDetail             // category expense breakdown (from summary box)
	overlayConfirmDelete              // confirm expense deletion (from expenses box)
)

// addFormFocus is the index of the focused field in the add-expense form.
const (
	addFormType addFormFocus = iota
	addFormAmount
	addFormDescription
	addFormDate
	addFormNumFields
)

type addFormFocus int

// expenseData holds loaded expense data from the database.
type expenseData struct {
	expenses      []types.Expense
	summary       []types.CategorySummary
	monthlyReport []types.MonthlyReport
	total         float64
}

// uiState holds viewport and UI interaction state.
type uiState struct {
	width  int
	height int

	selected selectedBox

	// Row selection and scrolling (one list per scrollable table)
	expensesList      scrollableList
	summaryList       scrollableList
	monthlyReportList scrollableList

	activeMonth time.Time

	overlay overlayKind
	err     error
}

// addExpenseForm holds the add-expense form inputs and focus state.
type addExpenseForm struct {
	description   textinput.Model
	amount        textinput.Model
	date          textinput.Model
	typeField     textinput.Model
	focused       addFormFocus
	typeCompleted bool // Track if Type field suggestion was just completed
}

type model struct {
	db     *sql.DB
	data   expenseData
	ui     uiState
	form   addExpenseForm
	styles Styles
}

// monthDataLoadedMsg is sent when month-specific data loading finishes.
type monthDataLoadedMsg struct {
	Expenses []types.Expense
	Summary  []types.CategorySummary
	Total    float64
	Err      error
}

// monthlyReportLoadedMsg is sent when the monthly report (all months) loading finishes.
type monthlyReportLoadedMsg struct {
	MonthlyReport []types.MonthlyReport
	Err           error
}

// expenseCreatedMsg is sent when an expense is created (success or error).
type expenseCreatedMsg struct {
	Err error
}

// expenseDeletedMsg is sent when an expense is deleted (success or error).
type expenseDeletedMsg struct {
	Err error
}

// formValidationErrMsg is sent when add form validation fails (so the model can set err).
type formValidationErrMsg struct {
	Err error
}

func newAddFormInput(placeholder string, width int) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetWidth(width)
	return ti
}

// setTextInputStyles sets the styles for a textinput using the v2 API
func setTextInputStyles(ti *textinput.Model, theme Theme) {
	styles := ti.Styles()
	styles.Focused.Prompt = lipgloss.NewStyle().Foreground(theme.Primary)
	styles.Focused.Text = lipgloss.NewStyle().Foreground(theme.Foreground).Background(theme.Background)
	styles.Focused.Placeholder = lipgloss.NewStyle().Foreground(theme.Muted).Background(theme.Background)
	styles.Blurred.Prompt = lipgloss.NewStyle().Foreground(theme.Muted)
	styles.Blurred.Text = lipgloss.NewStyle().Foreground(theme.Foreground).Background(theme.Background)
	styles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(theme.Muted).Background(theme.Background)
	styles.Cursor.Color = theme.Primary
	ti.SetStyles(styles)
}

func InitialModel(db *sql.DB) model {
	theme := DefaultTheme()
	styles := NewStyles(theme)

	// Form inputs (width set in View when we have m.ui.width)
	desc := newAddFormInput("", formWidth)
	desc.Prompt = "Description: "
	setTextInputStyles(&desc, theme)

	amount := newAddFormInput("", formWidth)
	amount.Prompt = fmt.Sprintf("Amount%s: ", strings.Repeat(".", promptWidth-promptOffsetAmount))
	setTextInputStyles(&amount, theme)

	date := newAddFormInput("YYYY-MM-DD or YYYY-MM-DD HH:MM:SS or today", formWidth)
	date.Prompt = fmt.Sprintf("Date%s: ", strings.Repeat(".", promptWidth-promptOffsetDate))
	setTextInputStyles(&date, theme)
	date.SetValue(time.Now().Format("2006-01-02"))

	typ := newAddFormInput("", formWidth)
	typ.Prompt = fmt.Sprintf("Type%s: ", strings.Repeat(".", promptWidth-promptOffsetType))
	setTextInputStyles(&typ, theme)
	typ.ShowSuggestions = true
	typ.SetSuggestions(types.ExpenseTypeSuggestions())

	typ.Focus()

	return model{
		db: db,
		data: expenseData{
			expenses:      []types.Expense{},
			summary:       []types.CategorySummary{},
			monthlyReport: []types.MonthlyReport{},
		},
		ui: uiState{
			selected:    expensesBox,
			activeMonth: time.Now(),
		},
		form: addExpenseForm{
			description: desc,
			amount:      amount,
			date:        date,
			typeField:   typ,
			focused:     addFormType,
		},
		styles: styles,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadMonthData(m.db, time.Time{}), loadMonthlyReportData(m.db))
}

// loadMonthData returns a command that loads expenses, summary, and total for a specific month.
func loadMonthData(db *sql.DB, date time.Time) tea.Cmd {
	if date.IsZero() {
		date = time.Now()
	}
	return func() tea.Msg {
		expenses, err := database.ListExpenses(db, date)
		if err != nil {
			return monthDataLoadedMsg{Err: err}
		}

		summary, err := database.GetExpensesSummary(db, date)
		if err != nil {
			return monthDataLoadedMsg{Err: err}
		}

		total, err := database.GetTotalExpenses(db, date)
		if err != nil {
			return monthDataLoadedMsg{Err: err}
		}

		return monthDataLoadedMsg{
			Expenses: expenses,
			Summary:  summary,
			Total:    total,
		}
	}
}

// loadMonthlyReportData returns a command that loads the monthly report (all months).
func loadMonthlyReportData(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		monthlyReport, err := database.GetMonthlyReport(db)
		if err != nil {
			return monthlyReportLoadedMsg{Err: err}
		}
		return monthlyReportLoadedMsg{MonthlyReport: monthlyReport}
	}
}

func (m *model) moveRowUp() {
	switch m.ui.selected {
	case expensesBox:
		m.ui.expensesList.moveUp()
	case summaryBox:
		m.ui.summaryList.moveUp()
	case monthlyReportBox:
		m.ui.monthlyReportList.moveUp()
	}
}

func (m *model) moveRowDown(maxVisibleRows int) {
	switch m.ui.selected {
	case expensesBox:
		m.ui.expensesList.SetLength(len(m.data.expenses))
		m.ui.expensesList.moveDown(maxVisibleRows)
	case summaryBox:
		m.ui.summaryList.SetLength(len(m.data.summary))
		m.ui.summaryList.moveDown(maxVisibleRows)
	case monthlyReportBox:
		m.ui.monthlyReportList.SetLength(len(m.data.monthlyReport))
		m.ui.monthlyReportList.moveDown(maxVisibleRows)
	}
}

func (m *model) resetRowSelection() {
	m.ui.expensesList.reset()
	m.ui.summaryList.reset()
	m.ui.monthlyReportList.reset()
}

// clampSelections ensures selection indices stay within valid bounds after data changes.
func (m *model) clampSelections() {
	if len(m.data.expenses) > 0 {
		if m.ui.expensesList.SelectedRow() >= len(m.data.expenses) {
			m.ui.expensesList.selectedRow = len(m.data.expenses) - 1
		}
	} else {
		m.ui.expensesList.reset()
	}
	if len(m.data.summary) > 0 {
		if m.ui.summaryList.SelectedRow() >= len(m.data.summary) {
			m.ui.summaryList.selectedRow = len(m.data.summary) - 1
		}
	} else {
		m.ui.summaryList.reset()
	}
	if len(m.data.monthlyReport) > 0 {
		if m.ui.monthlyReportList.SelectedRow() >= len(m.data.monthlyReport) {
			m.ui.monthlyReportList.selectedRow = len(m.data.monthlyReport) - 1
		}
	} else {
		m.ui.monthlyReportList.reset()
	}
}

func (m model) isSelected(box selectedBox) bool {
	return m.ui.selected == box
}

func (m *model) moveRowToTop() {
	switch m.ui.selected {
	case expensesBox:
		m.ui.expensesList.moveToTop()
	case summaryBox:
		m.ui.summaryList.moveToTop()
	case monthlyReportBox:
		m.ui.monthlyReportList.moveToTop()
	}
}

func (m *model) moveRowToBottom(maxVisibleRows int) {
	switch m.ui.selected {
	case expensesBox:
		m.ui.expensesList.SetLength(len(m.data.expenses))
		m.ui.expensesList.moveToBottom(maxVisibleRows)
	case summaryBox:
		m.ui.summaryList.SetLength(len(m.data.summary))
		m.ui.summaryList.moveToBottom(maxVisibleRows)
	case monthlyReportBox:
		m.ui.monthlyReportList.SetLength(len(m.data.monthlyReport))
		m.ui.monthlyReportList.moveToBottom(maxVisibleRows)
	}
}

// addFormInput returns the currently focused form input.
func (m *model) addFormInput() *textinput.Model {
	switch m.form.focused {
	case addFormDescription:
		return &m.form.description
	case addFormAmount:
		return &m.form.amount
	case addFormDate:
		return &m.form.date
	case addFormType:
		return &m.form.typeField
	default:
		return &m.form.typeField
	}
}

// addFormFocusNext moves focus to the next form field (wraps to first).
func (m *model) addFormFocusNext() {
	m.addFormInput().Blur()
	m.form.typeCompleted = false // Reset completion flag when moving focus
	m.form.focused = (m.form.focused + 1) % addFormNumFields
	m.addFormInput().Focus()
}

// addFormFocusPrev moves focus to the previous form field (wraps to last).
func (m *model) addFormFocusPrev() {
	m.addFormInput().Blur()
	m.form.typeCompleted = false // Reset completion flag when moving focus
	m.form.focused--
	if m.form.focused < 0 {
		m.form.focused = addFormNumFields - 1
	}
	m.addFormInput().Focus()
}

// hasMatchedSuggestions checks if the Type field has matched suggestions
func (m *model) hasMatchedSuggestions() bool {
	// Only Type field has suggestions enabled
	if m.form.focused != addFormType {
		return false
	}
	// Check if there are matched suggestions
	suggestions := m.form.typeField.MatchedSuggestions()
	return len(suggestions) > 0
}

// isValueCompleteSuggestion checks if the current Type field value is a complete match for a suggestion
func (m *model) isValueCompleteSuggestion() bool {
	if m.form.focused != addFormType {
		return false
	}
	currentValue := strings.ToLower(strings.TrimSpace(m.form.typeField.Value()))
	if currentValue == "" {
		return false
	}
	// Check against all available suggestions (not just matched ones)
	// because after accepting, the value is the full suggestion text
	suggestions := m.form.typeField.AvailableSuggestions()
	for _, suggestion := range suggestions {
		if strings.ToLower(suggestion) == currentValue {
			return true
		}
	}
	return false
}

// addFormSubmit validates inputs, creates expense, and returns a command that sends expenseCreatedMsg or formValidationErrMsg.
func (m *model) addFormSubmit() tea.Cmd {
	desc := strings.TrimSpace(m.form.description.Value())
	amountStr := strings.TrimSpace(m.form.amount.Value())
	dateStr := strings.TrimSpace(m.form.date.Value())
	typeStr := strings.TrimSpace(m.form.typeField.Value())

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		return func() tea.Msg { return formValidationErrMsg{Err: fmt.Errorf("amount must be a positive number")} }
	}
	var date time.Time
	if dateStr == "" || strings.ToLower(dateStr) == "today" {
		date = time.Now()
	} else {
		// Try parsing as "YYYY-MM-DD HH:MM:SS" first
		date, err = time.ParseInLocation("2006-01-02 15:04:05", dateStr, time.Local)
		if err != nil {
			// If that fails, try parsing as "YYYY-MM-DD" and add current time
			date, err = time.ParseInLocation("2006-01-02", dateStr, time.Local)
			if err != nil {
				return func() tea.Msg {
					return formValidationErrMsg{Err: fmt.Errorf("date must be YYYY-MM-DD or YYYY-MM-DD HH:MM:SS or 'today'")}
				}
			}
			// Add current time's hour, minute, and second in local timezone
			now := time.Now()
			date = time.Date(
				date.Year(), date.Month(), date.Day(),
				now.Hour(), now.Minute(), now.Second(), now.Nanosecond(),
				time.Local,
			)
		}
	}
	// Convert to UTC before storing
	date = date.UTC()
	expType, ok := types.ParseExpenseType(typeStr)
	if !ok {
		expType = types.ExpenseTypeOther
	}

	db := m.db
	return func() tea.Msg {
		_, err := database.CreateExpense(db, date, amount, desc, expType)
		return expenseCreatedMsg{Err: err}
	}
}

// addFormReset clears form fields and refocuses description.
func (m *model) addFormReset() {
	m.form.description.SetValue("")
	m.form.amount.SetValue("")
	m.form.date.SetValue(time.Now().Format("2006-01-02"))
	m.form.typeField.SetValue("")
	m.form.typeCompleted = false // Reset completion flag
	m.addFormInput().Blur()
	m.form.focused = addFormType
	m.form.typeField.Focus()
}

// updatePromptStyles updates prompt colors based on focus state
func (m *model) updatePromptStyles() {
	theme := m.styles.Theme

	// Focused prompt uses Primary color (bright)
	focusedStyle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	// Unfocused prompts use Muted color (subtle)
	unfocusedStyle := lipgloss.NewStyle().Foreground(theme.Muted).Bold(false)

	// Update each input's prompt style based on focus
	updateInputPromptStyle(&m.form.typeField, m.form.focused == addFormType, focusedStyle, unfocusedStyle)
	updateInputPromptStyle(&m.form.amount, m.form.focused == addFormAmount, focusedStyle, unfocusedStyle)
	updateInputPromptStyle(&m.form.description, m.form.focused == addFormDescription, focusedStyle, unfocusedStyle)
	updateInputPromptStyle(&m.form.date, m.form.focused == addFormDate, focusedStyle, unfocusedStyle)
}

// updateInputPromptStyle updates the prompt style for a single input
func updateInputPromptStyle(ti *textinput.Model, isFocused bool, focusedStyle, unfocusedStyle lipgloss.Style) {
	styles := ti.Styles()
	// Always update both states so the correct one is used based on focus
	styles.Focused.Prompt = focusedStyle
	styles.Blurred.Prompt = unfocusedStyle
	ti.SetStyles(styles)
}
