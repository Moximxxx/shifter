// Package logo provides the Shifter ASCII art logo.
package logo

import "github.com/charmbracelet/lipgloss"

const ascii = `
███████╗██╗  ██╗██╗███████╗████████╗███████╗██████╗
██╔════╝██║  ██║██║██╔════╝╚══██╔══╝██╔════╝██╔══██╗
███████╗███████║██║█████╗     ██║   █████╗  ██████╔╝
╚════██║██╔══██║██║██╔══╝     ██║   ██╔══╝  ██╔══██╗
███████║██║  ██║██║██║        ██║   ███████╗██║  ██║
╚══════╝╚═╝  ╚═╝╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═╝`

var (
	// Purple gradient for the logo
	logoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A78BFA")).
			Bold(true)

	// Tagline below the logo
	taglineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Italic(true).
			PaddingLeft(2)

	// Cached rendered logo
	cached string
)

const taglineEn = "One-click config porting between coding agents"
const taglineZh = "一键在不同 Coding Agent 之间迁移配置"

// Render returns the styled logo with optional tagline.
func Render(tagline string) string {
	if cached != "" && tagline == "" {
		return cached
	}
	if tagline == "" {
		cached = logoStyle.Render(ascii)
		return cached
	}
	result := logoStyle.Render(ascii) + "\n" + taglineStyle.Render(tagline)
	if tagline == "" {
		cached = result
	}
	return result
}

// Small returns a compact one-line brand name.
func Small() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A78BFA")).
		Bold(true).
		Render("🔄 Shifter")
}
