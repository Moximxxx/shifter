package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/engine/detect"
	"github.com/moximxxx/shifter/engine/flowhub"
	"github.com/moximxxx/shifter/pkg/i18n"
	"github.com/moximxxx/shifter/pkg/logo"
	"github.com/moximxxx/shifter/registry"
	"github.com/moximxxx/shifter/tui/styles"
)

const (
	fhList = iota
	fhDetail
	fhInstalled
)

// FlowHubModel is a standalone TUI for browsing FlowHub.
type FlowHubModel struct {
	width     int
	height    int
	quitting  bool
	screen    int // fhList, fhDetail, fhInstalled

	results    []flowhub.Workflow
	loaded     bool
	query      string
	cursor     int
	inputMode  bool
	selected   *flowhub.Workflow
	installMsg string
}

func NewFlowHubModel() FlowHubModel { return FlowHubModel{} }

func (m FlowHubModel) Init() tea.Cmd { return fetchFlowHubCmd }

func (m FlowHubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case flowhubResultsMsg:
		m.results = msg.results
		m.loaded = true

	case tea.KeyMsg:
		switch m.screen {
		case fhDetail, fhInstalled:
			if msg.String() == "esc" || msg.String() == "q" || msg.String() == "ctrl+c" {
				m.screen = fhList
				return m, nil
			}
			if m.screen == fhDetail && msg.String() == "enter" {
				return m, m.installSelected()
			}
			return m, nil
		}

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
			// Open detail for selected workflow
			filtered := getFlowHubFiltered(m.results, m.query)
			if m.cursor < len(filtered) {
				w := filtered[m.cursor]
				m.selected = &w
				m.screen = fhDetail
			}
			return m, nil
		case "/":
			m.inputMode = true
			return m, nil
		}
	}

	return m, nil
}

func (m FlowHubModel) installSelected() tea.Cmd {
	return func() tea.Msg {
		w := m.selected
		if w == nil {
			return nil
		}
		data, err := flowhub.Download(w.Name)
		if err != nil {
			return fmt.Errorf("download: %w", err)
		}
		var cfg canonical.ShifterConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse: %w", err)
		}
		// Try to auto-detect target agent
		results := detect.ScanAll()
		target := ""
		for _, r := range results {
			if r.Found && r.HasProjectConfig {
				target = r.ID
				break
			}
		}
		if target == "" {
			return fmt.Errorf("no target agent found")
		}
		a, err := registry.Get(target)
		if err != nil {
			return err
		}
		ctx := context.Background()
		result, err := a.Write(ctx, &cfg, adapter.WriteOptions{
			Scope: "project", ProjectRoot: ".", Backup: true,
		})
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
		return fmt.Sprintf("✓ Installed %s → %s (%d files)", w.Name, target, len(result.FilesWritten))
	}
}

func (m FlowHubModel) View() string {
	if m.quitting {
		return ""
	}

	switch m.screen {
	case fhDetail:
		return m.viewDetail()
	case fhInstalled:
		return m.viewInstalled()
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
	} else {
		for i, w := range filtered {
			line := fmt.Sprintf("%s v%s  ⭐%d", w.Name, w.Version, w.Downloads)
			if i == m.cursor {
				b.WriteString(styles.ActiveItem.Render("❯ " + line))
			} else {
				b.WriteString(styles.InactiveItem.Render("  " + line))
			}
			b.WriteString("\n")
			// Card-style: description + agent + tags below
			var parts []string
			if w.Description != "" {
				parts = append(parts, w.Description)
			}
			if w.Agent != "" {
				parts = append(parts, fmt.Sprintf("%s: %s", i18n.T("detail.source"), w.Agent))
			}
			if len(w.Tags) > 0 {
				parts = append(parts, strings.Join(w.Tags, ", "))
			}
			b.WriteString(styles.MutedText.Render(fmt.Sprintf("      %s", strings.Join(parts, "  |  "))))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.navigate")+"  / "+i18n.T("help.apply")+"  "+i18n.T("help.select")+"  Esc "+i18n.T("help.quit")))
	return b.String()
}

func (m FlowHubModel) viewDetail() string {
	w := m.selected
	if w == nil {
		return "No workflow selected"
	}
	var b strings.Builder
	b.WriteString(logo.Render(""))
	b.WriteString("\n")
	b.WriteString(styles.Title.Render(fmt.Sprintf("📦 %s v%s", w.Name, w.Version)))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Author:     %s\n", w.Author))
	b.WriteString(fmt.Sprintf("Agent:      %s\n", w.Agent))
	b.WriteString(fmt.Sprintf("Downloads:  ⭐%d\n", w.Downloads))
	if len(w.Tags) > 0 {
		b.WriteString(fmt.Sprintf("Tags:       %s\n", strings.Join(w.Tags, ", ")))
	}
	if w.Category != "" {
		b.WriteString(fmt.Sprintf("Category:   %s\n", w.Category))
	}
	if w.Description != "" {
		b.WriteString(fmt.Sprintf("\n%s\n", w.Description))
	}
	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.apply")+"  "+i18n.T("help.back")))
	return b.String()
}

func (m FlowHubModel) viewInstalled() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render(m.installMsg))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.back")))
	return b.String()
}

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

func StandaloneFlowHub() error {
	m := NewFlowHubModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
