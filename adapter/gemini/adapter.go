package gemini

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/pkg/convert"
	"github.com/moximxxx/shifter/pkg/paths"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) ID() string   { return "gemini-cli" }
func (a *Adapter) Name() string { return "Gemini CLI" }

func (a *Adapter) SearchPaths() []string {
	home := paths.MustHomeDir()
	return []string{
		filepath.Join(home, ".gemini"),
		".gemini",
	}
}

func (a *Adapter) Capabilities() canonical.CapabilityMatrix {
	return canonical.CapabilityMatrix{
		SupportsInstructions: true,
		SupportsMCP:          true,
		SupportsPermissions:  true,
		SupportsSettings:     true,
		Notes: map[string]string{
			"agents":    "Gemini CLI has no subagent system",
			"skills":    "Gemini CLI has no skill system; use GEMINI.md instead",
			"commands":  "Gemini CLI has no slash command system",
			"hooks":     "Gemini CLI has no hook system",
		},
	}
}

type geminiSettings struct {
	General    *geminiGeneral    `json:"general,omitempty"`
	UI         *geminiUI         `json:"ui,omitempty"`
	Model      *geminiModel      `json:"model,omitempty"`
	Tools      *geminiTools      `json:"tools,omitempty"`
	MCPServers map[string]geminiMCPServer `json:"mcpServers,omitempty"`
	Context    *geminiContext    `json:"context,omitempty"`
}

type geminiGeneral struct {
	VimMode             *bool  `json:"vimMode,omitempty"`
	PreferredEditor     string `json:"preferredEditor,omitempty"`
	DefaultApprovalMode string `json:"defaultApprovalMode,omitempty"`
}

type geminiUI struct {
	Theme     string `json:"theme,omitempty"`
	HideBanner *bool `json:"hideBanner,omitempty"`
}

type geminiModel struct {
	Name              string  `json:"name,omitempty"`
	MaxSessionTurns   int     `json:"maxSessionTurns,omitempty"`
}

