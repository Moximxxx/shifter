package opencode

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/pkg/format"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) ID() string   { return "opencode" }
func (a *Adapter) Name() string { return "OpenCode" }

func (a *Adapter) SearchPaths() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".config", "opencode"),
		".opencode",
		".agents",
	}
}

func (a *Adapter) Capabilities() canonical.CapabilityMatrix {
	return canonical.CapabilityMatrix{
		SupportsInstructions: true,
		SupportsAgents:       true,
		SupportsSkills:       true,
		SupportsCommands:     true,
		SupportsMCP:          true,
		SupportsPermissions:  true,
		SupportsSettings:     true,
		Notes: map[string]string{
			"hooks": "OpenCode has no hook system; use plugins instead",
		},
	}
}

type opencodeConfig struct {
	Model      string                        `json:"model,omitempty"`
	Permission map[string]string             `json:"permission,omitempty"`
	MCP        map[string]opencodeMCPServer  `json:"mcp,omitempty"`
	Provider   map[string]interface{}        `json:"provider,omitempty"`
	Skills     *opencodeSkills               `json:"skills,omitempty"`
	Agent      map[string]opencodeAgent      `json:"agent,omitempty"`
	Command    map[string]opencodeCommand    `json:"command,omitempty"`
}

type opencodeMCPServer struct {
	Type        string            `json:"type"`
	Command     []string          `json:"command,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	URL         string            `json:"url,omitempty"`
}

type opencodeAgent struct {
	Description string            `json:"description,omitempty"`
	Model       string            `json:"model,omitempty"`
	Prompt      string            `json:"prompt,omitempty"`
	Tools       map[string]bool   `json:"tools,omitempty"`
	Mode        string            `json:"mode,omitempty"`
	Color       string            `json:"color,omitempty"`
}

type opencodeSkills struct {
	Paths []string `json:"paths,omitempty"`
}

type opencodeCommand struct {
	Description string `json:"description,omitempty"`
	Template    *struct {
		File string `json:"file,omitempty"`
	} `json:"template,omitempty"`
	UserInput string `json:"userInput,omitempty"`
	Agent     string `json:"agent,omitempty"`
	Model     string `json:"model,omitempty"`
}

func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	result := adapter.DetectionResult{
		Summary: make(map[string]int),
	}
	home, _ := os.UserHomeDir()

	candidates := []string{
		filepath.Join(home, ".config", "opencode", "opencode.json"),
		filepath.Join(".opencode", "opencode.json"),
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			result.Found = true
			if filepath.IsAbs(p) {
				result.GlobalPaths = append(result.GlobalPaths, filepath.Dir(p))
			} else {
				result.ProjectPath = filepath.Dir(p)
			}
			result.Summary["config_file"] = 1
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

	var configPath string
	switch opts.Scope {
	case "global":
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".config", "opencode", "opencode.json")
	default:
		configPath = filepath.Join(opts.ProjectRoot, ".opencode", "opencode.json")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			configPath = filepath.Join(opts.ProjectRoot, "opencode.json")
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, nil
	}

	// Strip comments (basic JSONC support)
	jsonData := format.StripJSONComments(string(data))

	var oc opencodeConfig
	if err := json.Unmarshal([]byte(jsonData), &oc); err != nil {
		return cfg, nil
	}
	cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, configPath)

	// Model
	if oc.Model != "" {
		cfg.Settings.Model = oc.Model
	}

	// Permissions
	if len(oc.Permission) > 0 {
		ps := &canonical.PermissionSet{}
		for tool, mode := range oc.Permission {
			if mode == "allow" {
				ps.AllowRules = append(ps.AllowRules, canonical.PermissionRule{
					Pattern: tool,
					Action:  "allow",
				})
			} else {
				ps.DenyRules = append(ps.DenyRules, canonical.PermissionRule{
					Pattern: tool,
					Action:  "deny",
				})
			}
		}
		cfg.Permissions = ps
	}

	// MCP servers
	for name, srv := range oc.MCP {
		mcpType := srv.Type
		if mcpType == "" {
			mcpType = "stdio"
		}
		mcpCmd := ""
		if len(srv.Command) > 0 {
			mcpCmd = srv.Command[0]
		}
		var mcpArgs []string
		if len(srv.Command) > 1 {
			mcpArgs = srv.Command[1:]
		}
		cfg.MCPServers = append(cfg.MCPServers, canonical.MCPServerDef{
			Name:    name,
			Type:    mcpType,
			Command: mcpCmd,
			Args:    mcpArgs,
			Env:     srv.Environment,
			URL:     srv.URL,
			Enabled: true,
		})
	}

	// Agents
	for name, agent := range oc.Agent {
		tools := []string{}
		for tool, enabled := range agent.Tools {
			if enabled {
				tools = append(tools, tool)
			}
		}
		cfg.Agents = append(cfg.Agents, canonical.AgentDef{
			Name:         name,
			Description:  agent.Description,
			Model:        agent.Model,
			Tools:        tools,
			SystemPrompt: agent.Prompt,
			Mode:         agent.Mode,
			Color:        agent.Color,
		})
	}

	// Commands
	for name, cmd := range oc.Command {
		prompt := ""
		if cmd.Template != nil && cmd.Template.File != "" {
			templatePath := filepath.Join(opts.ProjectRoot, cmd.Template.File)
			if data, err := os.ReadFile(templatePath); err == nil {
				prompt = string(data)
			}
		}
		cfg.Commands = append(cfg.Commands, canonical.CommandDef{
			Name:        name,
			Description: cmd.Description,
			Prompt:      prompt,
		})
	}

	// Also read agents from .agents/ directory
	agentsDir := filepath.Join(opts.ProjectRoot, ".agents")
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			agentPath := filepath.Join(agentsDir, entry.Name())
			data, err := os.ReadFile(agentPath)
			if err != nil {
				continue
			}
			fm, body, err := format.ParseFrontMatter(string(data))
			if err != nil {
				continue
			}
			agent := canonical.AgentDef{
				SystemPrompt: strings.TrimSpace(body),
			}
			if name, ok := fm["name"].(string); ok {
				agent.Name = name
			} else {
				agent.Name = strings.TrimSuffix(entry.Name(), ".md")
			}
			if desc, ok := fm["description"].(string); ok {
				agent.Description = desc
			}
			if tools, ok := fm["tools"].([]interface{}); ok {
				for _, t := range tools {
					if ts, ok := t.(string); ok {
						agent.Tools = append(agent.Tools, ts)
					}
				}
			}
			if model, ok := fm["model"].(string); ok {
				agent.Model = model
			}
			cfg.Agents = append(cfg.Agents, agent)
		}
	}

	return cfg, nil
}

func (a *Adapter) Write(ctx context.Context, cfg *canonical.ShifterConfig, opts adapter.WriteOptions) (*adapter.WriteResult, error) {
	result := &adapter.WriteResult{}

	oc := opencodeConfig{}

	if cfg.Settings.Model != "" {
		oc.Model = cfg.Settings.Model
	}

	// MCP servers
	if len(cfg.MCPServers) > 0 {
		oc.MCP = make(map[string]opencodeMCPServer)
		for _, mcp := range cfg.MCPServers {
			oc.MCP[mcp.Name] = opencodeMCPServer{
				Type:        mcp.Type,
				Command:     append([]string{mcp.Command}, mcp.Args...),
				Environment: mcp.Env,
				URL:         mcp.URL,
			}
		}
	}

	// Agents
	if len(cfg.Agents) > 0 {
		oc.Agent = make(map[string]opencodeAgent)
		for _, agent := range cfg.Agents {
			tools := make(map[string]bool)
			for _, t := range agent.Tools {
				tools[strings.ToLower(t)] = true
			}
			oc.Agent[agent.Name] = opencodeAgent{
				Description: agent.Description,
				Model:       agent.Model,
				Prompt:      agent.SystemPrompt,
				Tools:       tools,
				Mode:        agent.Mode,
				Color:       agent.Color,
			}
		}
	}

	// Permissions
	if cfg.Permissions != nil {
		oc.Permission = make(map[string]string)
		for _, r := range cfg.Permissions.AllowRules {
			oc.Permission[r.Pattern] = "allow"
		}
		for _, r := range cfg.Permissions.DenyRules {
			oc.Permission[r.Pattern] = "ask"
		}
	}

	// Commands
	if len(cfg.Commands) > 0 {
		oc.Command = make(map[string]opencodeCommand)
		for _, cmd := range cfg.Commands {
			oc.Command[cmd.Name] = opencodeCommand{
				Description: cmd.Description,
			}
		}
	}

	// Loss warning: hooks (OpenCode doesn't support hooks)
	if len(cfg.Hooks) > 0 {
		result.LossWarnings = append(result.LossWarnings, canonical.LossWarning{
			Feature:     "hooks",
			SourceAgent: cfg.Meta.SourceAdapter,
			TargetAgent: a.ID(),
			Reason:      "OpenCode has no hook system; use plugins instead",
			Severity:    "warning",
		})
	}

	var configPath string
	switch opts.Scope {
	case "global":
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".config", "opencode", "opencode.json")
	default:
		configPath = filepath.Join(opts.ProjectRoot, ".opencode", "opencode.json")
	}

	if !opts.DryRun {
		os.MkdirAll(filepath.Dir(configPath), 0755)
		data, err := json.MarshalIndent(oc, "", "  ")
		if err != nil {
			return result, err
		}
		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return result, err
		}
	}
	result.FilesWritten = append(result.FilesWritten, configPath)

	// Write commands as markdown templates
	if len(cfg.Commands) > 0 {
		commandsDir := filepath.Join(opts.ProjectRoot, ".opencode", "commands")
		os.MkdirAll(commandsDir, 0755)
		for _, cmd := range cfg.Commands {
			p := filepath.Join(commandsDir, cmd.Name+".md")
			if !opts.DryRun {
				os.WriteFile(p, []byte(cmd.Prompt), 0644)
			}
			result.FilesWritten = append(result.FilesWritten, p)
		}
	}

	return result, nil
}

func (a *Adapter) Preview(ctx context.Context, cfg *canonical.ShifterConfig) (*adapter.DiffResult, error) {
	return &adapter.DiffResult{}, nil
}
