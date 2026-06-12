package canonical

import (
	"sort"
	"time"
)

// MergeStrategy defines how conflicts are resolved during bidirectional sync.
type MergeStrategy string

const (
	// MergeNewer takes the version with the most recent modification time for each field.
	MergeNewer MergeStrategy = "newer"

	// MergeUnion combines both sides (union of lists, newer-wins for scalars).
	MergeUnion MergeStrategy = "merge"

	// MergePreferSource prefers the source side on all conflicts.
	MergePreferSource MergeStrategy = "source"
)

// MergeResult holds the merged config and any conflicts that were found.
type MergeResult struct {
	Config    *ShifterConfig
	Conflicts []MergeConflict
}

// MergeConflict describes a conflict between two config versions.
type MergeConflict struct {
	Field   string
	SideA   string // human-readable description of side A value
	SideB   string
	Resolved string // which side won: "a", "b", "merged", "dropped"
}

// MergeConfigs merges two ShifterConfigs using the given strategy.
// SideA is the "ours" side, SideB is "theirs".
func MergeConfigs(a, b *ShifterConfig, strategy MergeStrategy) *MergeResult {
	result := &MergeResult{
		Config: &ShifterConfig{
			Meta: ConfigMeta{
				SourceAdapter: "merged",
				GeneratedAt:   time.Now(),
				Version:       "1.0.0",
			},
		},
	}

	// Merge metadata
	result.Config.Meta.SourcePaths = mergeStringSlices(a.Meta.SourcePaths, b.Meta.SourcePaths)

	// Merge instructions
	result.Config.Instructions = mergeInstructions(a.Instructions, b.Instructions, strategy)

	// Merge agents — union by name, newer wins on conflict
	result.Config.Agents = mergeAgents(a.Agents, b.Agents, strategy, &result.Conflicts)

	// Merge skills — union by name
	result.Config.Skills = mergeSkills(a.Skills, b.Skills, strategy, &result.Conflicts)

	// Merge commands — union by name
	result.Config.Commands = mergeCommands(a.Commands, b.Commands, strategy, &result.Conflicts)

	// Merge MCP servers — union by name, newer wins
	result.Config.MCPServers = mergeMCPServers(a.MCPServers, b.MCPServers, strategy, &result.Conflicts)

	// Merge permissions — union of rules
	result.Config.Permissions = mergePermissions(a.Permissions, b.Permissions, strategy)

	// Merge hooks — union
	result.Config.Hooks = mergeHooks(a.Hooks, b.Hooks, strategy, &result.Conflicts)

	// Merge settings — scalar field newer-wins
	result.Config.Settings = mergeSettings(a.Settings, b.Settings, strategy, &result.Conflicts)

	// Combine loss warnings
	result.Config.LossWarnings = append(a.LossWarnings, b.LossWarnings...)

	return result
}