type geminiTools struct {
	Sandbox string   `json:"sandbox,omitempty"`
	Allowed []string `json:"allowed,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
	Core    []string `json:"core,omitempty"`
}

type geminiMCPServer struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
	Trust   *bool             `json:"trust,omitempty"`
}

type geminiContext struct {
	FileName           []string `json:"fileName,omitempty"`
	IncludeDirectories []string `json:"includeDirectories,omitempty"`
}

func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	result := adapter.DetectionResult{
		Summary: make(map[string]int),
	}
	home := paths.MustHomeDir()
	candidates := []string{
		filepath.Join(home, ".gemini"),
		".gemini",
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			result.Found = true
			if filepath.IsAbs(dir) {
				result.GlobalPaths = append(result.GlobalPaths, dir)
			} else {
				result.ProjectPath = dir
			}
			settingsPath := filepath.Join(dir, "settings.json")
			if _, err := os.Stat(settingsPath); err == nil {
				result.Summary["settings_file"] = 1
			}
		}
	}
	return result, nil
}

func (a *Adapter) Read(ctx context.Context, opts adapter.ReadOptions) (*canonical.ShifterConfig, error) {
	cfg := &canonical.ShifterConfig{
		Meta: canonical.ConfigMeta{
			SourceAdapter: a.ID(),
			Version:       "1.0.0",
		},
	}

	var geminiDir string
	switch opts.Scope {
	case "global":
		home := paths.MustHomeDir()
		geminiDir = filepath.Join(home, ".gemini")
	default:
		geminiDir = filepath.Join(opts.ProjectRoot, ".gemini")
	}

	// Read settings.json
	settingsPath := filepath.Join(geminiDir, "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		// If not found, check project root too
		settingsPath = filepath.Join(opts.ProjectRoot, ".gemini", "settings.json")
		data, err = os.ReadFile(settingsPath)
		if err != nil {
			return cfg, nil
		}
	}

	var gs geminiSettings
	if err := json.Unmarshal(data, &gs); err != nil {
		return cfg, nil
	}
	cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, settingsPath)

	// Model
	if gs.Model != nil {
		if gs.Model.Name != "" {
			cfg.Settings.Model = gs.Model.Name
		}
		if gs.Model.MaxSessionTurns > 0 {
			cfg.Settings.MaxTurns = gs.Model.MaxSessionTurns
		}
	}

	// General settings
	if gs.General != nil {
		if gs.General.DefaultApprovalMode != "" {
			cfg.Settings.ApprovalMode = gs.General.DefaultApprovalMode
		}
		if gs.General.VimMode != nil && *gs.General.VimMode {
			cfg.Settings.VimMode = true
		}
	}

	// UI
	if gs.UI != nil && gs.UI.Theme != "" {
		cfg.Settings.Theme = gs.UI.Theme
	}

	// MCP servers
	for name, srv := range gs.MCPServers {
		enabled := true
		if srv.Trust != nil && !*srv.Trust {
			enabled = false
		}
		cfg.MCPServers = append(cfg.MCPServers, canonical.MCPServerDef{
			Name:    name,
			Type:    "stdio",
			Command: srv.Command,
			Args:    srv.Args,
			Env:     srv.Env,
			Timeout: srv.Timeout,
			Enabled: enabled,
			Trust:   srv.Trust,
		})
	}

	// Permissions
	if gs.Tools != nil {
		ps := &canonical.PermissionSet{
			DefaultMode: cfg.Settings.ApprovalMode,
		}
		for _, pattern := range gs.Tools.Allowed {
			ps.AllowRules = append(ps.AllowRules, canonical.PermissionRule{
				Pattern: pattern,
				Action:  "allow",
			})
		}
		for _, pattern := range gs.Tools.Exclude {
			ps.DenyRules = append(ps.DenyRules, canonical.PermissionRule{
				Pattern: pattern,
				Action:  "deny",
			})
		}
		cfg.Permissions = ps
	}

	// Read GEMINI.md or CONTEXT.md
	for _, name := range []string{"GEMINI.md", "CONTEXT.md"} {
		mdPath := filepath.Join(opts.ProjectRoot, name)
		if data, err := os.ReadFile(mdPath); err == nil {
			cfg.Instructions = append(cfg.Instructions, canonical.Instruction{
				Path:    name,
				Content: string(data),
				Scope:   "root",
			})
			cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, mdPath)
			break
		}
	}

	return cfg, nil
}

func (a *Adapter) Write(ctx context.Context, cfg *canonical.ShifterConfig, opts adapter.WriteOptions) (*adapter.WriteResult, error) {
	result := &adapter.WriteResult{}

	gs := geminiSettings{
		General: &geminiGeneral{},
		UI:      &geminiUI{},
		Model:   &geminiModel{},
		Tools:   &geminiTools{},
	}

	if cfg.Settings.Model != "" {
		gs.Model.Name = cfg.Settings.Model
	}
	if cfg.Settings.MaxTurns > 0 {
		gs.Model.MaxSessionTurns = cfg.Settings.MaxTurns
	}
	if cfg.Settings.ApprovalMode != "" {
		gs.General.DefaultApprovalMode = cfg.Settings.ApprovalMode
	}
	if cfg.Settings.VimMode {
		b := true
		gs.General.VimMode = &b
	}
	if cfg.Settings.Theme != "" {
		gs.UI.Theme = cfg.Settings.Theme
	}

	// MCP servers
	if len(cfg.MCPServers) > 0 {
		gs.MCPServers = make(map[string]geminiMCPServer)
		for _, mcp := range cfg.MCPServers {
			gs.MCPServers[mcp.Name] = geminiMCPServer{
				Command: mcp.Command,
				Args:    mcp.Args,
				Env:     mcp.Env,
				Timeout: mcp.Timeout,
				Trust:   mcp.Trust,
			}
		}
	}

	// Permissions
	if cfg.Permissions != nil {
		for _, r := range cfg.Permissions.AllowRules {
			gs.Tools.Allowed = append(gs.Tools.Allowed, r.Pattern)
		}
		for _, r := range cfg.Permissions.DenyRules {
			gs.Tools.Exclude = append(gs.Tools.Exclude, r.Pattern)
		}
	}

	// Loss warnings for unsupported features
	if len(cfg.Agents) > 0 {
		result.LossWarnings = append(result.LossWarnings, convert.NewLossWarning("agents", cfg.Meta.SourceAdapter, a.ID(), "Gemini CLI has no subagent system", "warning"))

	}
	if len(cfg.Skills) > 0 {
		result.LossWarnings = append(result.LossWarnings, convert.NewLossWarning("skills", cfg.Meta.SourceAdapter, a.ID(), "Gemini CLI has no skill system", "info"))

	}
	if len(cfg.Commands) > 0 {
		result.LossWarnings = append(result.LossWarnings, convert.NewLossWarning("commands", cfg.Meta.SourceAdapter, a.ID(), "Gemini CLI has no slash command system", "info"))

	}

	var geminiDir string
	switch opts.Scope {
	case "global":
		home := paths.MustHomeDir()
		geminiDir = filepath.Join(home, ".gemini")
	default:
		geminiDir = filepath.Join(opts.ProjectRoot, ".gemini")
	}

	settingsPath := filepath.Join(geminiDir, "settings.json")
	if !opts.DryRun {
		os.MkdirAll(geminiDir, paths.DirPerm)
		data, err := json.MarshalIndent(gs, "", "  ")
		if err != nil {
			return result, err
		}
		if err := os.WriteFile(settingsPath, data, paths.FilePerm); err != nil {
			return result, err
		}
	}
	result.FilesWritten = append(result.FilesWritten, settingsPath)

	// Write GEMINI.md
	for _, inst := range cfg.Instructions {
		if inst.Scope == "root" {
			geminiMDPath := filepath.Join(opts.ProjectRoot, "GEMINI.md")
			if !opts.DryRun {
				os.WriteFile(geminiMDPath, []byte(inst.Content), paths.FilePerm)
			}
			result.FilesWritten = append(result.FilesWritten, geminiMDPath)
			break
		}
	}

	return result, nil
}

