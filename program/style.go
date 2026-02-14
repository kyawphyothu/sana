package program

import (
	"fmt"
	"image/color"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
)

// Theme colors - centralized color palette
type Theme struct {
	Primary    color.Color
	Background color.Color
	Foreground color.Color
	Muted      color.Color
	Success    color.Color
	Error      color.Color
	Selected   color.Color
	Border     color.Color
}

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
		Border:     lipgloss.Color("#454666"), // Border color
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
	return s.Base.Width(width).Height(height).Padding(0, borderPadding)
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

// BorderOptions configures border drawing. Height 0 means content-based height.
type BorderOptions struct {
	Width       int
	Height      int
	Title       string
	Bold        bool
	BorderChars BorderChars
	Color       color.Color
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

// DrawBorder draws a border around content using the given options.
// When opts.Height is 0, height is determined by content; otherwise the box has fixed height.
func (s Styles) DrawBorder(content string, opts BorderOptions) string {
	lines := splitLines(content)

	// Default color to theme border if not set
	borderColor := opts.Color
	if borderColor == nil {
		borderColor = s.Theme.Border
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor).Background(s.Theme.Background)

	// Calculate inner width (excluding border characters and padding)
	innerWidth := opts.Width - innerWidthPadding
	if innerWidth < 1 {
		innerWidth = 1
	}

	innerHeight := 0
	if opts.Height > 0 {
		innerHeight = opts.Height - innerHeightPadding
		if innerHeight < 1 {
			innerHeight = 1
		}
	} else {
		innerHeight = len(lines)
		if innerHeight < 1 {
			innerHeight = 1
		}
	}

	var result strings.Builder

	// Top border with optional title
	topBorder := s.buildTopBorder(opts.Width, opts.BorderChars, opts.Title, borderStyle, opts.Bold)
	result.WriteString(topBorder + "\n")

	// Create padding style with background
	paddingStyle := lipgloss.NewStyle().Background(s.Theme.Background)

	for i := 0; i < innerHeight; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		}

		// Pad line to fit width (don't re-style, content is already styled)
		paddedLine := padToWidth(line, innerWidth, s.Theme.Background)

		borderedLine := borderStyle.Render(opts.BorderChars.Vertical) +
			paddingStyle.Render(" ") +
			paddedLine +
			paddingStyle.Render(" ") +
			borderStyle.Render(opts.BorderChars.Vertical)
		result.WriteString(borderedLine + "\n")
	}

	// Bottom border
	bottomBorder := opts.BorderChars.BottomLeft + strings.Repeat(opts.BorderChars.Horizontal, opts.Width-borderCornerCharsWidth) + opts.BorderChars.BottomRight
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
	remainingWidth := width - borderCornerCharsWidth - titleWidth
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
func padToWidth(s string, width int, bgColor color.Color) string {
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

// CategoryColor returns the color for a given expense type category
func CategoryColor(category string) color.Color {
	switch category {
	case "Food":
		return lipgloss.Color("#6BCB77") // soft green
	case "Transport":
		return lipgloss.Color("#4D96FF") // soft blue
	case "Bills":
		return lipgloss.Color("#FF6B6B") // soft red
	case "Shopping":
		return lipgloss.Color("#FFD93D") // warm yellow
	case "Health":
		return lipgloss.Color("#A66CFF") // violet
	case "Other":
		return lipgloss.Color("#9AA0B5") // gray-purple
	default:
		return lipgloss.Color("#9AA0B5") // default to gray-purple
	}
}

// colorToHex converts color.Color to hex string for parsing
func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}

// InvertColor inverts a hex color by converting RGB to inverted RGB
func InvertColor(c color.Color) color.Color {
	hex := colorToHex(c)
	if len(hex) < 7 {
		return c
	}
	hex = hex[1:] // strip #

	var r, g, b int
	if _, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); err != nil || len(hex) != 6 {
		return c
	}

	r = 255 - r
	g = 255 - g
	b = 255 - b
	return lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
}
