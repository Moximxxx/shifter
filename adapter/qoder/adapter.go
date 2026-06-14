package qoder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moximxxx/shifter/adapter"
	"github.com/moximxxx/shifter/canonical"
	"github.com/moximxxx/shifter/pkg/format"
	"github.com/moximxxx/shifter/pkg/paths"
	"gopkg.in/yaml.v3"
)

type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) ID() string   { return "qoder" }
func (a *Adapter) Name() string { return "Qoder" }

func (a *Adapter) SearchPaths() []string {
	home := paths.MustHomeDir()
	return []string{
		filepath.Join(home, ".qoder"),
		filepath.Join(home, ".qoder-cn"),
		".qoder",
	}
}

func (a *Adapter) Capabilities() canonical.CapabilityMatrix {
	return canonical.CapabilityMatrix{
		SupportsAgents: true,
		SupportsSkills: true,
		SupportsMCP:    true,
		Notes: map[string]string{
			"instructions": "Qoder has no project instruction file equivalent to CLAUDE.md",
			"commands":     "Qoder has no slash command system",
			"permissions":  "Qoder controls permissions via agent tool whitelists",
			"hooks":        "Qoder has no hook system",
			"mcp":          "MCP servers stored in .qoder/mcp.json",
		},
	}
}

func (a *Adapter) Detect() (adapter.DetectionResult, error) {
	result := adapter.DetectionResult{
		Summary: make(map[string]int),
	}
	home := paths.MustHomeDir()

	for _, dir := range []string{".qoder", filepath.Join(home, ".qoder"), filepath.Join(home, ".qoder-cn")} {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			result.Found = true
			if filepath.IsAbs(dir) {
				result.GlobalPaths = append(result.GlobalPaths, dir)
			} else {
				result.ProjectPath = dir
			}

			// Count agents
			agentsDir := filepath.Join(dir, "agents")
			if entries, err := os.ReadDir(agentsDir); err == nil {
				count := 0
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), ".md") {
						count++
					}
				}
				if count > 0 {
					result.Summary["agents"] = count
				}
			}

			// Count skills
			skillsDir := filepath.Join(dir, "skills")
			if info, err := os.Stat(skillsDir); err == nil && info.IsDir() {
				entries, _ := os.ReadDir(skillsDir)
				count := 0
				for _, e := range entries {
					if e.IsDir() {
						count++
					}
				}
				if count > 0 {
					result.Summary["skills"] = count
				}
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

	var qoderDir string
	home := paths.MustHomeDir()

	for _, dir := range []string{
		filepath.Join(opts.ProjectRoot, ".qoder"),
		filepath.Join(home, ".qoder"),
		filepath.Join(home, ".qoder-cn"),
	} {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			qoderDir = dir
			break
		}
	}

	if qoderDir == "" {
		return cfg, nil
	}
	cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, qoderDir)

	// Read agents
	agentsDir := filepath.Join(qoderDir, "agents")
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, entry := range entries {
			if !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			agentPath := filepath.Join(agentsDir, entry.Name())
			agent, err := readQoderAgentFile(agentPath)
			if err == nil {
				cfg.Agents = append(cfg.Agents, agent)
				cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, agentPath)
			}
		}
	}

	// Read skills
	skillsDir := filepath.Join(qoderDir, "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillPath := filepath.Join(skillsDir, entry.Name())
			skill, err := readQoderSkillDir(skillPath)
			if err == nil {
				cfg.Skills = append(cfg.Skills, skill)
				cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, skillPath)
			}
		}
	}

	// Read MCP servers from .qoder/mcp.json
	mcpPath := filepath.Join(qoderDir, "mcp.json")
	if data, err := os.ReadFile(mcpPath); err == nil {
		var mcpConfig map[string]struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env"`
		}
		if json.Unmarshal(data, &mcpConfig) == nil {
			for name, srv := range mcpConfig {
				cfg.MCPServers = append(cfg.MCPServers, canonical.MCPServerDef{
					Name:    name,
					Type:    "stdio",
					Command: srv.Command,
					Args:    srv.Args,
					Env:     srv.Env,
					Enabled: true,
				})
			}
			cfg.Meta.SourcePaths = append(cfg.Meta.SourcePaths, mcpPath)
		}
	}

	return cfg, nil
}

