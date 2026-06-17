package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/engine/detect"
	"github.com/moximxxx/shifter/engine/flowhub"
	"github.com/moximxxx/shifter/engine/port"
	"github.com/moximxxx/shifter/engine/profile"
	"github.com/moximxxx/shifter/pkg/i18n"
	"github.com/moximxxx/shifter/pkg/logo"
	"github.com/moximxxx/shifter/pkg/settings"
	"github.com/moximxxx/shifter/registry"
	"github.com/moximxxx/shifter/tui/styles"
)

// aspectItem represents a configurable aspect for porting.
type aspectItem struct {
	Name       string
	Label      string
	Selected   bool
	Lossy      bool
	LossReason string
	Count      int
}

// scanAgentsCmd is a tea.Cmd that scans for coding agents.
func scanAgentsCmd() tea.Msg {
	results := detect.ScanAll()
	return scanDoneMsg{results: results}
}

// scanDoneMsg carries the scan results.
type scanDoneMsg struct {
	results []detect.Result
}

// WizardScreen represents the wizard flow.
type WizardScreen int

const (
	WizWelcome WizardScreen = iota
	WizMenu
	WizSaveSelectAgent
	WizSaveName
	WizSaveDesc
	WizSaveDone
	WizLoadSelectProfile
	WizLoadSelectTarget
	WizLoadDone
	WizPortSelectSource
	WizPortSelectTarget
	WizPortAspects
	WizPortDone
	WizSettings
	WizTemplates
	WizTemplateDetail
	WizFlowHub
)

