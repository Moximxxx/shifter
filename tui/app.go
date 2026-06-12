// Package tui provides the Bubble Tea terminal UI for Shifter.
package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/moximxxx/shifter/engine/detect"
	"github.com/moximxxx/shifter/engine/port"
	"github.com/moximxxx/shifter/tui/styles"
)

// Screen represents the current UI screen.
type Screen int

const (
	ScreenDetect Screen = iota
	ScreenSelectSource
	ScreenSelectTarget
	ScreenAspects
	ScreenPreview
	ScreenResult
)

// Model is the top-level Bubble Tea model.
type Model struct {
	screen      Screen
	width       int
	height      int
	spinner     spinner.Model
	loading     bool
	errorMsg    string
	quitting    bool

	// Detection results
	detectResults []detect.Result

	// Selection state
	sourceIdx    int
	targetIdx    int
	sourceID     string
	targetID     string

	// Aspect checkboxes (true = selected)
	aspects      []aspectItem
	cursorIdx    int

	// Port result
	portResult   *port.Result
	portErr      error
}

type aspectItem struct {
	Name        string
	Label       string
	Selected    bool
	Lossy       bool    // true if lossy for target
	LossReason  string
	Count       int     // item count
}

// NewModel creates the initial TUI model.
func NewModel() Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return Model{
		screen:  ScreenDetect,
		spinner: s,
		loading: true,
		sourceIdx: -1,
		targetIdx: -1,
	}
}

// Init is the Bubble Tea initialization command.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		scanAgentsCmd,
	)
}

type scanDoneMsg struct {
	results []detect.Result
}

func scanAgentsCmd() tea.Msg {
	results := detect.ScanAll()
	return scanDoneMsg{results: results}
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.screen == ScreenDetect || m.screen == ScreenResult {
				m.quitting = true
				return m, tea.Quit
			}
			// Go back to previous screen
			m.goBack()
			return m, nil
		case "enter":
			return m.handleEnter()
		case "up", "k":
			if m.cursorIdx > 0 {
				m.cursorIdx--
			}
		case "down", "j":
			m.cursorIdx++
		case " ":
			// Toggle checkbox on aspects screen
			if m.screen == ScreenAspects && m.cursorIdx < len(m.aspects) {
				m.aspects[m.cursorIdx].Selected = !m.aspects[m.cursorIdx].Selected
			}
		}

	case scanDoneMsg:
		m.detectResults = msg.results
		m.loading = false
		m.screen = ScreenSelectSource

		// Auto-select first found agent as source
		for i, r := range m.detectResults {
			if r.Found {
				m.sourceIdx = i
				m.sourceID = r.ID
				m.cursorIdx = i
				break
			}
		}

	case port.Result:
		m.portResult = &msg
		m.loading = false
		m.screen = ScreenResult

	case error:
		m.portErr = msg
		m.loading = false

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m *Model) goBack() {
	switch m.screen {
	case ScreenSelectSource:
		m.screen = ScreenDetect
	case ScreenSelectTarget:
		m.screen = ScreenSelectSource
		m.targetIdx = -1
		m.targetID = ""
	case ScreenAspects:
		m.screen = ScreenSelectTarget
	case ScreenPreview:
		m.screen = ScreenAspects
	}
	m.cursorIdx = 0
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenSelectSource:
		if m.cursorIdx >= 0 && m.cursorIdx < len(m.detectResults) {
			r := m.detectResults[m.cursorIdx]
			if r.Found {
				m.sourceIdx = m.cursorIdx
				m.sourceID = r.ID
				m.screen = ScreenSelectTarget
				m.cursorIdx = 0
				// Default to first other found agent
				for i, r2 := range m.detectResults {
					if r2.Found && r2.ID != m.sourceID {
						m.cursorIdx = i
						break
					}
				}
			}
		}

	case ScreenSelectTarget:
		if m.cursorIdx >= 0 && m.cursorIdx < len(m.detectResults) {
			r := m.detectResults[m.cursorIdx]
			if r.Found && r.ID != m.sourceID {
				m.targetIdx = m.cursorIdx
				m.targetID = r.ID
				m.screen = ScreenAspects
				m.buildAspects()
				m.cursorIdx = 0
				return m, nil
			}
		}

	case ScreenAspects:
		m.screen = ScreenPreview
		return m, m.executePort()

	case ScreenResult:
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) buildAspects() {
	src := m.detectResults[m.sourceIdx]

	m.aspects = []aspectItem{
		{Name: "instructions", Label: "Project Instructions", Selected: true, Count: src.Summary["instructions"]},
		{Name: "agents", Label: "Agents/Subagents", Selected: true, Count: src.Summary["agents"]},
		{Name: "skills", Label: "Skills", Selected: true, Count: src.Summary["skills"],
			Lossy: !supportsAspect(m.targetID, "skills"), LossReason: "Target has limited skill support"},
		{Name: "commands", Label: "Slash Commands", Selected: true, Count: src.Summary["commands"],
			Lossy: !supportsAspect(m.targetID, "commands"), LossReason: "Target has no native slash commands"},
		{Name: "mcp", Label: "MCP Servers", Selected: true, Count: src.Summary["mcp"]},
		{Name: "permissions", Label: "Permissions", Selected: true,
			Lossy: !supportsAspect(m.targetID, "permissions"), LossReason: "Permission models differ"},
		{Name: "hooks", Label: "Hooks", Selected: true,
			Lossy: !supportsAspect(m.targetID, "hooks"), LossReason: "Target has no hook system"},
		{Name: "settings", Label: "Settings", Selected: false},
	}
}

