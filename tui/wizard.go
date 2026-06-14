package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/engine/detect"
	"github.com/moximxxx/shifter/engine/port"
	"github.com/moximxxx/shifter/engine/profile"
	"github.com/moximxxx/shifter/pkg/i18n"
	"github.com/moximxxx/shifter/pkg/logo"
	"github.com/moximxxx/shifter/pkg/settings"
	"github.com/moximxxx/shifter/registry"
	"github.com/moximxxx/shifter/tui/styles"
)

// WizardScreen represents the wizard flow.
type WizardScreen int

const (
	WizWelcome WizardScreen = iota
	WizMenu
	WizSaveSelectAgent
	WizSaveName
	WizSaveDone
	WizLoadSelectProfile
	WizLoadSelectTarget
	WizLoadDone
	WizPortSelectSource
	WizPortSelectTarget
	WizPortAspects
	WizPortDone
	WizSettings
)

// WizardModel is the interactive config wizard.
type WizardModel struct {
	screen    WizardScreen
	prevScreen WizardScreen
	width     int
	height    int
	quitting  bool
	errorMsg  string
	loading   bool
	doneMsg   string

	// First-run
	isFirstRun bool
	langChoice int // 0=en, 1=zh

	// Detection
	detectResults []detect.Result

	// Profiles
	profileList []profile.Profile
	profilesLoaded bool

	// Selection state
	cursorIdx  int
	inputText  string // for name input
	inputMode  bool

	// Selected values
	selectedAgentID string
	selectedProfile *profile.Profile
	selectedTarget  string

	// Aspect checkboxes
	aspects   []aspectItem
	sourceID  string
	targetID  string
}

// TUI version string
var tuiVersion = "dev"

// SetVersion sets the version string for the TUI.
func SetVersion(v string) { tuiVersion = v }

// NewWizardModel creates the interactive wizard.
func NewWizardModel() WizardModel {
	m := WizardModel{
		screen:    WizWelcome,
		cursorIdx: 0,
	}
	// Check if first run
	if settings.IsFirstRun() {
		m.isFirstRun = true
	} else {
		// Skip welcome, go directly to menu
		m.screen = WizMenu
	}
	return m
}

func (m WizardModel) Init() tea.Cmd {
	return scanAgentsCmd
}

func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case scanDoneMsg:
		m.detectResults = msg.results

	case profileListMsg:
		m.profileList = msg.profiles
		m.profilesLoaded = true

	case profileSaveMsg:
		m.loading = false
		m.doneMsg = fmt.Sprintf("✓ Profile %q saved!\n\n%s", msg.name, msg.summary)
		m.screen = WizSaveDone

	case portDoneMsg:
		m.loading = false
		if msg.result != nil {
			m.doneMsg = fmt.Sprintf("✓ Port Complete: %s → %s\n\n%d files written, %d warnings",
				msg.result.Source, msg.result.Target,
				len(msg.result.FilesWritten), len(msg.result.LossWarnings))
		}
		m.screen = WizPortDone

	case loadDoneMsg:
		m.loading = false
		m.doneMsg = fmt.Sprintf("✓ Profile %q applied to %s\n\n%d files written",
			msg.profileName, msg.target, len(msg.files))
		m.screen = WizLoadDone

	case error:
		m.loading = false
		m.errorMsg = msg.Error()
		m.doneMsg = "❌ Error: " + m.errorMsg
		m.screen = WizSaveDone

	case tea.KeyMsg:
		if m.inputMode {
			return m.handleInputMode(msg)
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			if m.screen == WizWelcome {
				// Default to English and quit welcome
				m.quitting = true
				return m, tea.Quit
			}
			if m.screen == WizMenu {
				m.quitting = true
				return m, tea.Quit
			}
			m.prevScreen, m.screen = m.screen, m.prevScreen
			m.cursorIdx = 0
			m.errorMsg = ""
			return m, nil

		case "q":
			if m.screen == WizWelcome || m.screen == WizMenu || m.screen == WizSaveDone || m.screen == WizLoadDone || m.screen == WizPortDone {
				m.quitting = true
				return m, tea.Quit
			}

		case "up", "k":
			if m.screen == WizWelcome || m.screen == WizSettings {
				if m.langChoice > 0 {
					m.langChoice--
				}
				return m, nil
			}
			if m.cursorIdx > 0 {
				m.cursorIdx--
			}
		case "down", "j":
			if m.screen == WizWelcome || m.screen == WizSettings {
				if m.langChoice < 1 {
					m.langChoice++
				}
				return m, nil
			}
			if m.canMoveDown() {
				m.cursorIdx++
			}
		case "enter":
			return m.handleEnter()
		case " ":
			if m.screen == WizPortAspects && m.cursorIdx < len(m.aspects) {
				m.aspects[m.cursorIdx].Selected = !m.aspects[m.cursorIdx].Selected
			}
		}
	}

	return m, nil
}

