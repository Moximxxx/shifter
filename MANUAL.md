# 🔄 Shifter — Operation Manual

> Cross-agent config migration — Agents, Skills, MCP, Hooks, Permissions. Configure once, use everywhere.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/Moximxxx/shifter/release/install.sh | sh
```

## Core Concepts

```
Source Agent ──Read──▶ Canonical JSON ──Write──▶ Target Agent
(Claude Code)          (Private Format)          (Codex/OpenCode/Qoder)
```

Shifter uses an internal canonical model — a universal superset that bridges all supported agents:

1. **Read** — parse native config from source agent
2. **Convert** — translate to canonical JSON
3. **Write** — format to target agent's native config
4. **Track** — features that can't perfectly port generate clear loss warnings

### Three Workflows

| Workflow | Command | Purpose |
|----------|---------|---------|
| 🔀 **Port** | `shifter port A --to B` | Transfer config between agents in current project |
| 💾 **Save** | `shifter save <name> --source A` | Save agent config as reusable global profile |
| 📥 **Load** | `shifter load <name> --to B` | Apply saved profile to any agent in any project |

## Quick Start

```bash
# Step 1: see what agents are configured
shifter detect

# Step 2: preview before applying
shifter port claude-code --to codex --dry-run

# Step 3: apply
shifter port claude-code --to codex

# Step 4: save as reusable profile
shifter save my-setup --source claude-code --desc "My dev config"

# Step 5: use in another project
cd ~/projects/other-project
shifter load my-setup --to codex
```

### With Environment Variables

```bash
eval "$(shifter env init)"
export SHIFTER_SOURCE=claude-code
export SHIFTER_TARGET=codex
shifter port          # no args needed
```

## Command Reference

### `shifter port <source> --to <target>`
One-directional config transfer.

```bash
shifter port claude-code --to codex                     # full transfer
shifter port claude-code --to codex --dry-run           # preview mode
shifter port claude-code --to codex --aspects agents,mcp # select aspects
shifter port claude-code --to codex --scope global      # global scope
```

**Aspects**: `instructions`, `agents`, `skills`, `commands`, `mcp`, `permissions`, `hooks`, `settings`

### `shifter export <agent>`
Export agent config to canonical JSON (private format).

```bash
shifter export claude-code -o my-workflow.shifter.json
shifter export claude-code --pretty   # stdout
```

### `shifter import <file> --to <agent>`
Import canonical JSON and apply to target agent.

```bash
shifter import my-workflow.shifter.json --to codex
shifter export claude-code | shifter import - --to opencode  # pipe mode
```

### `shifter sync <agent-a> <agent-b>`
Bidirectional sync with merge strategies.

```bash
shifter sync claude-code codex                        # newer strategy (default)
shifter sync claude-code codex --strategy merge       # union of lists
shifter sync claude-code codex --strategy source      # prefer first agent
shifter sync claude-code codex --dry-run              # preview
```

### `shifter save <name> --source <agent>`
Save current project's agent config as a global profile (stored in `~/.shifter/profiles/`).

```bash
shifter save team-setup --source claude-code --desc "Team standard config"
```

### `shifter load <name> --to <agent>`
Load a saved profile into the current project.

```bash
shifter load team-setup --to codex
shifter load team-setup --to opencode    # same profile, different agent
```

### `shifter profiles`
Manage saved profiles.

```bash
shifter profiles                        # list all
shifter profiles delete old-profile     # remove
```

### `shifter backup <agent>`
Manage config backups. Every write auto-creates a timestamped backup.

```bash
shifter backup claude-code              # create backup
shifter backup --list                   # list all backups
shifter backup --restore <path>         # restore from backup
```

### `shifter detect`
Scan for configured coding agents.

```bash
shifter detect
shifter detect --json                   # machine-readable
```

### `shifter status`
Show project agent configuration status.

```bash
shifter status
shifter status --json
```

### `shifter env`
Display or configure Shifter environment variables.

```bash
shifter env                             # show current config
shifter env init                        # generate shell configuration
```

### `shifter ui` / `shifter wizard`
Launch interactive terminal wizard.

```
🔄 Shifter — Interactive Wizard