func supportsAspect(agentID, aspect string) bool {
	switch aspect {
	case "skills":
		return agentID == "claude-code" || agentID == "opencode" || agentID == "qoder"
	case "commands":
		return agentID == "claude-code" || agentID == "opencode"
	case "permissions":
		return agentID != "cline"
	case "hooks":
		return agentID == "claude-code" || agentID == "codex"
	}
	return true
}

func (m *Model) executePort() tea.Cmd {
	return func() tea.Msg {
		var aspects []string
		for _, a := range m.aspects {
			if a.Selected {
				aspects = append(aspects, a.Name)
			}
		}

		ctx := context.Background()
		result, err := port.Port(ctx, port.Options{
			Source:      m.sourceID,
			Target:      m.targetID,
			Scope:       "project",
			ProjectRoot: ".",
			Backup:      true,
			Aspects:     aspects,
		})
		if err != nil {
			return err
		}
		return *result
	}
}

// View renders the current screen.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.loading && m.screen == ScreenDetect {
		return m.viewLoading()
	}

	switch m.screen {
	case ScreenDetect:
		return m.viewDetect()
	case ScreenSelectSource:
		return m.viewSelect("source")
	case ScreenSelectTarget:
		return m.viewSelect("target")
	case ScreenAspects:
		return m.viewAspects()
	case ScreenPreview:
		return m.viewProcessing()
	case ScreenResult:
		return m.viewResult()
	}

	return ""
}

func (m Model) viewLoading() string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		fmt.Sprintf("%s Scanning for coding agents...", m.spinner.View()),
	)
}

func (m Model) viewDetect() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render("🔄 Shifter — Agent Detection"))
	b.WriteString("\n\n")

	for _, r := range m.detectResults {
		if r.Found {
			parts := []string{"✓", r.Name}
			for k, v := range r.Summary {
				parts = append(parts, fmt.Sprintf("%d %s", v, k))
			}
			b.WriteString(styles.Check.String())
			b.WriteString(" ")
			b.WriteString(r.Name)
			b.WriteString("\n")
		} else {
			b.WriteString(styles.Cross.String())
			b.WriteString(" ")
			b.WriteString(r.Name)
			b.WriteString(" (not configured)")
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render("Press Enter to continue • q to quit"))
	return b.String()
}

