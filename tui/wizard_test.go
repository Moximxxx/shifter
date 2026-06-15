package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/moximxxx/shifter/engine/detect"
)

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
	m.cursorIdx = 3 // Settings
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

func TestWizard_FirstFoundIdx(t *testing.T) {
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
