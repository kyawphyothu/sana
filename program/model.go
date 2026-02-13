package program

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
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

type model struct {
	db       *sql.DB
	expenses []types.Expense
	summary  []types.CategorySummary // Category summaries from database
	total    float64                 // Grand total from database
	styles   Styles

	width  int
	height int

	selected selectedBox

	// Row selection and scrolling
	expensesSelectedRow  int
	expensesScrollOffset int
	summarySelectedRow   int
	summaryScrollOffset  int

	// Overlay state
	showOverlay bool

	// Add expense form
	addDescription     textinput.Model
	addAmount          textinput.Model
	addDate            textinput.Model
	addType            textinput.Model
	addFormFocused     addFormFocus
	typeFieldCompleted bool // Track if Type field suggestion was just completed

	err error
}

// dataLoadedMsg is sent when data loading finishes (in Init).
type dataLoadedMsg struct {
	Expenses []types.Expense
	Summary  []types.CategorySummary
	Total    float64
	Err      error
}

// expenseCreatedMsg is sent when an expense is created (success or error).
type expenseCreatedMsg struct {
	Err error
}

// formValidationErrMsg is sent when add form validation fails (so the model can set err).
type formValidationErrMsg struct {
	Err error
}

// blinkMsg is sent periodically to trigger cursor blinking in text inputs.
type blinkMsg struct{}

// blinkCmd returns a command that sends a blinkMsg after the blink interval.
func blinkCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return blinkMsg{}
	})
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

	// Form inputs (width set in View when we have m.width)
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
		db:                   db,
		expenses:             []types.Expense{},
		summary:              []types.CategorySummary{},
		styles:               styles,
		selected:             expensesBox,
		expensesSelectedRow:  0,
		expensesScrollOffset: 0,
		summarySelectedRow:   0,
		summaryScrollOffset:  0,
		showOverlay:          false,
		addDescription:       desc,
		addAmount:            amount,
		addDate:              date,
		addType:              typ,
		addFormFocused:       addFormType,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(loadData(m.db), blinkCmd())
}