func (m *WizardModel) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.inputText != "" {
			m.inputMode = false
			return m.handleSaveProfile()
		}
	case "backspace":
		if len(m.inputText) > 0 {
			m.inputText = m.inputText[:len(m.inputText)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.inputText += msg.String()
		}
	}
	return m, nil
}

func (m *WizardModel) canMoveDown() bool {
	switch m.screen {
	case WizWelcome:
		return m.cursorIdx < 1 // en, zh
	case WizMenu:
		return m.cursorIdx < 3 // Save, Load, Port, Settings
	case WizSaveSelectAgent, WizPortSelectSource, WizPortSelectTarget:
		return m.cursorIdx < len(m.detectResults)-1
	case WizLoadSelectProfile:
		return m.cursorIdx < len(m.profileList)-1
	case WizLoadSelectTarget:
		return m.cursorIdx < len(m.detectResults)-1
	case WizPortAspects:
		return m.cursorIdx < len(m.aspects)-1
	case WizSettings:
		return m.cursorIdx < 1 // en, zh
	}
	return false
}

func (m *WizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case WizWelcome:
		// Save language choice
		lang := "en"
		if m.langChoice == 1 {
			lang = "zh"
		}
		i18n.SetLang(lang)
		s, _ := settings.Load()
		if s != nil {
			s.Lang = lang
			s.FirstRun = false
			settings.Save(s)
		}
		m.screen = WizMenu
		m.cursorIdx = 0
		m.isFirstRun = false
		return m, nil

	case WizMenu:
		switch m.cursorIdx {
		case 0: // Save
			m.prevScreen = WizMenu
			m.screen = WizSaveSelectAgent
			m.cursorIdx = 0
		case 1: // Load
			m.prevScreen = WizMenu
			m.screen = WizLoadSelectProfile
			m.cursorIdx = 0
			if !m.profilesLoaded {
				return m, loadProfilesCmd
			}
		case 2: // Port
			m.prevScreen = WizMenu
			m.screen = WizPortSelectSource
			m.cursorIdx = m.firstFoundIdx()
		case 3: // Settings
			m.prevScreen = WizMenu
			m.screen = WizSettings
			m.cursorIdx = 0
			// Preselect current language
			if i18n.Lang() == "zh" {
				m.langChoice = 1
			} else {
				m.langChoice = 0
			}
		}
		return m, nil

	case WizSaveSelectAgent:
		if f := m.foundAgents(); m.cursorIdx < len(f) {
			m.selectedAgentID = f[m.cursorIdx].ID
			m.prevScreen = WizSaveSelectAgent
			m.screen = WizSaveName
			m.inputMode = true
			m.inputText = ""
		}
		return m, nil

	case WizPortSelectSource:
		if f := m.foundAgents(); m.cursorIdx < len(f) {
			m.sourceID = f[m.cursorIdx].ID
			m.prevScreen = WizPortSelectSource
			m.screen = WizPortSelectTarget
			m.cursorIdx = 0
		}
		return m, nil

	case WizPortSelectTarget:
		if f := m.foundAgents(); m.cursorIdx < len(f) {
			if f[m.cursorIdx].ID != m.sourceID {
				m.targetID = f[m.cursorIdx].ID
				m.prevScreen = WizPortSelectTarget
				m.screen = WizPortAspects
				m.cursorIdx = 0
				m.buildPortAspects()
			}
		}
		return m, nil

	case WizPortAspects:
		m.prevScreen = WizPortAspects
		return m, m.executeWizardPort()

	case WizLoadSelectProfile:
		if m.cursorIdx < len(m.profileList) {
			p := m.profileList[m.cursorIdx]
			m.selectedProfile = &p
			m.prevScreen = WizLoadSelectProfile
			m.screen = WizLoadSelectTarget
			m.cursorIdx = 0
		}
		return m, nil

	case WizLoadSelectTarget:
		if f := m.foundAgents(); m.cursorIdx < len(f) {
			m.selectedTarget = f[m.cursorIdx].ID
			return m, m.executeWizardLoad()
		}
		return m, nil

	case WizSettings:
		// Save language choice and go back to menu
		lang := "en"
		if m.langChoice == 1 {
			lang = "zh"
		}
		i18n.SetLang(lang)
		s, _ := settings.Load()
		if s != nil {
			s.Lang = lang
			s.FirstRun = false
			settings.Save(s)
		}
		m.screen = WizMenu
		m.cursorIdx = 3
		return m, nil
	}

	return m, nil
}