// WizardModel is the interactive config wizard.
type WizardModel struct {
	screen    WizardScreen
	prevScreen WizardScreen
	backStack []WizardScreen // navigation history for Esc
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
	saveDesc        string // description for save workflow

	// Aspect checkboxes
	aspects   []aspectItem
	sourceID  string
	targetID  string

	// Result banner shown on menu after operation
	resultBanner string

	// FlowHub
	flowhubResults []flowhub.Workflow
	flowhubLoaded bool
	flowhubQuery  string
	flowhubCursor int
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

	case flowhubResultsMsg:
		m.flowhubResults = msg.results
		m.flowhubLoaded = true

	case profileSaveMsg:
		m.loading = false
		m.resultBanner = fmt.Sprintf("✓ Profile %q saved (%s)", msg.name, msg.summary)
		m.screen = WizMenu
		m.cursorIdx = 0

	case portDoneMsg:
		m.loading = false
		if msg.result != nil {
			m.resultBanner = fmt.Sprintf("✓ Port: %s → %s — %d files, %d warnings",
				msg.result.Source, msg.result.Target,
				len(msg.result.FilesWritten), len(msg.result.LossWarnings))
		}
		m.screen = WizMenu
		m.cursorIdx = 0

	case loadDoneMsg:
		m.loading = false
		m.resultBanner = fmt.Sprintf("✓ Profile %q applied to %s (%d files)",
			msg.profileName, msg.target, len(msg.files))
		m.screen = WizMenu
		m.cursorIdx = 0

	case error:
		m.loading = false
		m.resultBanner = "❌ " + msg.Error()
		m.screen = WizMenu
		m.cursorIdx = 0

	case tea.KeyMsg:
		if m.inputMode {
			return m.handleInputMode(msg)
		}
		// FlowHub: backspace removes last char
		if m.screen == WizFlowHub && msg.String() == "backspace" {
			if len(m.flowhubQuery) > 0 {
				m.flowhubQuery = m.flowhubQuery[:len(m.flowhubQuery)-1]
				m.flowhubCursor = 0
			}
			return m, nil
		}
		// FlowHub: typing directly filters — update query and reset cursor
		if m.screen == WizFlowHub && len(msg.String()) == 1 && msg.String() != " " {
			m.flowhubQuery += msg.String()
			m.flowhubCursor = 0
			return m, nil
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
			// Pop from back stack to go back
			if len(m.backStack) > 0 {
				m.screen = m.backStack[len(m.backStack)-1]
				m.backStack = m.backStack[:len(m.backStack)-1]
			}
			m.cursorIdx = 0
			m.errorMsg = ""
			return m, nil

		case "q":
			if m.screen == WizWelcome || m.screen == WizMenu {
				m.quitting = true
				return m, tea.Quit
			}

		case "up", "k":
			if m.screen == WizFlowHub && m.flowhubCursor > 0 {
				m.flowhubCursor--
				return m, nil
			}
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
			if m.screen == WizFlowHub {
				filtered := getFiltered(m)
				if m.flowhubCursor < len(filtered)-1 {
					m.flowhubCursor++
				}
				return m, nil
			}
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
			if m.screen == WizFlowHub && m.inputMode {
				m.inputMode = false
				return m, nil
			}
			return m.handleEnter()
		case "ctrl+p":
			if m.screen == WizTemplateDetail && m.selectedProfile != nil && m.selectedProfile.Config != nil {
				return m, publishToFlowHubCmd(m.selectedProfile)
			}
			return m, nil
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
		if m.screen == WizFlowHub {
			m.inputMode = false
			return m, nil
		}
		if m.inputText == "" {
			return m, nil
		}
		m.inputMode = false
		if m.screen == WizSaveName {
			// Name done → go to description input
			m.saveDesc = "" // reset
			m.screen = WizSaveDesc
			m.inputMode = true
			m.inputText = ""
			return m, nil
		}
		// Description done → save
		return m.handleSaveProfile()
	case "esc", "ctrl+c":
		// Cancel input — go back to menu
		m.inputMode = false
		m.screen = WizMenu
		m.cursorIdx = 0
		m.inputText = ""
		return m, nil
	case "backspace":
		if m.screen == WizFlowHub {
			if len(m.flowhubQuery) > 0 {
				m.flowhubQuery = m.flowhubQuery[:len(m.flowhubQuery)-1]
			}
			return m, nil
		}
		if len(m.inputText) > 0 {
			m.inputText = m.inputText[:len(m.inputText)-1]
		}
	default:
		if m.screen == WizFlowHub {
			if len(msg.String()) == 1 {
				m.flowhubQuery += msg.String()
			}
			return m, nil
		}
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
		return m.cursorIdx < 5 // Save, Load, Port, Templates, FlowHub, Settings
	case WizTemplates:
		return m.cursorIdx < len(m.profileList)-1
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
			m.backStack = append(m.backStack, m.screen)
			m.cursorIdx = 0
			agents := m.foundAgents()
			if len(agents) == 0 {
				return m, nil
			}
			if len(agents) == 1 {
				m.selectedAgentID = agents[0].ID
				m.screen = WizSaveName
				m.inputMode = true
				m.inputText = ""
				return m, nil
			}
			m.screen = WizSaveSelectAgent
		case 1: // Load
			m.prevScreen = WizMenu
			m.backStack = append(m.backStack, m.screen)
			m.screen = WizLoadSelectProfile
			m.cursorIdx = 0
			if !m.profilesLoaded {
				return m, loadProfilesCmd
			}
		case 2: // Port
			m.prevScreen = WizMenu
			m.backStack = append(m.backStack, m.screen)
			m.cursorIdx = 0
			agents := m.foundAgents()
			if len(agents) == 0 {
				return m, nil
			}
			if len(agents) == 1 {
				m.sourceID = agents[0].ID
				m.screen = WizPortSelectTarget
				return m, nil
			}
			m.cursorIdx = m.firstFoundIdx()
			m.screen = WizPortSelectSource
		case 3: // Templates
			m.prevScreen = WizMenu
			m.backStack = append(m.backStack, m.screen)
			m.screen = WizTemplates
			m.cursorIdx = 0
			if !m.profilesLoaded {
				return m, loadProfilesCmd
			}
		case 4: // FlowHub
			m.backStack = append(m.backStack, m.screen)
			m.screen = WizFlowHub
			m.cursorIdx = 0
			if !m.flowhubLoaded {
				return m, fetchFlowHubCmd
			}
			return m, nil
		case 5: // Settings
			m.prevScreen = WizMenu
			m.backStack = append(m.backStack, m.screen)
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

	case WizTemplates:
		if m.cursorIdx < len(m.profileList) {
			m.selectedProfile = &m.profileList[m.cursorIdx]
			m.prevScreen = WizTemplates
			m.backStack = append(m.backStack, m.screen)
			m.screen = WizTemplateDetail
		}
		return m, nil

	case WizTemplateDetail:
		m.screen = WizTemplates
		return m, nil

	case WizFlowHub:
		filtered := getFlowHubFiltered(m.flowhubResults, m.flowhubQuery)
		if m.flowhubCursor < len(filtered) {
			w := filtered[m.flowhubCursor]
			m.resultBanner = fmt.Sprintf("To install: shifter flow install %s --to <agent>", w.Name)
			m.screen = WizMenu
			m.cursorIdx = 0
		}
		return m, nil

	case WizSaveSelectAgent:
		if f := m.foundAgents(); m.cursorIdx < len(f) {
			m.selectedAgentID = f[m.cursorIdx].ID
			m.prevScreen = WizSaveSelectAgent
			m.backStack = append(m.backStack, m.screen)
			m.screen = WizSaveName
			m.inputMode = true
			m.inputText = ""
		}
		return m, nil

	case WizPortSelectSource:
		if f := m.foundAgents(); m.cursorIdx < len(f) {
			m.sourceID = f[m.cursorIdx].ID
			m.backStack = append(m.backStack, m.screen)
				m.screen = WizPortSelectTarget
			m.cursorIdx = 0
			// Keep prevScreen pointing to menu so Esc goes back to menu
		}
		return m, nil

	case WizPortSelectTarget:
		if f := m.foundAgents(); m.cursorIdx < len(f) {
			if f[m.cursorIdx].ID != m.sourceID {
				m.targetID = f[m.cursorIdx].ID
				m.prevScreen = WizPortSelectTarget
				m.backStack = append(m.backStack, m.screen)
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
			m.backStack = append(m.backStack, m.screen)
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
			Description: m.saveDesc,
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

type flowhubResultsMsg struct {
	results []flowhub.Workflow
}

func fetchFlowHubCmd() tea.Msg {
	results, _ := flowhub.Search("")
	return flowhubResultsMsg{results: results}
}

func publishToFlowHubCmd(p *profile.Profile) tea.Cmd {
	return func() tea.Msg {
		if p == nil || p.Config == nil {
			return fmt.Errorf("no config to publish")
		}
		workflowJSON, err := json.MarshalIndent(p.Config, "", "  ")
		if err != nil {
			return err
		}
		meta := flowhub.GenerateMetadata(p.Name, p.SourceAgent, p.Description, nil)
		metaJSON, _ := json.MarshalIndent(meta, "", "  ")

		files := map[string][]byte{
			"workflow.shifter.json": workflowJSON,
			"metadata.json":         metaJSON,
		}
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			token = os.Getenv("GH_TOKEN")
		}
		if token == "" {
			// Try gh CLI config
			home, _ := os.UserHomeDir()
			data, _ := os.ReadFile(home + "/.config/gh/hosts.yml")
			if data != nil {
				for _, line := range strings.Split(string(data), "\n") {
					if strings.Contains(line, "oauth_token:") || strings.Contains(line, "token:") {
						parts := strings.SplitN(line, ":", 2)
						if len(parts) == 2 {
							token = strings.TrimSpace(parts[1])
							break
						}
					}
				}
			}
		}
		if token == "" {
			return fmt.Errorf("GITHUB_TOKEN not set. Create one at https://github.com/settings/tokens\n\nThen: export GITHUB_TOKEN=ghp_xxxx")
		}
		prURL, err := flowhub.Publish(flowhub.PublishRequest{
			Name:    p.Name,
			Files:   files,
			Message: fmt.Sprintf("Publish %s from shifter ui", p.Name),
			Token:   token,
		})
		if err != nil {
			return fmt.Errorf("publish: %w", err)
		}
		return fmt.Sprintf("✓ Published to FlowHub!\n  PR: %s", prURL)
	}
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
		// Only include agents with project-level config
		if r.Found && r.HasProjectConfig {
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

func (m WizardModel) viewHeader() string {
	var b strings.Builder
	b.WriteString(logo.Render(""))
	b.WriteString("\n")
	b.WriteString(styles.MutedText.Render(tuiVersion))
	b.WriteString("\n")
	return b.String()
}

func (m WizardModel) viewCurrentScreen() string {
	switch m.screen {
	case WizWelcome:
		return m.viewWelcome()
	case WizMenu:
		return m.viewHeader() + "\n" + m.viewMenu()
	case WizSaveSelectAgent:
		return m.viewHeader() + "\n" + m.viewAgentSelect(i18n.T("save.title"), i18n.T("save.subtitle"))
	case WizSaveName:
		return m.viewHeader() + "\n" + m.viewInput("💾 Profile Name", "save.name_prompt", "save.name_hint")
	case WizSaveDesc:
		return m.viewHeader() + "\n" + m.viewInput("💾 Profile Description", "save.desc_prompt", "save.desc_hint")
	case WizPortSelectSource:
		return m.viewHeader() + "\n" + m.viewAgentSelect(i18n.T("port.title"), i18n.T("port.source_subtitle"))
	case WizPortSelectTarget:
		return m.viewHeader() + "\n" + m.viewAgentSelect(i18n.Tf("port.target_title", map[string]string{"source": m.sourceID}), i18n.T("port.target_subtitle"))
	case WizPortAspects:
		return m.viewHeader() + "\n" + m.viewAspectsSelect()
	case WizLoadSelectProfile:
		return m.viewHeader() + "\n" + m.viewProfileSelect()
	case WizLoadSelectTarget:
		return m.viewHeader() + "\n" + m.viewAgentSelect(i18n.T("load.select_target"), i18n.T("load.target_subtitle"))
	case WizSettings:
		return m.viewHeader() + "\n" + m.viewSettings()
	case WizTemplates:
		return m.viewHeader() + "\n" + m.viewTemplates()
	case WizTemplateDetail:
		return m.viewHeader() + "\n" + m.viewTemplateDetail()
	case WizFlowHub:
		return m.viewHeader() + "\n" + m.viewFlowHub()
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
	// Result banner (from save/load/port operations)
	if m.resultBanner != "" {
		b.WriteString(styles.Border.Render(m.resultBanner))
		b.WriteString("\n\n")
	}

	b.WriteString(styles.Title.Render("🔄 " + i18n.T("menu.title")))
	b.WriteString("\n\n")

	agents := m.foundAgents()
	if len(agents) == 0 {
		b.WriteString(styles.MutedText.Render("⚠ " + i18n.T("detect.not_configured")))
		b.WriteString("\n\n")
	} else {
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
	}
	b.WriteString(i18n.T("menu.what_do") + "\n\n")

	items := []string{
		i18n.T("menu.save"),
		i18n.T("menu.load"),
		i18n.T("menu.port"),
		"📋 " + i18n.T("templates.title"),
		"🌐 FlowHub",
		"⚙  " + i18n.T("settings.title"),
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

func (m WizardModel) viewInput(title, promptKey, hintKey string) string {
	var b strings.Builder
	b.WriteString(styles.Title.Render(title))
	b.WriteString("\n\n")
	b.WriteString(i18n.T(promptKey))
	b.WriteString(": ")
	b.WriteString(styles.ActiveItem.Render(m.inputText))
	if !m.inputMode {
		b.WriteString("_")
	}
	b.WriteString("\n\n")
	b.WriteString(styles.MutedText.Render(i18n.T(hintKey)))
	b.WriteString("\n\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.type_name")))
	return b.String()
}

func (m WizardModel) viewNameInput() string {
	return m.viewInput(
		"💾 Save Profile — Name",
		"save.name_prompt",
		"save.name_hint",
	)
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

func (m WizardModel) viewTemplates() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render("📋 " + i18n.T("templates.title")))
	b.WriteString("\n\n")

	if len(m.profileList) == 0 {
		b.WriteString("No saved templates found.\n\n")
		b.WriteString("Use '💾 Save' from the main menu to create one.\n")
		b.WriteString("\n" + styles.HelpBar.Render(i18n.T("help.back")))
		return b.String()
	}

	for i, p := range m.profileList {
		line := fmt.Sprintf("%s", p.Name)
		if p.Description != "" {
			line += fmt.Sprintf(" — %s", p.Description)
		}
		if i == m.cursorIdx {
			b.WriteString(styles.ActiveItem.Render("❯ " + line))
		} else {
			b.WriteString(styles.InactiveItem.Render("  " + line))
		}
		b.WriteString("\n")
		b.WriteString(styles.MutedText.Render(fmt.Sprintf("      %s", p.Summary())))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("help.navigate")+" • "+i18n.T("help.select")+" • "+i18n.T("help.back")))
	return b.String()
}

func (m WizardModel) viewTemplateDetail() string {
	var b strings.Builder
	p := m.selectedProfile
	if p == nil || p.Config == nil {
		return "No profile selected"
	}

	b.WriteString(styles.Title.Render(fmt.Sprintf("📋 %s", p.Name)))
	b.WriteString("\n\n")

	if p.Description != "" {
		b.WriteString(p.Description)
		b.WriteString("\n\n")
	}
	b.WriteString(styles.MutedText.Render(fmt.Sprintf("Source: %s  |  Updated: %s", p.SourceAgent, p.UpdatedAt.Format("2006-01-02 15:04"))))
	b.WriteString("\n\n")

	cfg := p.Config
	b.WriteString(fmt.Sprintf("  Agents:      %d\n", len(cfg.Agents)))
	for _, a := range cfg.Agents {
		b.WriteString(fmt.Sprintf("    • %s — %s\n", a.Name, a.Description))
	}
	b.WriteString(fmt.Sprintf("  Skills:      %d\n", len(cfg.Skills)))
	for _, s := range cfg.Skills {
		b.WriteString(fmt.Sprintf("    • %s\n", s.Name))
	}
	b.WriteString(fmt.Sprintf("  Commands:    %d\n", len(cfg.Commands)))
	b.WriteString(fmt.Sprintf("  MCP Servers: %d\n", len(cfg.MCPServers)))
	for _, m := range cfg.MCPServers {
		b.WriteString(fmt.Sprintf("    • %s (%s)\n", m.Name, m.Type))
	}
	b.WriteString(fmt.Sprintf("  Hooks:       %d\n", len(cfg.Hooks)))
	if cfg.Permissions != nil {
		b.WriteString(fmt.Sprintf("  Permissions: %d allow, %d deny\n",
			len(cfg.Permissions.AllowRules), len(cfg.Permissions.DenyRules)))
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render("Ctrl+P publish to FlowHub  •  "+i18n.T("help.back")))
	return b.String()
}

func (m WizardModel) viewFlowHub() string {
	var b strings.Builder
	b.WriteString(styles.Title.Render("🌐 " + i18n.T("flowhub.title")))
	b.WriteString("\n\n")

	if !m.flowhubLoaded {
		b.WriteString("Loading workflows from FlowHub...\n")
		b.WriteString("\n" + styles.HelpBar.Render(i18n.T("help.back")))
		return b.String()
	}

	// Search box
	b.WriteString("🔍 " + i18n.T("flowhub.search") + ": ")
	b.WriteString(styles.ActiveItem.Render(m.flowhubQuery))
	if !m.inputMode {
		b.WriteString("_")
	}
	b.WriteString("\n\n")

	// Filter results
	var filtered []flowhub.Workflow
	q := strings.ToLower(m.flowhubQuery)
	for _, w := range m.flowhubResults {
		if q == "" || strings.Contains(strings.ToLower(w.Name), q) ||
			strings.Contains(strings.ToLower(w.Description), q) {
			filtered = append(filtered, w)
		}
	}

	if len(filtered) == 0 {
		b.WriteString(styles.MutedText.Render(i18n.T("flowhub.no_results")))
		b.WriteString("\n")
	} else {
		for i, w := range filtered {
			line := fmt.Sprintf("%s v%s  ⭐%d", w.Name, w.Version, w.Downloads)
			if i == m.flowhubCursor {
				b.WriteString(styles.ActiveItem.Render("❯ " + line))
			} else {
				b.WriteString(styles.InactiveItem.Render("  " + line))
			}
			b.WriteString("\n")
			b.WriteString(styles.MutedText.Render(fmt.Sprintf("      %s  |  %s", w.Description, w.Agent)))
			if len(w.Tags) > 0 {
				b.WriteString(styles.MutedText.Render(fmt.Sprintf("  |  %s", strings.Join(w.Tags, ", "))))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpBar.Render(i18n.T("flowhub.help")))
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

// getFiltered returns FlowHub results matching current query
func getFiltered(m WizardModel) []flowhub.Workflow {
	q := strings.ToLower(m.flowhubQuery)
	var filtered []flowhub.Workflow
	for _, w := range m.flowhubResults {
		if q == "" || strings.Contains(strings.ToLower(w.Name), q) ||
			strings.Contains(strings.ToLower(w.Description), q) {
			filtered = append(filtered, w)
		}
	}
	return filtered
}
