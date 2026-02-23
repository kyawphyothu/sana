package program

import "testing"

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	if theme.Primary == nil {
		t.Error("DefaultTheme().Primary should be set")
	}
	if theme.Background == nil {
		t.Error("DefaultTheme().Background should be set")
	}
	if theme.Foreground == nil {
		t.Error("DefaultTheme().Foreground should be set")
	}
	if theme.Muted == nil {
		t.Error("DefaultTheme().Muted should be set")
	}
	if theme.Success == nil {
		t.Error("DefaultTheme().Success should be set")
	}
	if theme.Error == nil {
		t.Error("DefaultTheme().Error should be set")
	}
	if theme.Selected == nil {
		t.Error("DefaultTheme().Selected should be set")
	}
	if theme.Border == nil {
		t.Error("DefaultTheme().Border should be set")
	}
}

func TestNewStyles(t *testing.T) {
	theme := DefaultTheme()
	styles := NewStyles(theme)
	if styles.Theme.Primary != theme.Primary {
		t.Error("NewStyles theme should match input theme")
	}
	// Ensure styles can render without panicking
	if styles.Title.Render("test") == "" {
		t.Error("Title style should render non-empty string")
	}
	if styles.Muted.Render("muted") == "" {
		t.Error("Muted style should render non-empty string")
	}
}
