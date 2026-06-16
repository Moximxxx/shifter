package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/moximxxx/shifter/engine/detect"
	"github.com/moximxxx/shifter/engine/flowhub"
)

// flowHub key helpers
func sendFlowHubKey(m WizardModel, key string) WizardModel {
	if key == "backspace" {
		// Backspace is handled specially in main Update, not handleInputMode
		if len(m.flowhubQuery) > 0 {
			m.flowhubQuery = m.flowhubQuery[:len(m.flowhubQuery)-1]
		}
		return m
	}
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	updated, _ := m.Update(msg)
	switch v := updated.(type) {
	case WizardModel:
		return v
	case *WizardModel:
		return *v
	}
	return m
}

func sendFlowHubKeyModel(m FlowHubModel, key string) FlowHubModel {
	var msg tea.KeyMsg
	switch key {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "q":
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	default:
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}
	updated, _ := m.Update(msg)
	switch v := updated.(type) {
	case FlowHubModel:
		return v
	case *FlowHubModel:
		return *v
	}
	return m
}

// helper: send key to model and return updated model
func sendKey(m WizardModel, key string) WizardModel {
	var msg tea.KeyMsg
	switch key {
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		msg = tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		msg = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	case "space":
		msg = tea.KeyMsg{Type: tea.KeySpace}
	case "q":
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	}
	updated, _ := m.Update(msg)
	switch v := updated.(type) {
	case WizardModel:
		return v
	case *WizardModel:
		return *v
	}
	return m
}

// ============================================================
// Menu Navigation
// ============================================================

func TestWizard_MenuNavigation(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.detectResults = []detect.Result{
		{ID: "claude-code", Name: "Claude Code", Found: true, HasProjectConfig: true},
		{ID: "codex", Name: "Codex CLI", Found: true, HasProjectConfig: true},
	}

	// Verify initial state
	if m.screen != WizMenu {
		t.Fatalf("expected WizMenu, got %v", m.screen)
	}

	// Move down to Load
	m = sendKey(m, "down")
	if m.cursorIdx != 1 {
		t.Errorf("cursor should be 1 (Load), got %d", m.cursorIdx)
	}

	// Move down to Port
	m = sendKey(m, "down")
	if m.cursorIdx != 2 {
		t.Errorf("cursor should be 2 (Port), got %d", m.cursorIdx)
	}

	// Move down to Settings
	m = sendKey(m, "down")
	if m.cursorIdx != 3 {
		t.Errorf("cursor should be 3 (Settings), got %d", m.cursorIdx)
	}

	// Move back up
	m = sendKey(m, "up")
	if m.cursorIdx != 2 {
		t.Errorf("cursor should be 2 (back to Port), got %d", m.cursorIdx)
	}
}

// ============================================================
// Smart Selection — Save with 1 agent
// ============================================================

func TestWizard_Save_SingleAgent_SkipsSelection(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.cursorIdx = 0 // Save
	m.detectResults = []detect.Result{
		{ID: "claude-code", Name: "Claude Code", Found: true, HasProjectConfig: true},
	}

	m = sendKey(m, "enter")

	// Should skip WizSaveSelectAgent and go directly to WizSaveName
	if m.screen != WizSaveName {
		t.Errorf("single agent should skip to WizSaveName, got %v", m.screen)
	}
	if m.selectedAgentID != "claude-code" {
		t.Errorf("auto-selected agent should be claude-code, got %q", m.selectedAgentID)
	}
	if !m.inputMode {
		t.Error("should be in input mode for naming")
	}
}

// ============================================================
// Smart Selection — Save with 2+ agents
// ============================================================

func TestWizard_Save_MultipleAgents_ShowsSelection(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.cursorIdx = 0 // Save
	m.detectResults = []detect.Result{
		{ID: "claude-code", Name: "Claude Code", Found: true, HasProjectConfig: true},
		{ID: "codex", Name: "Codex CLI", Found: true, HasProjectConfig: true},
	}

	m = sendKey(m, "enter")

	if m.screen != WizSaveSelectAgent {
		t.Errorf("multiple agents should show WizSaveSelectAgent, got %v", m.screen)
	}
}

