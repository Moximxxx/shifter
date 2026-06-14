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

)

// Render returns the logo with optional tagline.
// Logo art is purple; tagline is dimmed for visual hierarchy.
// Always resets terminal colors first.
func Render(tagline string) string {
	reset := "\033[0m"
	logo := logoStyle.Render(ascii)
	if tagline == "" {
		return reset + logo + reset
	}
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render
	return reset + logo + "\n  " + dimStyle(tagline) + reset
}

// Version returns the logo with version string.
// Logo is purple; version is dimmed.
func Version(version string) string {
	reset := "\033[0m"
	logo := logoStyle.Render(ascii)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render
	return reset + logo + "\n" + dimStyle(version) + reset
}

// Small returns a compact one-line brand name — logo icon purple, text dim.
func Small() string {
	icon := lipgloss.NewStyle().Foreground(lipgloss.Color(logoColor)).Bold(true).Render("🔄")
	text := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Bold(true).Render("Shifter")
	return "\033[0m" + icon + " " + text
}

// RenderString is like Render but uses fmt.Sprintf for the tagline.
func RenderString(format string, args ...interface{}) string {
	return Render(fmt.Sprintf(format, args...))
}
