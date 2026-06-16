package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/moximxxx/shifter/engine/flowhub"
	"github.com/moximxxx/shifter/pkg/logo"
	"github.com/moximxxx/shifter/pkg/i18n"
	"github.com/moximxxx/shifter/tui/styles"
)

// FlowHubModel is a standalone TUI for browsing FlowHub.
// Unlike WizardModel, it has no main menu — just search + results.
type FlowHubModel struct {
	width    int
	height   int
	quitting bool

	results  []flowhub.Workflow
	loaded   bool
	query    string
	cursor   int
	inputMode bool
}

// NewFlowHubModel creates a standalone FlowHub browser.
func NewFlowHubModel() FlowHubModel {
	return FlowHubModel{}
}

func (m FlowHubModel) Init() tea.Cmd {
	return fetchFlowHubCmd
}

func (m FlowHubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case flowhubResultsMsg:
		m.results = msg.results
		m.loaded = true

	case tea.KeyMsg:
		if m.inputMode {
			switch msg.String() {
			case "enter":
				m.inputMode = false
				return m, nil
			case "esc", "ctrl+c":
				m.inputMode = false
				return m, nil
			case "backspace":
				if len(m.query) > 0 {
					m.query = m.query[:len(m.query)-1]
				}
			default:
				if len(msg.String()) == 1 {
					m.query += msg.String()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			filtered := getFlowHubFiltered(m.results, m.query)
			if m.cursor < len(filtered)-1 {
				m.cursor++
			}
		case "enter":
			if !m.loaded {
				return m, fetchFlowHubCmd
			}
			// Toggle search mode
			m.inputMode = !m.inputMode
			return m, nil
		case "/":
			// Quick search shortcut
			m.inputMode = true
			return m, nil
		}
	}

	return m, nil
}

func (m FlowHubModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(logo.Render(""))
	b.WriteString("\n")
	b.WriteString(styles.MutedText.Render("FlowHub"))
	b.WriteString("\n\n")
	b.WriteString(styles.Title.Render("🌐 " + i18n.T("flowhub.title")))
	b.WriteString("\n\n")

	if !m.loaded {
		b.WriteString(i18n.T("flowhub.loading") + "\n")
		b.WriteString("\n" + styles.HelpBar.Render(i18n.T("help.quit")))
		return b.String()
	}

	b.WriteString("🔍 " + i18n.T("flowhub.search") + ": ")
	if m.inputMode {
		b.WriteString(styles.ActiveItem.Render(m.query + "_"))
	} else {
		b.WriteString(styles.MutedText.Render(m.query))
		b.WriteString(styles.MutedText.Render("  / " + i18n.T("flowhub.search")))
	}
	b.WriteString("\n\n")

	filtered := getFlowHubFiltered(m.results, m.query)
	if len(filtered) == 0 {
		b.WriteString(styles.MutedText.Render(i18n.T("flowhub.no_results")))
		b.WriteString("\n")
	} else {
		for i, w := range filtered {
			line := fmt.Sprintf("%s v%s  ⭐%d", w.Name, w.Version, w.Downloads)
			if i == m.cursor {
				b.WriteString(styles.ActiveItem.Render("❯ " + line))
			} else {
				b.WriteString(styles.InactiveItem.Render("  " + line))
			}
			b.WriteString("\n      " + styles.MutedText.Render(w.Description))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.navigate")+"  / "+i18n.T("help.apply")+"  Esc "+i18n.T("help.quit")))
	return b.String()
}

// getFlowHubFiltered filters workflows by query.
func getFlowHubFiltered(results []flowhub.Workflow, query string) []flowhub.Workflow {
	q := strings.ToLower(query)
	var filtered []flowhub.Workflow
	for _, w := range results {
		if q == "" || strings.Contains(strings.ToLower(w.Name), q) ||
			strings.Contains(strings.ToLower(w.Description), q) {
			filtered = append(filtered, w)
		}
	}
	return filtered
}

// StandaloneFlowHub launches a standalone FlowHub TUI.
func StandaloneFlowHub() error {
	m := NewFlowHubModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