What would you like to do?

  ❯ 💾 Save — capture config as reusable profile
    📥 Load — apply saved profile to this project
    🔀 Port — transfer between agents in this project
```

**Keybindings**:

| Key | Action |
|-----|--------|
| `↑` `↓` / `j` `k` | Navigate |
| `Enter` | Confirm |
| `Space` | Toggle checkbox |
| `Esc` | Go back |
| `q` | Quit |

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `SHIFTER_SOURCE` | Default source agent |
| `SHIFTER_TARGET` | Default target agent |
| `SHIFTER_SCOPE` | Default scope: `project` / `global` / `all` |
| `SHIFTER_DRY_RUN` | Set `1` to always preview first |
| `SHIFTER_NO_BACKUP` | Set `1` to skip backups |
| `SHIFTER_FORCE` | Set `1` to skip confirmations |
| `SHIFTER_ASPECTS` | Default aspects: `agents,skills,mcp` |
| `SHIFTER_PROFILE` | Default profile name |
| `SHIFTER_STRATEGY` | Default strategy: `newer` / `merge` / `source` |
| `SHIFTER_PROJECT_ROOT` | Default project directory |

## Supported Agents

| Agent | Format | Agents | Skills | Commands | MCP | Hooks |
|-------|--------|:---:|:---:|:---:|:---:|:---:|
| **Claude Code** | JSON + MD | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Codex CLI** | TOML | ✓ | ✓ | embed | ✓ | ✓ |
| **OpenCode** | JSONC | ✓ | ✓ | ✓ | ✓ | — |
| **Qoder** | MD frontmatter | ✓ | ✓ | — | ✓ | — |
| **Gemini CLI** | JSON | — | — | — | ✓ | — |
| **Cline** | Markdown | embed | — | — | — | — |
| **Aider** | YAML | — | — | — | — | — |

## Config File Locations

### Shifter Storage

| Path | Purpose |
|------|---------|
| `~/.shifter/profiles/` | Saved config profiles (JSON) |
| `~/.shifter/backups/` | Automatic backups (tar.gz) |

### Agent Config Locations

| Agent | Project | Global |
|-------|---------|--------|
| Claude Code | `.claude/` | `~/.claude/` |
| Codex CLI | `.codex/` | `~/.codex/` |
| OpenCode | `opencode.jsonc`, `.opencode/` | `~/.config/opencode/` |
| Qoder | `.qoder/` | `~/.qoder/` |
| Gemini CLI | `.gemini/` | `~/.gemini/` |
| Cline | `.clinerules`, `.clinerules/` | — |
| Aider | `.aider.conf.yml` | `~/.aider.conf.yml` |

## Loss Severity Levels

| Level | Meaning | Example |
|:---:|------|---------|
| `info` | Feature works differently | Slash commands embedded as instructions |
| `warning` | Partial loss | Subagents can't auto-activate |
| `critical` | Complete loss | Hook system incompatible |

## FAQ

### Does porting overwrite existing config?

Default is merge mode. Use `--dry-run` to preview changes first.

### How do I undo a migration?

Every write auto-creates a backup. Restore with:

```bash
shifter backup --list
shifter backup --restore ~/.shifter/backups/shifter-backup-<timestamp>.tar.gz
```

### Can I share profiles across machines?

Profiles are plain JSON files in `~/.shifter/profiles/`. Use git or any sync tool:

```bash
cd ~/.shifter/profiles
git init && git add -A && git commit -m "my profiles"
```

### How do I add support for a new agent?

1. Implement `AgentAdapter` interface in `adapter/<name>/`
2. Register in `registry/registry.go`
3. PR welcome!

## Quick Reference

```bash
shifter detect                                # what agents exist?
shifter port claude-code --to codex           # migrate config
shifter port A --to B --dry-run               # preview first
shifter port A --to B --aspects agents,mcp    # select what to port
shifter save my-config --source claude-code   # save as profile
shifter load my-config --to opencode          # apply profile
shifter profiles                              # list profiles
shifter sync claude-code codex               # bidirectional sync
shifter export claude-code > workflow.json    # export canonical
shifter import workflow.json --to codex       # import canonical
shifter backup --list                         # list backups
shifter backup --restore <file>               # undo
shifter env                                    # show env config
shifter ui                                     # interactive TUI
```