func (m *WizardModel) handleSaveProfile() (tea.Model, tea.Cmd) {
	m.prevScreen = WizSaveName
	m.loading = true

	return m, func() tea.Msg {
		ctx := context.Background()
		a, err := registry.Get(m.selectedAgentID)
		if err != nil {
			return err
		}

		cfg, err := a.Read(ctx, adapter.ReadOptions{
			Scope:       "project",
			ProjectRoot: ".",
		})
		if err != nil {
			return err
		}

		p := profile.Profile{
			Name:        m.inputText,
			Description: fmt.Sprintf("%s config captured from project", a.Name()),
			SourceAgent: m.selectedAgentID,
			Config:      cfg,
			CreatedAt:   time.Now(),
		}

		if err := profile.Save(p); err != nil {
			return err
		}

		return profileSaveMsg{name: m.inputText, summary: p.Summary()}
	}
}

type profileSaveMsg struct {
	name    string
	summary string
}

func (m *WizardModel) executeWizardPort() tea.Cmd {
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
		return portDoneMsg{result: result}
	}
}

type portDoneMsg struct {
	result *port.Result
}

func (m *WizardModel) executeWizardLoad() tea.Cmd {
	return func() tea.Msg {
		if m.selectedProfile == nil || m.selectedProfile.Config == nil {
			return fmt.Errorf("no profile selected")
		}

		ctx := context.Background()
		tgtAdapter, err := registry.Get(m.selectedTarget)
		if err != nil {
			return err
		}

		writeResult, err := tgtAdapter.Write(ctx, m.selectedProfile.Config, adapter.WriteOptions{
			Scope:       "project",
			ProjectRoot: ".",
			Backup:      true,
		})
		if err != nil {
			return err
		}

		return loadDoneMsg{
			profileName: m.selectedProfile.Name,
			target:      m.selectedTarget,
			files:       writeResult.FilesWritten,
			warnings:    writeResult.LossWarnings,
		}
	}
}

type loadDoneMsg struct {
	profileName string
	target      string
	files       []string
	warnings    []canonical.LossWarning
}

type profileListMsg struct {
	profiles []profile.Profile
}

func loadProfilesCmd() tea.Msg {
	profiles, _ := profile.List()
	return profileListMsg{profiles: profiles}
}

func (m *WizardModel) firstFoundIdx() int {
	for i, r := range m.detectResults {
		if r.Found {
			return i
		}
	}
	return 0
}

func (m *WizardModel) foundAgents() []detect.Result {
	var f []detect.Result
	for _, r := range m.detectResults {
		if r.Found {
			f = append(f, r)
		}
	}
	return f
}

func (m *WizardModel) buildPortAspects() {
	// Find source summary
	var srcSummary map[string]int
	for _, r := range m.detectResults {
		if r.ID == m.sourceID {
			srcSummary = r.Summary
			break
		}
	}

	m.aspects = []aspectItem{
		{Name: "instructions", Label: i18n.T("aspect.instructions"), Selected: true, Count: srcSummary["instructions"]},
		{Name: "agents", Label: i18n.T("aspect.agents"), Selected: true, Count: srcSummary["agents"]},
		{Name: "skills", Label: i18n.T("aspect.skills"), Selected: true, Count: srcSummary["skills"]},
		{Name: "commands", Label: i18n.T("aspect.commands"), Selected: true, Count: srcSummary["commands"]},
		{Name: "mcp", Label: i18n.T("aspect.mcp"), Selected: true},
		{Name: "permissions", Label: i18n.T("aspect.permissions"), Selected: true},
		{Name: "hooks", Label: i18n.T("aspect.hooks"), Selected: true},
		{Name: "settings", Label: i18n.T("aspect.settings"), Selected: false},
	}
}