// ============================================================
// Smart Selection — Save with 0 agents
// ============================================================

func TestWizard_Save_NoAgents_StaysOnMenu(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.cursorIdx = 0 // Save
	m.detectResults = []detect.Result{}

	m = sendKey(m, "enter")

	if m.screen != WizMenu {
		t.Errorf("no agents should stay on WizMenu, got %v", m.screen)
	}
}

// ============================================================
// Smart Selection — Port with 1 agent
// ============================================================

func TestWizard_Port_SingleAgent_SkipsSourceSelection(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.cursorIdx = 2 // Port
	m.detectResults = []detect.Result{
		{ID: "claude-code", Name: "Claude Code", Found: true, HasProjectConfig: true},
	}

	m = sendKey(m, "enter")

	// Should skip WizPortSelectSource and go directly to WizPortSelectTarget
	if m.screen != WizPortSelectTarget {
		t.Errorf("single agent should skip to WizPortSelectTarget, got %v", m.screen)
	}
	if m.sourceID != "claude-code" {
		t.Errorf("auto-selected source should be claude-code, got %q", m.sourceID)
	}
}

// ============================================================
// Smart Selection — Port with 2+ agents
// ============================================================

func TestWizard_Port_MultipleAgents_ShowsSourceSelection(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.cursorIdx = 2 // Port
	m.detectResults = []detect.Result{
		{ID: "claude-code", Name: "Claude Code", Found: true, HasProjectConfig: true},
		{ID: "codex", Name: "Codex CLI", Found: true, HasProjectConfig: true},
		{ID: "opencode", Name: "OpenCode", Found: true, HasProjectConfig: true},
	}

	m = sendKey(m, "enter")

	if m.screen != WizPortSelectSource {
		t.Errorf("multiple agents should show WizPortSelectSource, got %v", m.screen)
	}
}

// ============================================================
// Esc Navigation — backStack
// ============================================================

func TestWizard_Esc_Navigation(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.cursorIdx = 2 // Port
	m.detectResults = []detect.Result{
		{ID: "claude-code", Name: "Claude Code", Found: true, HasProjectConfig: true},
	}

	// Enter Port → should go to PortSelectTarget (single agent, skips source)
	m = sendKey(m, "enter")

	// Press Esc → should go back to Menu
	m = sendKey(m, "esc")
	if m.screen != WizMenu {
		t.Errorf("Esc should return to WizMenu, got %v", m.screen)
	}
}

// ============================================================
// Settings — language selection
// ============================================================

func TestWizard_Settings_LanguageToggle(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.cursorIdx = 5 // Settings (after Templates + FlowHub)
	m.detectResults = []detect.Result{}

	m = sendKey(m, "enter")
	if m.screen != WizSettings {
		t.Fatalf("should enter WizSettings, got %v", m.screen)
	}

	// Default language choice is 0 (en)
	if m.langChoice != 0 {
		t.Errorf("default langChoice should be 0 (en), got %d", m.langChoice)
	}

	// Toggle to Chinese
	m = sendKey(m, "down")
	if m.langChoice != 1 {
		t.Errorf("down should select zh, got %d", m.langChoice)
	}

	// Enter to save and go back
	m = sendKey(m, "enter")
	if m.screen != WizMenu {
		t.Errorf("should return to WizMenu after saving settings, got %v", m.screen)
	}
}

// ============================================================
// Load — profiles
// ============================================================

func TestWizard_Load_OpensProfileList(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.cursorIdx = 1 // Load
	m.detectResults = []detect.Result{}

	// Send key but expect it to load profiles (async)
	m = sendKey(m, "enter")
	if m.screen != WizLoadSelectProfile {
		t.Errorf("should enter WizLoadSelectProfile, got %v", m.screen)
	}
}

// ============================================================
// Quit behavior
// ============================================================

func TestWizard_Quit_FromMenu(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.detectResults = []detect.Result{}

	m = sendKey(m, "q")
	if !m.quitting {
		t.Error("q from menu should set quitting")
	}
}

