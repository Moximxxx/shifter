# 🔄 Shifter

**One-click config porting between coding agents.**

Shifter moves your custom configurations — agents, skills, MCP servers, permissions, hooks, settings — between different coding agents. Configure once, use everywhere.

```bash
# Port Claude Code config to Codex in one command
shifter port claude-code --to codex

# Or use the interactive TUI
shifter ui
```

## Supported Agents

| Agent | Instructions | Agents | Skills | Commands | MCP | Permissions | Hooks |
|-------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Claude Code** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Codex CLI** | ✓ | ✓ | ✓ | embed | ✓ | ✓ | ✓ |
| **OpenCode** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | — |
| **Gemini CLI** | ✓ | — | — | — | ✓ | ✓ | — |
| **Qoder** | — | ✓ | ✓ | — | ✓ | — | — |
| **Cline** | ✓ | embed | — | — | — | — | — |
| **Aider** | ✓ | — | — | — | — | ✓ | — |

`—` = not supported natively; `embed` = embedded as instructions

## Installation

### Homebrew (macOS/Linux)
```bash
brew install moximxxx/tap/shifter
```

### One-liner
```bash
curl -fsSL https://raw.githubusercontent.com/moximxxx/shifter/main/install.sh | sh
```

### Go
```bash
go install github.com/moximxxx/shifter@latest
```

### Manual download
Download the latest binary from [GitHub Releases](https://github.com/moximxxx/shifter/releases).

## Quick Start

```bash
# See what agents are configured
shifter detect

# Check project status
shifter status

# Port Claude Code → Codex (dry run)
shifter port claude-code --to codex --dry-run

# Port Claude Code → Codex (apply)
shifter port claude-code --to codex

# Port only specific aspects
shifter port claude-code --to codex --aspects agents,mcp

# Bidirectional sync
shifter sync claude-code codex --strategy newer

# Interactive TUI
shifter ui

# Backup agent configs
shifter backup claude-code

# Restore from backup
shifter backup --restore ~/.shifter/backups/shifter-backup-2026-06-12T210000.tar.gz
```

## How It Works

Shifter uses an internal **canonical model** — a universal configuration representation that bridges all supported agents:

```
Source Agent → Read native config → Canonical Model → Write native config → Target Agent
```

- **Semantic mapping**: Understands what each setting means, not just format conversion
- **Loss tracking**: Every feature that can't perfectly port generates a clear warning
- **Auto-backup**: Every write creates a timestamped backup (undo is always available)
- **Safe writes**: Validates syntax before touching disk — never writes broken config

## Features

### 🔍 Smart Detection
Automatically finds all configured coding agents on your system.

### 📋 Aspect Filtering
Port exactly what you want: `--aspects agents,skills,mcp`

### ⚠️ Loss Warnings
Clear feedback on what couldn't be perfectly ported and why.

### 💾 Automatic Backups
Every write creates a timestamped tar.gz backup. `shifter backup --restore` for one-click undo.

### 🔄 Bidirectional Sync
Keep two agents continuously in sync with merge strategies (newer, merge, source).

### 🎨 Interactive TUI
`shifter ui` — step-by-step guided porting with preview.

## Project Structure

```
shifter/
├── adapter/          # Agent adapters (one per coding agent)
│   ├── claudecode/   # Claude Code
│   ├── codex/        # Codex CLI
│   ├── opencode/     # OpenCode
│   ├── gemini/       # Gemini CLI
│   ├── qoder/        # Qoder
│   ├── cline/        # Cline
│   └── aider/        # Aider
├── canonical/        # Universal config model + merge logic
├── engine/           # Business logic
│   ├── port/         # One-directional transfer
│   ├── sync/         # Bidirectional sync
│   ├── detect/       # Agent scanner
│   └── backup/       # Backup/restore
├── registry/         # Adapter registry
├── tui/              # Bubble Tea terminal UI
├── cmd/              # CLI commands (cobra)
└── pkg/format/       # Format utilities (frontmatter, JSONC)
```

## Development

```bash
# Build
make build

# Run tests
make test

# Full CI check
make check
```

## License

MIT — see [LICENSE](LICENSE) file.