// View renders the wizard screen.
func (m WizardModel) View() string {
	if m.quitting {
		return ""
	}

	switch {
	case m.doneMsg != "":
		return m.viewDone()
	default:
		return m.viewCurrentScreen()
	}
}

func (m WizardModel) viewCurrentScreen() string {
	switch m.screen {
	case WizWelcome:
		return m.viewWelcome()
	case WizMenu:
		return m.viewMenu()
	case WizSaveSelectAgent:
		return m.viewAgentSelect(i18n.T("save.title"), i18n.T("save.subtitle"))
	case WizSaveName:
		return m.viewNameInput()
	case WizPortSelectSource:
		return m.viewAgentSelect(i18n.T("port.title"), i18n.T("port.source_subtitle"))
	case WizPortSelectTarget:
		return m.viewAgentSelect(i18n.Tf("port.target_title", map[string]string{"source": m.sourceID}), i18n.T("port.target_subtitle"))
	case WizPortAspects:
		return m.viewAspectsSelect()
	case WizLoadSelectProfile:
		return m.viewProfileSelect()
	case WizLoadSelectTarget:
		return m.viewAgentSelect(i18n.T("load.select_target"), i18n.T("load.target_subtitle"))
	case WizSettings:
		return m.viewSettings()
	}
	return ""
}

func (m WizardModel) viewDone() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render(m.doneMsg))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("result.press_quit")))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
}