// loadData returns a command that loads expenses, summary, and total from the database
func loadData(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		expenses, err := database.ListExpenses(db)
		if err != nil {
			return dataLoadedMsg{Err: err}
		}

		summary, err := database.GetExpensesSummary(db)
		if err != nil {
			return dataLoadedMsg{Err: err}
		}

		total, err := database.GetTotalExpenses(db)
		if err != nil {
			return dataLoadedMsg{Err: err}
		}

		return dataLoadedMsg{
			Expenses: expenses,
			Summary:  summary,
			Total:    total,
			Err:      nil,
		}
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
			m.adjustScrollOffset(maxVisibleRows)
		}
	case summaryBox:
		maxRow := len(m.summary) - 1
		if m.summarySelectedRow < maxRow {
			m.summarySelectedRow++
			m.adjustScrollOffset(maxVisibleRows)
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

func (m *model) moveRowToTop() {
	switch m.selected {
	case expensesBox:
		m.expensesSelectedRow = 0
		m.expensesScrollOffset = 0
	case summaryBox:
		m.summarySelectedRow = 0
		m.summaryScrollOffset = 0
	}
}

func (m *model) moveRowToBottom(maxVisibleRows int) {
	switch m.selected {
	case expensesBox:
		m.expensesSelectedRow = len(m.expenses) - 1
		m.adjustScrollOffset(maxVisibleRows)
	case summaryBox:
		m.summarySelectedRow = len(m.summary) - 1
		m.adjustScrollOffset(maxVisibleRows)
	}
}

// addFormInput returns the currently focused form input.
func (m *model) addFormInput() *textinput.Model {
	switch m.addFormFocused {
	case addFormDescription:
		return &m.addDescription
	case addFormAmount:
		return &m.addAmount
	case addFormDate:
		return &m.addDate
	case addFormType:
		return &m.addType
	default:
		return &m.addType
	}
}

// addFormFocusNext moves focus to the next form field (wraps to first).
func (m *model) addFormFocusNext() {
	m.addFormInput().Blur()
	m.typeFieldCompleted = false // Reset completion flag when moving focus
	m.addFormFocused = (m.addFormFocused + 1) % addFormNumFields
	m.addFormInput().Focus()
}

// addFormFocusPrev moves focus to the previous form field (wraps to last).
func (m *model) addFormFocusPrev() {
	m.addFormInput().Blur()
	m.typeFieldCompleted = false // Reset completion flag when moving focus
	m.addFormFocused--
	if m.addFormFocused < 0 {
		m.addFormFocused = addFormNumFields - 1
	}
	m.addFormInput().Focus()
}

// hasMatchedSuggestions checks if the Type field has matched suggestions
func (m *model) hasMatchedSuggestions() bool {
	// Only Type field has suggestions enabled
	if m.addFormFocused != addFormType {
		return false
	}
	// Check if there are matched suggestions
	suggestions := m.addType.MatchedSuggestions()
	return len(suggestions) > 0
}

// isValueCompleteSuggestion checks if the current Type field value is a complete match for a suggestion
func (m *model) isValueCompleteSuggestion() bool {
	if m.addFormFocused != addFormType {
		return false
	}
	currentValue := strings.ToLower(strings.TrimSpace(m.addType.Value()))
	if currentValue == "" {
		return false
	}
	// Check against all available suggestions (not just matched ones)
	// because after accepting, the value is the full suggestion text
	suggestions := m.addType.AvailableSuggestions()
	for _, suggestion := range suggestions {
		if strings.ToLower(suggestion) == currentValue {
			return true
		}
	}
	return false
}

// addFormSubmit validates inputs, creates expense, and returns a command that sends expenseCreatedMsg or formValidationErrMsg.
func (m *model) addFormSubmit() tea.Cmd {
	desc := strings.TrimSpace(m.addDescription.Value())
	amountStr := strings.TrimSpace(m.addAmount.Value())
	dateStr := strings.TrimSpace(m.addDate.Value())
	typeStr := strings.TrimSpace(m.addType.Value())

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
	m.addDescription.SetValue("")
	m.addAmount.SetValue("")
	m.addDate.SetValue(time.Now().Format("2006-01-02"))
	m.addType.SetValue("")
	m.typeFieldCompleted = false // Reset completion flag
	m.addFormInput().Blur()
	m.addFormFocused = addFormType
	m.addType.Focus()
}

// updatePromptStyles updates prompt colors based on focus state
func (m *model) updatePromptStyles() {
	theme := m.styles.Theme

	// Focused prompt uses Primary color (bright)
	focusedStyle := lipgloss.NewStyle().Foreground(theme.Primary).Bold(true)
	// Unfocused prompts use Muted color (subtle)
	unfocusedStyle := lipgloss.NewStyle().Foreground(theme.Muted).Bold(false)

	// Update each input's prompt style based on focus
	updateInputPromptStyle(&m.addType, m.addFormFocused == addFormType, focusedStyle, unfocusedStyle)
	updateInputPromptStyle(&m.addAmount, m.addFormFocused == addFormAmount, focusedStyle, unfocusedStyle)
	updateInputPromptStyle(&m.addDescription, m.addFormFocused == addFormDescription, focusedStyle, unfocusedStyle)
	updateInputPromptStyle(&m.addDate, m.addFormFocused == addFormDate, focusedStyle, unfocusedStyle)
}

// updateInputPromptStyle updates the prompt style for a single input
func updateInputPromptStyle(ti *textinput.Model, isFocused bool, focusedStyle, unfocusedStyle lipgloss.Style) {
	styles := ti.Styles()
	// Always update both states so the correct one is used based on focus
	styles.Focused.Prompt = focusedStyle
	styles.Blurred.Prompt = unfocusedStyle
	ti.SetStyles(styles)
}

// adjustScrollOffset adjusts the scroll offset to the selected row (ensure row visible)
func (m *model) adjustScrollOffset(maxVisibleRows int) {
	switch m.selected {
	case expensesBox:
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
	case summaryBox:
		if m.summarySelectedRow >= m.summaryScrollOffset+maxVisibleRows {
			m.summaryScrollOffset = m.summarySelectedRow - maxVisibleRows + 1
		}
		// Ensure we don't scroll past the end
		maxScrollOffset := len(m.summary) - maxVisibleRows
		if maxScrollOffset < 0 {
			maxScrollOffset = 0
		}
		if m.summaryScrollOffset > maxScrollOffset {
			m.summaryScrollOffset = maxScrollOffset
		}
	}
}
