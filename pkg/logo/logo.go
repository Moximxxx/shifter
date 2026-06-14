// Package logo provides the Shifter ASCII art logo.
// All output is explicitly wrapped with ANSI reset to prevent color leaking
// from surrounding terminal context (e.g., colored install scripts).
package logo

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

const ascii = `
███████╗██╗  ██╗██╗███████╗████████╗███████╗██████╗
██╔════╝██║  ██║██║██╔════╝╚══██╔══╝██╔════╝██╔══██╗
███████╗███████║██║█████╗     ██║   █████╗  ██████╔╝
╚════██║██╔══██║██║██╔══╝     ██║   ██╔══╝  ██╔══██╗
███████║██║  ██║██║██║        ██║   ███████╗██║  ██║
╚══════╝╚═╝  ╚═╝╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═╝`

const logoColor = "#7C3AED" // Shifter purple — matches TUI Primary

var (
	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(logoColor)).
			Bold(true)

	taglineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(logoColor)).
			Italic(true).
			PaddingLeft(2)
)

// Render returns the logo with optional tagline.
// Always resets terminal colors first to prevent leaking from context.
func Render(tagline string) string {
	reset := "\033[0m"
	logo := logoStyle.Render(ascii)
	if tagline == "" {
		return reset + logo + reset
	}
	return reset + logo + "\n" + taglineStyle.Render(tagline) + reset
}

// Version returns the logo with version string in purple.
func Version(version string) string {
	reset := "\033[0m"
	logo := logoStyle.Render(ascii)
	verStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(logoColor)).
		Bold(true).
		Render(version)
	return reset + logo + "\n" + verStyle + reset
}

// Small returns a compact one-line brand name in purple.
func Small() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(logoColor)).
		Bold(true).
		Render("\033[0m🔄 Shifter")
}

// RenderString is like Render but uses fmt.Sprintf for the tagline.
func RenderString(format string, args ...interface{}) string {
	return Render(fmt.Sprintf(format, args...))
}