func (m WizardModel) viewWelcome() string {
	var b strings.Builder
	b.WriteString(logo.Render(""))
	b.WriteString("\n")
	b.WriteString(styles.MutedText.Render(tuiVersion))
	b.WriteString("\n\n")
	b.WriteString(styles.Title.Render("🔄 " + i18n.T("welcome.title")))
	b.WriteString("\n\n")
	b.WriteString(i18n.T("welcome.select_lang"))
	b.WriteString("\n\n")

	langs := []string{i18n.T("welcome.lang_en"), i18n.T("welcome.lang_zh")}
	for i, name := range langs {
		prefix := "  "
		if i == m.langChoice {
			prefix = "❯ "
			b.WriteString(styles.ActiveItem.Render(prefix + name))
		} else {
			b.WriteString(styles.InactiveItem.Render(prefix + name))
		}
		if i == 0 {
			b.WriteString("  🇺🇸")
		} else {
			b.WriteString("  🇨🇳")
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.MutedText.Render(i18n.T("welcome.later")))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.navigate") + " • " + i18n.T("help.select")))
	return b.String()
}

func (m WizardModel) viewMenu() string {
	var b strings.Builder
	b.WriteString(logo.Render(""))
	b.WriteString("\n")
	b.WriteString(styles.MutedText.Render(tuiVersion))
	b.WriteString("\n\n")
	b.WriteString(styles.Title.Render("🔄 " + i18n.T("menu.title")))
	b.WriteString("\n\n")

	agents := m.foundAgents()
	if len(agents) == 0 {
		b.WriteString(i18n.T("detect.not_configured") + "\n")
		b.WriteString("\n" + styles.HelpBar.Render(i18n.T("help.quit")))
		return b.String()
	}

	b.WriteString(i18n.Tf("menu.found_agents", map[string]string{"count": fmt.Sprintf("%d", len(agents))}))
	b.WriteString("\n")
	for _, a := range agents {
		b.WriteString(fmt.Sprintf("  ✓ %s", a.Name))
		for k, v := range a.Summary {
			b.WriteString(fmt.Sprintf(" (%d %s)", v, k))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(i18n.T("menu.what_do") + "\n\n")

	items := []string{
		i18n.T("menu.save"),
		i18n.T("menu.load"),
		i18n.T("menu.port"),
		"⚙  " + i18n.T("settings.title") + " — Change language and preferences",
	}

	for i, item := range items {
		if i == m.cursorIdx {
			b.WriteString(styles.ActiveItem.Render("❯ " + item))
		} else {
			b.WriteString(styles.InactiveItem.Render("  " + item))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.navigate") + " • " + i18n.T("help.select") + " • " + i18n.T("help.quit")))
	return b.String()
}

func (m WizardModel) viewAgentSelect(title, subtitle string) string {
	var b strings.Builder
	b.WriteString(styles.Title.Render(title))
	b.WriteString("\n\n")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	agents := m.foundAgents()
	for i, a := range agents {
		line := fmt.Sprintf("  %s", a.Name)
		for k, v := range a.Summary {
			line += fmt.Sprintf(" (%d %s)", v, k)
		}
		if i == m.cursorIdx {
			b.WriteString(styles.ActiveItem.Render("❯ " + line))
		} else {
			b.WriteString(styles.InactiveItem.Render("  " + line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.navigate") + " • " + i18n.T("help.select") + " • " + i18n.T("help.back")))
	return b.String()
}

func (m WizardModel) viewNameInput() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render("💾 Save Profile — Name"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("Saving config from: %s\n\n", m.selectedAgentID))
	b.WriteString("Profile name: ")
	b.WriteString(styles.ActiveItem.Render(m.inputText))
	if !m.inputMode {
		b.WriteString("_")
	}
	b.WriteString("\n\n")
	b.WriteString(styles.MutedText.Render("(letters, numbers, hyphens, underscores)"))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.type_name")))
	return b.String()
}

func (m WizardModel) viewProfileSelect() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render("📥 Load Profile — Select"))
	b.WriteString("\n\n")

	if !m.profilesLoaded {
		b.WriteString("Loading profiles...\n")
		return b.String()
	}

	if len(m.profileList) == 0 {
		b.WriteString(i18n.T("load.no_profiles")+"\n\n")
		b.WriteString("Use '💾 Save' from the main menu to create one.\n")
		b.WriteString("\n" + styles.HelpBar.Render(i18n.T("help.back")))
		return b.String()
	}

	b.WriteString(i18n.T("load.subtitle")+"\n\n")

	for i, p := range m.profileList {
		line := fmt.Sprintf("%s", p.Name)
		if p.SourceAgent != "" {
			line += fmt.Sprintf("  (from %s)", p.SourceAgent)
		}
		line += "\n    " + styles.MutedText.Render(p.Summary())
		if p.Description != "" {
			line += "  " + styles.MutedText.Render(p.Description)
		}

		if i == m.cursorIdx {
			b.WriteString(styles.ActiveItem.Render("❯ " + line))
		} else {
			b.WriteString(styles.InactiveItem.Render("  " + line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.navigate") + " • " + i18n.T("help.select") + " • " + i18n.T("help.back")))
	return b.String()
}

func (m WizardModel) viewSettings() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render("⚙ " + i18n.T("settings.title")))
	b.WriteString("\n\n")
	b.WriteString(i18n.T("settings.language"))
	b.WriteString("\n\n")

	langs := []string{i18n.T("welcome.lang_en"), i18n.T("welcome.lang_zh")}
	for i, name := range langs {
		prefix := "  "
		if i == m.langChoice {
			prefix = "❯ "
			b.WriteString(styles.ActiveItem.Render(prefix + name))
		} else {
			b.WriteString(styles.InactiveItem.Render(prefix + name))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.MutedText.Render(i18n.T("settings.restart_hint")))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.navigate") + " • " + i18n.T("help.select") + " • " + i18n.T("help.back")))
	return b.String()
}

func (m WizardModel) viewAspectsSelect() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render(fmt.Sprintf("🔀 Port: %s → %s", m.sourceID, m.targetID)))
	b.WriteString("\n\n")
	b.WriteString(i18n.T("port.aspects_subtitle")+"\n\n")

	for i, a := range m.aspects {
		checkbox := "[ ]"
		if a.Selected {
			checkbox = "[✓]"
		}
		line := fmt.Sprintf("%s %s", checkbox, a.Label)
		if a.Count > 0 {
			line += fmt.Sprintf(" (%d found)", a.Count)
		}
		if a.Lossy {
			line += " " + styles.LossWarn.Render("⚠ lossy")
		}
		if i == m.cursorIdx {
			b.WriteString(styles.ActiveItem.Render("❯ " + line))
		} else {
			b.WriteString(styles.InactiveItem.Render("  " + line))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.navigate") + " • " + i18n.T("help.toggle") + " • " + i18n.T("help.apply") + " • " + i18n.T("help.back")))
	return b.String()
}
