package program

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme colors - centralized color palette
type Theme struct {
	Primary    lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
	Muted      lipgloss.Color
	Success    lipgloss.Color
	Error      lipgloss.Color
	Selected   lipgloss.Color
}

// DefaultTheme returns Sana's purple theme
// func DefaultTheme() Theme {
// 	return Theme{
// 		Primary:    lipgloss.Color("#9C9ECF"), // Sana's purple
// 		Background: lipgloss.Color("#20212E"), // Cozy dark
// 		Foreground: lipgloss.Color("#E6E7F2"), // Light text
// 		Muted:      lipgloss.Color("#A1A1B8"), // Secondary text
// 		Success:    lipgloss.Color("#A8D5BA"), // Green
// 		Error:      lipgloss.Color("#E8A5A5"), // Red
// 		Selected:   lipgloss.Color("#7DD3FC"), // Bright cyan for selected items
// 	}
// }

// Background        → #0F1117
// Panel BG          → #161925
// Text              → #E6E8F0
// Muted Text        → #7C809F

// Primary Accent    → #9C9ECF
// Focused Border    → #9C9ECF
// Selected Item     → FG #9C9ECF / BG #252A40

// Error             → #E07A7A
// Success           → #7BD88F
// Warning           → #F2C97D
// DefaultTheme returns Sana's purple theme
func DefaultTheme() Theme {
	return Theme{
		Primary:    lipgloss.Color("#9C9ECF"), // Sana's purple - kept as requested
		Background: lipgloss.Color("#0F1117"), // Much darker, richer background
		Foreground: lipgloss.Color("#E6E8F0"), // Brighter, slightly purple-tinted text
		Muted:      lipgloss.Color("#7A7B9A"), // Darker muted text for better contrast
		Success:    lipgloss.Color("#88D4AB"), // Brighter green
		Error:      lipgloss.Color("#FF9B9B"), // Brighter red
		Selected:   lipgloss.Color("#5FC9F8"), // Brighter cyan
	}
}

// Styles contains all UI styles
type Styles struct {
	Theme  Theme
	Border lipgloss.Border

	// Container styles
	Base   lipgloss.Style
	Parent lipgloss.Style

	// Text styles
	Title  lipgloss.Style
	Header lipgloss.Style
	Line   lipgloss.Style
	Muted  lipgloss.Style

	// Interactive styles
	Selected lipgloss.Style
}

// NewStyles creates a new Styles instance with the given theme
func NewStyles(theme Theme) Styles {
	border := lipgloss.RoundedBorder()

	return Styles{
		Theme:  theme,
		Border: border,

		// Base container with border
		Base: lipgloss.NewStyle().
			// Border(border, true, true, true, true).
			// BorderForeground(theme.Primary).
			Background(theme.Background).
			Foreground(theme.Foreground),

		// Parent container (no border, just background)
		Parent: lipgloss.NewStyle().
			Background(theme.Background),

		// Title text (large headers, figlet text)
		Title: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Background(theme.Background).
			Bold(true),

		// Section headers
		Header: lipgloss.NewStyle().
			Foreground(theme.Primary).
			Background(theme.Background).
			Bold(true),

		// Regular text
		Line: lipgloss.NewStyle().
			Foreground(theme.Foreground).
			Background(theme.Background),

		// Muted/secondary text
		Muted: lipgloss.NewStyle().
			Foreground(theme.Muted).
			Background(theme.Background),

		// Selected/highlighted items
		Selected: lipgloss.NewStyle().
			Foreground(theme.Background).
			Background(theme.Primary).
			Bold(true),
	}
}

// Box creates a styled box with the given width and height
func (s Styles) Box(width, height int) lipgloss.Style {
	return s.Base.Width(width).Height(height).Padding(0, 2)
}

// CenteredBox creates a centered box
func (s Styles) CenteredBox(width, height int) lipgloss.Style {
	return s.Box(width, height).
		Align(lipgloss.Center, lipgloss.Center)
}

// Manual border drawing functions