func (a *Adapter) Write(ctx context.Context, cfg *canonical.ShifterConfig, opts adapter.WriteOptions) (*adapter.WriteResult, error) {
	result := &adapter.WriteResult{}

	var qoderDir string
	switch opts.Scope {
	case "global":
		home := paths.MustHomeDir()
		qoderDir = filepath.Join(home, ".qoder")
	default:
		qoderDir = filepath.Join(opts.ProjectRoot, ".qoder")
	}

	// Write agents
	if len(cfg.Agents) > 0 {
		agentsDir := filepath.Join(qoderDir, "agents")
		if !opts.DryRun {
			os.MkdirAll(agentsDir, 0755)
		}
		for _, agent := range cfg.Agents {
			agentPath := filepath.Join(agentsDir, agent.Name+".md")
			content := formatQoderAgentFile(agent)
			if !opts.DryRun {
				os.WriteFile(agentPath, []byte(content), 0644)
			}
			result.FilesWritten = append(result.FilesWritten, agentPath)
		}
	}

	// Write skills
	if len(cfg.Skills) > 0 {
		skillsDir := filepath.Join(qoderDir, "skills")
		if !opts.DryRun {
			os.MkdirAll(skillsDir, 0755)
		}
		for _, skill := range cfg.Skills {
			skillDir := filepath.Join(skillsDir, skill.Name)
			skillPath := filepath.Join(skillDir, "SKILL.md")
			content := formatQoderSkillFile(skill)
			if !opts.DryRun {
				os.MkdirAll(skillDir, 0755)
				os.WriteFile(skillPath, []byte(content), 0644)
				for name, content := range skill.Scripts {
					os.MkdirAll(filepath.Join(skillDir, "scripts"), 0755)
					os.WriteFile(filepath.Join(skillDir, "scripts", name), []byte(content), 0755)
				}
				for name, content := range skill.References {
					os.WriteFile(filepath.Join(skillDir, name), []byte(content), 0644)
				}
			}
			result.FilesWritten = append(result.FilesWritten, skillPath)
		}
	}

	// Write MCP servers to .qoder/mcp.json
	if len(cfg.MCPServers) > 0 {
		mcpConfig := make(map[string]struct {
			Command string            `json:"command"`
			Args    []string          `json:"args"`
			Env     map[string]string `json:"env,omitempty"`
		})
		for _, mcp := range cfg.MCPServers {
			mcpConfig[mcp.Name] = struct {
				Command string            `json:"command"`
				Args    []string          `json:"args"`
				Env     map[string]string `json:"env,omitempty"`
			}{Command: mcp.Command, Args: mcp.Args, Env: mcp.Env}
		}
		mcpPath := filepath.Join(qoderDir, "mcp.json")
		if !opts.DryRun {
			data, _ := json.MarshalIndent(mcpConfig, "", "  ")
			os.WriteFile(mcpPath, data, 0644)
		}
		result.FilesWritten = append(result.FilesWritten, mcpPath)
	}

	// Loss warnings
	if len(cfg.Commands) > 0 {
		result.LossWarnings = append(result.LossWarnings, canonical.LossWarning{
			Feature:     "commands",
			SourceAgent: cfg.Meta.SourceAdapter,
			TargetAgent: a.ID(),
			Reason:      "Qoder has no slash command system",
			Severity:    "info",
		})
	}
	if len(cfg.Hooks) > 0 {
		result.LossWarnings = append(result.LossWarnings, canonical.LossWarning{
			Feature:     "hooks",
			SourceAgent: cfg.Meta.SourceAdapter,
			TargetAgent: a.ID(),
			Reason:      "Qoder has no hook system",
			Severity:    "warning",
		})
	}

	return result, nil
}