func (m Model) viewSelect(kind string) string {
	var b strings.Builder
	title := "Select Source Agent"
	if kind == "target" {
		title = "Select Target Agent"
		if m.sourceID != "" {
			srcName := m.sourceID
			for _, r := range m.detectResults {
				if r.ID == m.sourceID {
					srcName = r.Name
				}
			}
			title += fmt.Sprintf(" (from: %s)", srcName)
		}
	}

	b.WriteString(styles.Title.Render(title))
	b.WriteString("\n\n")

	for i, r := range m.detectResults {
		if kind == "target" && r.ID == m.sourceID {
			// Skip source agent in target selection
			b.WriteString(styles.MutedText.Render(fmt.Sprintf("    %s (source)", r.Name)))
			b.WriteString("\n")
			continue
		}

		prefix := "  "
		if i == m.cursorIdx {
			prefix = "❯ "
			b.WriteString(styles.ActiveItem.Render(prefix + r.Name))
		} else {
			b.WriteString(styles.InactiveItem.Render(prefix + r.Name))
		}

		if r.Found {
			for k, v := range r.Summary {
				b.WriteString(fmt.Sprintf(" (%d %s)", v, k))
			}
		} else {
			b.WriteString(" (not configured)")
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render("↑↓ navigate • Enter select • q back"))
	return b.String()
}

func (m Model) viewAspects() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render(fmt.Sprintf("Port: %s → %s", m.sourceID, m.targetID)))
	b.WriteString("\n\n")
	b.WriteString("Select aspects to port:\n\n")

	for i, a := range m.aspects {
		checkbox := "[ ]"
		if a.Selected {
			checkbox = "[✓]"
		}

		line := fmt.Sprintf("  %s %s", checkbox, a.Label)
		if i == m.cursorIdx {
			line = "❯ " + line[2:]
			b.WriteString(styles.ActiveItem.Render(line))
		} else {
			b.WriteString(styles.InactiveItem.Render(line))
		}

		if a.Count > 0 {
			b.WriteString(fmt.Sprintf(" (%d found)", a.Count))
		}

		if a.Lossy {
			b.WriteString(" ")
			b.WriteString(styles.LossWarn.Render("⚠ lossy: " + a.LossReason))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render("↑↓ navigate • space toggle • Enter port • q back"))
	return b.String()
}

func (m Model) viewProcessing() string {
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		fmt.Sprintf("%s Porting configuration from %s to %s...\n\nThis may take a moment.", m.spinner.View(), m.sourceID, m.targetID),
	)
}

func (m Model) viewResult() string {
	var b strings.Builder

	if m.portErr != nil {
		b.WriteString(styles.Title.Render("❌ Port Failed"))
		b.WriteString("\n\n")
		b.WriteString(styles.Cross.String() + " " + m.portErr.Error())
		b.WriteString("\n\n")
		b.WriteString(styles.HelpBar.Render("Press Enter or q to quit"))
		return b.String()
	}

	if m.portResult == nil {
		return ""
	}

	b.WriteString(styles.Title.Render("✓ Port Complete"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("Source: %s → Target: %s\n\n", m.portResult.Source, m.portResult.Target))

	if len(m.portResult.FilesWritten) > 0 {
		b.WriteString("Files written:\n")
		for _, f := range m.portResult.FilesWritten {
			b.WriteString(fmt.Sprintf("  %s %s\n", styles.Check, f))
		}
	}

	if len(m.portResult.LossWarnings) > 0 {
		b.WriteString("\nLoss warnings:\n")
		for _, w := range m.portResult.LossWarnings {
			b.WriteString(fmt.Sprintf("  %s [%s] %s: %s\n", styles.WarnIcon, w.Severity, w.Feature, w.Reason))
		}
	}

	if rc := m.portResult.Config; rc != nil {
		b.WriteString(fmt.Sprintf("\nSummary: %d agents, %d skills, %d commands, %d MCP servers, %d hooks",
			len(rc.Agents), len(rc.Skills), len(rc.Commands), len(rc.MCPServers), len(rc.Hooks)))
	}

	b.WriteString("\n\n")
	b.WriteString(styles.HelpBar.Render("Press Enter or q to quit"))
	return b.String()
}