// BorderChars defines box-drawing characters for borders
type BorderChars struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
}

// RoundedBorderChars returns rounded box-drawing characters
func RoundedBorderChars() BorderChars {
	return BorderChars{
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
		Horizontal:  "─",
		Vertical:    "│",
	}
}

// SharpBorderChars returns sharp box-drawing characters
func SharpBorderChars() BorderChars {
	return BorderChars{
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
	}
}

// DoubleBorderChars returns double-line box-drawing characters
func DoubleBorderChars() BorderChars {
	return BorderChars{
		TopLeft:     "╔",
		TopRight:    "╗",
		BottomLeft:  "╚",
		BottomRight: "╝",
		Horizontal:  "═",
		Vertical:    "║",
	}
}

// DrawBorder manually draws a border around content
func (s Styles) DrawBorder(content string, width int, borderChars BorderChars, borderColor lipgloss.Color) string {
	return s.DrawBorderWithTitle(content, width, borderChars, borderColor, "")
}

// DrawBorderWithTitle manually draws a border around content with optional title in top border
func (s Styles) DrawBorderWithTitle(content string, width int, borderChars BorderChars, borderColor lipgloss.Color, title string) string {
	return s.DrawBorderWithTitleBold(content, width, borderChars, borderColor, title, false)
}

// DrawBorderWithTitleBold manually draws a border with optional title and bold option
func (s Styles) DrawBorderWithTitleBold(content string, width int, borderChars BorderChars, borderColor lipgloss.Color, title string, bold bool) string {
	lines := splitLines(content)
	
	borderStyle := lipgloss.NewStyle().Foreground(borderColor).Background(s.Theme.Background)
	
	// Calculate inner width (excluding border characters and padding)
	innerWidth := width - 4 // 2 for borders + 2 for padding
	if innerWidth < 1 {
		innerWidth = 1
	}
	
	var result strings.Builder
	
	// Top border with optional title
	topBorder := s.buildTopBorder(width, borderChars, title, borderStyle, bold)
	result.WriteString(topBorder + "\n")
	
	// Create padding style with background
	paddingStyle := lipgloss.NewStyle().Background(s.Theme.Background)
	
	// Content lines with side borders
	for _, line := range lines {
		// Pad line to fit width (don't re-style, content is already styled)
		paddedLine := padToWidth(line, innerWidth, s.Theme.Background)
		
		borderedLine := borderStyle.Render(borderChars.Vertical) + 
			paddingStyle.Render(" ") + 
			paddedLine + 
			paddingStyle.Render(" ") + 
			borderStyle.Render(borderChars.Vertical)
		result.WriteString(borderedLine + "\n")
	}
	
	// Bottom border
	bottomBorder := borderChars.BottomLeft + strings.Repeat(borderChars.Horizontal, width-2) + borderChars.BottomRight
	result.WriteString(borderStyle.Render(bottomBorder))
	
	return result.String()
}

// DrawBorderWithHeight draws a border with specific height, padding content vertically
func (s Styles) DrawBorderWithHeight(content string, width, height int, borderChars BorderChars, borderColor lipgloss.Color) string {
	return s.DrawBorderWithHeightAndTitle(content, width, height, borderChars, borderColor, "")
}

// DrawBorderWithHeightAndTitle draws a border with specific height and optional title in top border
func (s Styles) DrawBorderWithHeightAndTitle(content string, width, height int, borderChars BorderChars, borderColor lipgloss.Color, title string) string {
	return s.DrawBorderWithHeightAndTitleBold(content, width, height, borderChars, borderColor, title, false)
}

