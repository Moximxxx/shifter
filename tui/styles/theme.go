// Package styles provides the Lipgloss theme for the Shifter TUI.
package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	Primary   = lipgloss.Color("#7C3AED") // Purple
	Success   = lipgloss.Color("#10B981") // Green
	Warning   = lipgloss.Color("#F59E0B") // Amber
	Error     = lipgloss.Color("#EF4444") // Red
	Muted     = lipgloss.Color("#6B7280") // Gray
	Highlight = lipgloss.Color("#F9FAFB") // Near white

	// Title style
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		Padding(0, 1)

	// Section header
	Section = lipgloss.NewStyle().
		Bold(true).
		Foreground(Highlight).
		PaddingLeft(1)

	// Success checkmark
	Check = lipgloss.NewStyle().
		Foreground(Success).
		SetString("✓")

	// Error cross
	Cross = lipgloss.NewStyle().
		Foreground(Error).
		SetString("✗")

	// Warning icon
	WarnIcon = lipgloss.NewStyle().
		Foreground(Warning).
		SetString("⚠")

	// Info text
	MutedText = lipgloss.NewStyle().
		Foreground(Muted)

	// Active item — cursor indicator
	ActiveItem = lipgloss.NewStyle().
		Foreground(Primary).
		Bold(true)

	// Inactive item — same padding as active to prevent cursor jump
	InactiveItem = lipgloss.NewStyle()

	// Border
	Border = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Padding(1)

	// Button
	Button = lipgloss.NewStyle().
		Foreground(Highlight).
		Background(Primary).
		Padding(0, 3).
		Margin(1, 0)

	// Loss warning
	LossWarn = lipgloss.NewStyle().
		Foreground(Warning).
		Italic(true)

	// Help bar
	HelpBar = lipgloss.NewStyle().
		Foreground(Muted).
		Padding(0, 1)
)
