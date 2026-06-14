# 🔄 Shifter

[📖 中文文档](README.zh-CN.md) | [📋 Manual](MANUAL.md)

**One-click config porting between coding agents.** Configure once, use everywhere.

```bash
shifter port claude-code --to codex       # one command
shifter ui                                # interactive wizard
```

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/Moximxxx/shifter/main/install.sh | sh
```

## Supported Agents

| Agent | Agents | Skills | Commands | MCP | Permissions | Hooks |
|-------|:---:|:---:|:---:|:---:|:---:|:---:|
| **Claude Code** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Codex CLI** | ✓ | ✓ | embed | ✓ | ✓ | ✓ |
| **OpenCode** | ✓ | ✓ | ✓ | ✓ | ✓ | — |
| **Qoder** | ✓ | ✓ | — | ✓ | — | — |
| **Gemini CLI** | — | — | — | ✓ | ✓ | — |
| **Cline** | embed | — | — | — | — | — |
| **Aider** | — | — | — | — | ✓ | — |

> `—` = not natively supported; `embed` = preserved as embedded instructions

## Quick Start

```bash
shifter detect                            # scan for configured agents
shifter port claude-code --to codex       # port config
shifter port --dry-run                    # preview first (with SHIFTER_SOURCE/TARGET set)
shifter ui                                # interactive TUI
```

### Environment defaults (optional)

```bash
eval "$(shifter env init)"
export SHIFTER_SOURCE=claude-code
export SHIFTER_TARGET=codex
shifter port          # no args needed
```

## Commands

| Command | Purpose |
|---------|---------|
| `shifter port <src> --to <tgt>` | One-directional config transfer |
| `shifter export <agent>` | Export to canonical JSON (private format) |
| `shifter import <file> --to <agent>` | Import from canonical JSON |
| `shifter sync <a> <b>` | Bidirectional sync |
| `shifter save <name> --source <agent>` | Save config as reusable profile |
| `shifter load <name> --to <agent>` | Apply saved profile |
| `shifter detect` | Scan system for configured agents |
| `shifter status` | Show project agent status |
| `shifter backup <agent>` | Create/restore timestamped backups |
| `shifter env` | Show environment configuration |
| `shifter ui` / `shifter wizard` | Interactive terminal wizard |

## How It Works

```
Source Agent ──Read──▶ Canonical JSON ──Write──▶ Target Agent
(Claude Code)          (Private Format)          (Codex/OpenCode/Qoder)
```

- **Semantic mapping**: understands config meaning, not just format conversion
- **Loss tracking**: features that can't perfectly port generate clear warnings
- **Auto-backup**: every write creates a timestamped backup — undo is always available
- **Safe writes**: validates syntax before touching disk

## Project Structure

```
shifter/
├── adapter/          # Agent adapters (one per coding agent)
│   ├── claudecode/   # Claude Code (JSON + Markdown frontmatter)
│   ├── codex/        # Codex CLI (TOML)
│   ├── opencode/     # OpenCode (JSONC)
│   ├── qoder/        # Qoder (Markdown frontmatter)
│   ├── gemini/       # Gemini CLI (JSON)
│   ├── cline/        # Cline (Markdown rules)
│   └── aider/        # Aider (YAML)
├── canonical/        # Universal config model + merge logic
├── engine/           # Business logic (port, sync, detect, backup, profile)
├── registry/         # Adapter factory for adding new agents
├── tui/              # Bubble Tea interactive terminal UI
├── cmd/              # CLI commands (cobra)
├── pkg/              # Utilities (env, format, paths)
├── test/             # Real-world config fixtures for all 7 agents
├── dist/             # Release binaries
└── install.sh        # One-liner installer
```

## License

MIT