// DrawBorderWithHeightAndTitleBold draws a border with specific height, optional title, and bold option
func (s Styles) DrawBorderWithHeightAndTitleBold(content string, width, height int, borderChars BorderChars, borderColor lipgloss.Color, title string, bold bool) string {
	lines := splitLines(content)

	borderStyle := lipgloss.NewStyle().Foreground(borderColor).Background(s.Theme.Background)

	// Calculate inner dimensions
	innerWidth := width - 4   // 2 for borders + 2 for padding
	innerHeight := height - 2 // 2 for top and bottom borders
	if innerWidth < 1 {
		innerWidth = 1
	}
	if innerHeight < 1 {
		innerHeight = 1
	}

	var result strings.Builder

	// Top border with optional title
	topBorder := s.buildTopBorder(width, borderChars, title, borderStyle, bold)
	result.WriteString(topBorder + "\n")

	// Create padding style with background
	paddingStyle := lipgloss.NewStyle().Background(s.Theme.Background)
	
	// Pad lines to fill height
	for i := 0; i < innerHeight; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		} else {
			line = ""
		}
		
		// Pad line to fit width (don't re-style, content is already styled)
		paddedLine := padToWidth(line, innerWidth, s.Theme.Background)
		
		borderedLine := borderStyle.Render(borderChars.Vertical) + 
			paddingStyle.Render(" ") + 
			paddedLine + 
			paddingStyle.Render(" ") + 
			borderStyle.Render(borderChars.Vertical)
		result.WriteString(borderedLine + "\n")
	}

	// Bottom border
	bottomBorder := borderChars.BottomLeft + strings.Repeat(borderChars.Horizontal, width-2) + borderChars.BottomRight
	result.WriteString(borderStyle.Render(bottomBorder))

	return result.String()
}

// Helper functions

// buildTopBorder creates the top border line with optional title and bold option
func (s Styles) buildTopBorder(width int, borderChars BorderChars, title string, borderStyle lipgloss.Style, bold bool) string {
	if title == "" {
		// No title, just plain border
		return borderStyle.Render(borderChars.TopLeft + strings.Repeat(borderChars.Horizontal, width-2) + borderChars.TopRight)
	}
	
	// Style the title text (always apply foreground and background)
	titleStyle := lipgloss.NewStyle().Foreground(borderStyle.GetForeground()).Background(s.Theme.Background)
	if bold {
		titleStyle = titleStyle.Bold(true)
	}
	titleText := titleStyle.Render(title)
	
	// Add spacing around title: "─ Title ─"
	titleWithSpacing := " " + titleText + " "
	titleWidth := lipgloss.Width(titleWithSpacing)
	
	// Calculate remaining horizontal line space
	remainingWidth := width - 2 - titleWidth // -2 for corner chars
	if remainingWidth < 0 {
		remainingWidth = 0
	}
	
	// Split remaining space (more on the right)
	leftWidth := 1
	rightWidth := remainingWidth - leftWidth
	if rightWidth < 0 {
		rightWidth = 0
	}
	
	// Build: ╭─ Title ─────────╮
	// Corner and lines styled with borderStyle, title styled separately if bold
	leftPart := borderStyle.Render(borderChars.TopLeft + strings.Repeat(borderChars.Horizontal, leftWidth))
	rightPart := borderStyle.Render(strings.Repeat(borderChars.Horizontal, rightWidth) + borderChars.TopRight)
	
	// Space padding around title with background
	spacePadding := lipgloss.NewStyle().Background(s.Theme.Background).Render(" ")
	
	return leftPart + spacePadding + titleText + spacePadding + rightPart
}

// splitLines splits content by newlines
func splitLines(content string) []string {
	return strings.Split(strings.TrimRight(content, "\n"), "\n")
}

// padToWidth pads a string (that may contain ANSI codes) to the specified width
func padToWidth(s string, width int, bgColor lipgloss.Color) string {
	// Use lipgloss to measure the actual rendered width
	currentWidth := lipgloss.Width(s)
	
	if currentWidth >= width {
		// Truncate if needed using lipgloss's truncate
		return lipgloss.NewStyle().Width(width).MaxWidth(width).Inline(true).Render(s)
	}
	
	// Create a style with background for the entire width
	// This ensures the background fills the entire line
	lineStyle := lipgloss.NewStyle().
		Background(bgColor).
		Width(width).
		Inline(true)
	
	return lineStyle.Render(s)
}