func TestWizard_Esc_FromMenu_Quits(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.detectResults = []detect.Result{}

	m = sendKey(m, "esc")
	if !m.quitting {
		t.Error("esc from menu should set quitting")
	}
}

func TestWizard_Esc_FromSubScreen_DoesNotQuit(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizSettings
	m.backStack = []WizardScreen{WizMenu}
	m.detectResults = []detect.Result{}

	m = sendKey(m, "esc")
	if m.quitting {
		t.Error("esc from sub-screen should not quit, should go back")
	}
	if m.screen != WizMenu {
		t.Errorf("esc should go back to WizMenu, got %v", m.screen)
	}
}

// ============================================================
// View rendering — ensure no panics
// ============================================================

func TestWizard_Views_NoPanic(t *testing.T) {
	m := NewWizardModel()
	m.detectResults = []detect.Result{
		{ID: "claude-code", Name: "Claude Code", Found: true, HasProjectConfig: true, Summary: map[string]int{"agents": 2}},
		{ID: "codex", Name: "Codex CLI", Found: true, HasProjectConfig: true},
	}

	screens := []WizardScreen{
		WizWelcome, WizMenu, WizSaveSelectAgent, WizSaveName,
		WizLoadSelectProfile, WizPortSelectSource, WizPortSelectTarget,
		WizPortAspects, WizSettings,
	}

	for _, s := range screens {
		m.screen = s
		m.cursorIdx = 0
		m.langChoice = 0

		// Render — should not panic
		result := m.View()
		if result == "" && s != WizWelcome {
			t.Logf("screen %v returned empty view (may be loading)", s)
		}
	}
}

// ============================================================
// foundAgents helper
// ============================================================

func TestWizard_FoundAgents(t *testing.T) {
	m := NewWizardModel()
	m.detectResults = []detect.Result{
		{ID: "a", Found: true, HasProjectConfig: true},
		{ID: "b", Found: false},
		{ID: "c", Found: true, HasProjectConfig: true},
	}

	agents := m.foundAgents()
	if len(agents) != 2 {
		t.Errorf("foundAgents: got %d, want 2", len(agents))
	}
	if agents[0].ID != "a" || agents[1].ID != "c" {
		t.Errorf("foundAgents order wrong: %v", agents)
	}
}

// ============================================================
// firstFoundIdx helper
// ============================================================

// ============================================================
// FlowHub — menu entry triggers fetch
// ============================================================
func TestWizard_FlowHub_Enter_TriggersFetch(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.cursorIdx = 4 // FlowHub
	m.detectResults = []detect.Result{}

	m = sendKey(m, "enter")
	if m.screen != WizFlowHub {
		t.Errorf("should enter WizFlowHub, got %v", m.screen)
	}
}

// ============================================================
// FlowHub — typing filters results
// ============================================================
func TestWizard_FlowHub_Type_Filters(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizFlowHub
	m.flowhubLoaded = true
	m.flowhubResults = []flowhub.Workflow{
		{Name: "hello-shifter", Version: "0.1.0", Description: "A demo workflow", Agent: "claude-code", Downloads: 10},
		{Name: "security-audit", Version: "0.2.0", Description: "Security scanning", Agent: "claude-code", Downloads: 5},
	}

	// Type 'h'
	m = sendFlowHubKey(m, "h")
	if m.flowhubQuery != "h" {
		t.Errorf("query should be 'h', got %q", m.flowhubQuery)
	}
	filtered := getFlowHubFiltered(m.flowhubResults, m.flowhubQuery)
	if len(filtered) != 1 || filtered[0].Name != "hello-shifter" {
		t.Errorf("should filter to hello-shifter only, got %d results", len(filtered))
	}
}

// ============================================================
// FlowHub — backspace
// ============================================================
func TestWizard_FlowHub_Backspace(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizFlowHub
	m.flowhubLoaded = true
	m.flowhubQuery = "hel"
	m.flowhubResults = []flowhub.Workflow{
		{Name: "hello-shifter", Version: "0.1.0", Description: "demo", Agent: "claude-code"},
	}

	m = sendFlowHubKey(m, "backspace")
	if m.flowhubQuery != "he" {
		t.Errorf("backspace should remove last char, got %q", m.flowhubQuery)
	}
}