func mergeStringSlices(a, b []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range a {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range b {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	sort.Strings(result)
	return result
}

func mergeInstructions(a, b []Instruction, strategy MergeStrategy) []Instruction {
	merged := make(map[string]Instruction)
	for _, inst := range a {
		merged[inst.Path] = inst
	}
	for _, inst := range b {
		if existing, ok := merged[inst.Path]; ok {
			switch strategy {
			case MergeNewer, MergePreferSource:
				// Keep existing (a)
				_ = existing
			case MergeUnion:
				// Concatenate content
				inst.Content = existing.Content + "\n\n---\n\n" + inst.Content
				merged[inst.Path] = inst
			}
		} else {
			merged[inst.Path] = inst
		}
	}
	result := make([]Instruction, 0, len(merged))
	for _, inst := range merged {
		result = append(result, inst)
	}
	return result
}

func mergeAgents(a, b []AgentDef, strategy MergeStrategy, conflicts *[]MergeConflict) []AgentDef {
	merged := make(map[string]AgentDef)
	for _, agent := range a {
		merged[agent.Name] = agent
	}
	for _, agent := range b {
		if _, ok := merged[agent.Name]; ok {
			if strategy == MergeNewer {
				// Side B is newer
				merged[agent.Name] = agent
			}
			// For source strategy, keep side A (already in map)
			if strategy == MergeUnion {
				*conflicts = append(*conflicts, MergeConflict{
					Field:   "agents." + agent.Name,
					SideA:   merged[agent.Name].Description,
					SideB:   agent.Description,
					Resolved: "a",
				})
			}
		} else {
			merged[agent.Name] = agent
		}
	}
	result := make([]AgentDef, 0, len(merged))
	for _, agent := range merged {
		result = append(result, agent)
	}
	return result
}

func mergeSkills(a, b []SkillDef, strategy MergeStrategy, conflicts *[]MergeConflict) []SkillDef {
	merged := make(map[string]SkillDef)
	for _, skill := range a {
		merged[skill.Name] = skill
	}
	for _, skill := range b {
		if _, ok := merged[skill.Name]; ok {
			if strategy == MergeNewer {
				merged[skill.Name] = skill
			}
		} else {
			merged[skill.Name] = skill
		}
	}
	result := make([]SkillDef, 0, len(merged))
	for _, skill := range merged {
		result = append(result, skill)
	}
	return result
}

func mergeCommands(a, b []CommandDef, strategy MergeStrategy, conflicts *[]MergeConflict) []CommandDef {
	merged := make(map[string]CommandDef)
	for _, cmd := range a {
		merged[cmd.Name] = cmd
	}
	for _, cmd := range b {
		if _, ok := merged[cmd.Name]; ok {
			if strategy == MergeNewer {
				merged[cmd.Name] = cmd
			}
		} else {
			merged[cmd.Name] = cmd
		}
	}
	result := make([]CommandDef, 0, len(merged))
	for _, cmd := range merged {
		result = append(result, cmd)
	}
	return result
}

func mergeMCPServers(a, b []MCPServerDef, strategy MergeStrategy, conflicts *[]MergeConflict) []MCPServerDef {
	merged := make(map[string]MCPServerDef)
	for _, mcp := range a {
		merged[mcp.Name] = mcp
	}
	for _, mcp := range b {
		if _, ok := merged[mcp.Name]; ok {
			if strategy == MergeNewer {
				merged[mcp.Name] = mcp
			}
		} else {
			merged[mcp.Name] = mcp
		}
	}
	result := make([]MCPServerDef, 0, len(merged))
	for _, mcp := range merged {
		result = append(result, mcp)
	}
	return result
}

func mergePermissions(a, b *PermissionSet, strategy MergeStrategy) *PermissionSet {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}

	merged := &PermissionSet{
		DefaultMode: a.DefaultMode,
	}
	if strategy == MergeNewer && b.DefaultMode != "" {
		merged.DefaultMode = b.DefaultMode
	}

	// Union allow rules, deduplicate by pattern
	allowSeen := make(map[string]bool)
	for _, r := range a.AllowRules {
		if !allowSeen[r.Pattern] {
			merged.AllowRules = append(merged.AllowRules, r)
			allowSeen[r.Pattern] = true
		}
	}
	for _, r := range b.AllowRules {
		if !allowSeen[r.Pattern] {
			merged.AllowRules = append(merged.AllowRules, r)
			allowSeen[r.Pattern] = true
		}
	}

	denySeen := make(map[string]bool)
	for _, r := range a.DenyRules {
		if !denySeen[r.Pattern] {
			merged.DenyRules = append(merged.DenyRules, r)
			denySeen[r.Pattern] = true
		}
	}
	for _, r := range b.DenyRules {
		if !denySeen[r.Pattern] {
			merged.DenyRules = append(merged.DenyRules, r)
			denySeen[r.Pattern] = true
		}
	}

	return merged
}

func mergeHooks(a, b []HookDef, strategy MergeStrategy, conflicts *[]MergeConflict) []HookDef {
	seen := make(map[string]bool)
	var merged []HookDef

	key := func(h HookDef) string {
		return h.Event + "|" + h.Matcher + "|" + h.Command
	}

	for _, h := range a {
		if !seen[key(h)] {
			merged = append(merged, h)
			seen[key(h)] = true
		}
	}
	for _, h := range b {
		if !seen[key(h)] {
			merged = append(merged, h)
			seen[key(h)] = true
		}
	}
	return merged
}

func mergeSettings(a, b SettingsMap, strategy MergeStrategy, conflicts *[]MergeConflict) SettingsMap {
	result := a // start with side A

	if strategy == MergeNewer {
		if b.Model != "" {
			result.Model = b.Model
		}
		if b.SmallModel != "" {
			result.SmallModel = b.SmallModel
		}
		if b.ReasoningEffort != "" {
			result.ReasoningEffort = b.ReasoningEffort
		}
		if b.ApprovalMode != "" {
			result.ApprovalMode = b.ApprovalMode
		}
		if b.SandboxMode != "" {
			result.SandboxMode = b.SandboxMode
		}
	}

	// Union Extra fields
	if result.Extra == nil {
		result.Extra = make(map[string]interface{})
	}
	for k, v := range b.Extra {
		result.Extra[k] = v
	}

	return result
}