func (a *Adapter) Preview(ctx context.Context, cfg *canonical.ShifterConfig) (*adapter.DiffResult, error) {
	return &adapter.DiffResult{}, nil
}

func readQoderAgentFile(path string) (canonical.AgentDef, error) {
	agent := canonical.AgentDef{}
	content, err := os.ReadFile(path)
	if err != nil {
		return agent, err
	}

	fm, body, err := format.ParseFrontMatter(string(content))
	if err != nil {
		return agent, err
	}

	agent.SystemPrompt = strings.TrimSpace(body)
	if name, ok := fm["name"].(string); ok {
		agent.Name = name
	} else {
		agent.Name = strings.TrimSuffix(filepath.Base(path), ".md")
	}
	if desc, ok := fm["description"].(string); ok {
		agent.Description = desc
	}
	if tools, ok := fm["tools"].(string); ok {
		for _, t := range strings.Split(tools, ",") {
			agent.Tools = append(agent.Tools, strings.TrimSpace(t))
		}
	}
	if model, ok := fm["model"].(string); ok {
		agent.Model = model
	}
	if skills, ok := fm["skills"].([]interface{}); ok {
		for _, s := range skills {
			if sm, ok := s.(map[string]interface{}); ok {
				if sn, ok := sm["skillName"].(string); ok {
					agent.Skills = append(agent.Skills, sn)
				}
			}
		}
	}
	if mcp, ok := fm["mcpServers"].([]interface{}); ok {
		for _, m := range mcp {
			if ms, ok := m.(string); ok {
				agent.MCPServers = append(agent.MCPServers, ms)
			}
		}
	}

	return agent, nil
}

func readQoderSkillDir(dir string) (canonical.SkillDef, error) {
	skill := canonical.SkillDef{
		Scripts:    make(map[string]string),
		References: make(map[string]string),
	}

	skillMDPath := filepath.Join(dir, "SKILL.md")
	content, err := os.ReadFile(skillMDPath)
	if err != nil {
		return skill, err
	}

	fm, body, err := format.ParseFrontMatter(string(content))
	if err != nil {
		return skill, err
	}

	skill.Markdown = strings.TrimSpace(body)
	if name, ok := fm["name"].(string); ok {
		skill.Name = name
	} else {
		skill.Name = filepath.Base(dir)
	}
	if desc, ok := fm["description"].(string); ok {
		skill.Description = desc
	}

	// Read supporting files
	for _, name := range []string{"REFERENCE.md", "EXAMPLES.md"} {
		refPath := filepath.Join(dir, name)
		if data, err := os.ReadFile(refPath); err == nil {
			skill.References[name] = string(data)
		}
	}

	return skill, nil
}

func formatQoderAgentFile(agent canonical.AgentDef) string {
	fm := make(map[string]interface{})
	fm["name"] = agent.Name
	if agent.Description != "" {
		fm["description"] = agent.Description
	}
	if len(agent.Tools) > 0 {
		fm["tools"] = strings.Join(agent.Tools, ", ")
	}
	if agent.Model != "" {
		fm["model"] = agent.Model
	}
	if len(agent.Skills) > 0 {
		fm["skills"] = agent.Skills
	}
	if len(agent.MCPServers) > 0 {
		fm["mcpServers"] = agent.MCPServers
	}

	fmBytes, _ := yaml.Marshal(fm)
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(string(fmBytes))
	b.WriteString("---\n\n")
	b.WriteString(agent.SystemPrompt)
	if !strings.HasSuffix(agent.SystemPrompt, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

func formatQoderSkillFile(skill canonical.SkillDef) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("name: %s\n", skill.Name))
	if skill.Description != "" {
		b.WriteString(fmt.Sprintf("description: %s\n", skill.Description))
	}
	b.WriteString("---\n\n")
	b.WriteString(skill.Markdown)
	if !strings.HasSuffix(skill.Markdown, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}