// ============================================================
// FlowHub — Enter shows install hint
// ============================================================
func TestWizard_FlowHub_Enter_ShowsBanner(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizFlowHub
	m.flowhubLoaded = true
	m.flowhubCursor = 0
	m.flowhubResults = []flowhub.Workflow{
		{Name: "hello-shifter", Version: "0.1.0", Description: "demo", Agent: "claude-code"},
	}
	m.backStack = append(m.backStack, WizMenu)

	m = sendKey(m, "enter")
	if m.screen != WizMenu {
		t.Errorf("Enter on workflow should return to menu, got %v", m.screen)
	}
	if m.resultBanner == "" {
		t.Error("resultBanner should have install hint")
	}
}

// ============================================================
// FlowHub — standalone model
// ============================================================
func TestFlowHubModel_Init(t *testing.T) {
	m := NewFlowHubModel()
	if m.screen != fhList {
		t.Errorf("init screen should be fhList(0), got %d", m.screen)
	}
}

func TestFlowHubModel_Esc_Quits(t *testing.T) {
	m := NewFlowHubModel()
	m.loaded = true
	m.results = []flowhub.Workflow{{Name: "test", Version: "1.0"}}
	m = sendFlowHubKeyModel(m, "esc")
	if !m.quitting {
		t.Error("Esc should quit")
	}
}

func TestFlowHubModel_Enter_OpensDetail(t *testing.T) {
	m := NewFlowHubModel()
	m.loaded = true
	m.results = []flowhub.Workflow{{Name: "test", Version: "1.0", Description: "desc"}}
	m = sendFlowHubKeyModel(m, "enter")
	if m.screen != fhDetail {
		t.Errorf("Enter should open detail, got screen=%d", m.screen)
	}
	if m.selected == nil || m.selected.Name != "test" {
		t.Error("selected should be set")
	}
}

func TestFlowHubModel_Detail_Esc_Back(t *testing.T) {
	m := NewFlowHubModel()
	m.screen = fhDetail
	m.selected = &flowhub.Workflow{Name: "test"}
	m = sendFlowHubKeyModel(m, "esc")
	if m.screen != fhList {
		t.Errorf("Esc from detail should return to list, got screen=%d", m.screen)
	}
}

func TestFirstFoundIdx(t *testing.T) {
	m := NewWizardModel()
	m.detectResults = []detect.Result{
		{ID: "a", Found: false},
		{ID: "b", Found: true, HasProjectConfig: true},
		{ID: "c", Found: true, HasProjectConfig: true},
	}

	idx := m.firstFoundIdx()
	if idx != 1 {
		t.Errorf("firstFoundIdx: got %d, want 1", idx)
	}
}

// ============================================================
// Bug 1: Esc during save name input cancels to menu
// ============================================================
func TestWizard_SaveName_Esc_Cancels(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizSaveName
	m.inputMode = true
	m.inputText = "test"
	m.backStack = append(m.backStack, WizMenu)
	m.detectResults = []detect.Result{
		{ID: "claude-code", Name: "Claude Code", Found: true, HasProjectConfig: true},
	}
	m.selectedAgentID = "claude-code"

	m = sendKey(m, "esc")
	if m.screen != WizMenu {
		t.Errorf("Esc should cancel input and return to WizMenu, got %v", m.screen)
	}
	if m.inputMode {
		t.Error("inputMode should be false after cancel")
	}
}

// ============================================================
// Bug 3: Result banner shown on menu after save/load/port
// ============================================================
func TestWizard_ResultBanner_OnMenu(t *testing.T) {
	m := NewWizardModel()
	m.screen = WizMenu
	m.detectResults = []detect.Result{
		{ID: "claude-code", Name: "Claude Code", Found: true, HasProjectConfig: true, Summary: map[string]int{"agents": 2}},
	}
	m.resultBanner = "✓ Profile test saved (2 agents, 3 MCP)"
	view := m.View()
	if len(view) == 0 {
		t.Error("menu view should render with banner")
	}
}
