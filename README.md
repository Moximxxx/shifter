# 🔄 Shifter

[📖 中文文档](README.zh-CN.md)

<p align="center">
  <strong>One-click config porting between coding agent.</strong><br>
  Configure agents, skills, MCP servers, hooks once — use in Claude Code, Codex, OpenCode, Qoder, and more.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/version-0.1.0-7C3AED?style=flat-square" alt="version">
  <img src="https://img.shields.io/badge/go-1.24%2B-00ADD8?style=flat-square&logo=go" alt="go">
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="license">
  <img src="https://img.shields.io/badge/agents-7-10B981?style=flat-square" alt="agents">
</p>

---

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/Moximxxx/shifter/release/install.sh | sh
```

## Quick Start

```bash
shifter detect                              # scan configured agents
shifter port claude-code --to codex         # migrate config
shifter ui                                  # interactive wizard
```

Or with env defaults (no args needed):

```bash
export SHIFTER_SOURCE=claude-code
export SHIFTER_TARGET=codex
shifter port        # that's it
```

## What It Does

```mermaid
flowchart LR
    A[Claude Code<br>.claude/] -->|Read| B[(Shifter Canonical<br>JSON)]
    C[OpenCode<br>opencode.jsonc] -->|Read| B
    D[Qoder<br>.qoder/] -->|Read| B
    E[... 7 agents] -->|Read| B
    B -->|Write| F[Codex CLI<br>.codex/]
    B -->|Write| G[OpenCode<br>opencode.jsonc]
    B -->|Write| H[Qoder<br>.qoder/]
    B -->|Write| I[... any target]
```

Read native config → universal canonical JSON → write to any target agent. Semantic mapping handles format gaps; incompatible features generate clear severity-leveled warnings.

## Features

|     | Feature | Description |
|:---:|---------|-------------|
| 🔀 | **Port** | Transfer config between agents: `shifter port claude-code --to codex --aspects agents,mcp` |
| 💾 | **Profiles** | Save/load reusable config templates: `shifter save team-setup --source claude-code` |
| 📤 | **Export/Import** | Canonical JSON as portable format: `shifter export claude-code \| shifter import - --to codex` |
| 🔄 | **Sync** | Bidirectional merge with strategies: `shifter sync claude-code codex --strategy newer` |
| 🎨 | **TUI Wizard** | Interactive guided workflow: `shifter ui` |
| ⚠️ | **Loss Tracking** | Every incompatible feature generates a clear severity-leveled warning |
| 💾 | **Auto-Backup** | Every write creates a timestamped tar.gz — undo anytime |

## Supported Agents

| Agent | Agents | Skills | MCP | Hooks | Format |
|-------|:---:|:---:|:---:|:---:|--------|
| **Claude Code** | ✓ | ✓ | ✓ | ✓ | JSON + MD |
| **Codex CLI** | ✓ | ✓ | ✓ | ✓ | TOML |
| **OpenCode** | ✓ | ✓ | ✓ | — | JSONC |
| **Qoder** | ✓ | ✓ | ✓ | — | MD frontmatter |
| **Gemini CLI** | — | — | ✓ | — | JSON |
| **Cline** | embed | — | — | — | Markdown |
| **Aider** | — | — | — | — | YAML |

> `embed` = content preserved as embedded instructions. Full capability matrix in the [Manual](MANUAL.md).

## Commands

```bash
shifter port      <src> --to <tgt>    # one-directional transfer
shifter export    <agent>             # export to canonical JSON
shifter import    <file> --to <tgt>   # import from canonical JSON
shifter sync      <a> <b>             # bidirectional sync
shifter save      <name> --source <a> # save config as profile
shifter load      <name> --to <b>     # apply saved profile
shifter detect                        # scan for configured agents
shifter status                        # show project status
shifter backup    <agent>             # create/restore backups
shifter profiles                      # manage saved profiles
shifter env                           # show environment config
shifter ui                            # interactive TUI wizard
```

## Architecture

```mermaid
flowchart TB
    subgraph Parsing[Parsing Center]
        CC[Claude Code<br>.claude/]
        CX[Codex CLI<br>.codex/]
        OC[OpenCode<br>opencode.jsonc]
        QD[Qoder<br>.qoder/]
    end

    subgraph Core[Shifter Core]
        CAN[(Canonical JSON<br>Private Format)]
        REG[Adapter Factory<br>registry/]
        ENG[Business Engines<br>port/sync/detect/backup]
    end

    subgraph Target[Target Agents]
        TCX[Codex CLI]
        TOC[OpenCode]
        TQD[Qoder]
        TCC[Claude Code]
    end

    CC -->|Read| CAN
    CX -->|Read| CAN
    OC -->|Read| CAN
    QD -->|Read| CAN
    CAN --> REG
    REG -->|Write| TCX
    REG -->|Write| TOC
    REG -->|Write| TQD
    REG -->|Write| TCC
    ENG --> CAN
```

Adding a new agent: implement `AgentAdapter` (Read + Write), register in `registry/`, done.

## Environment Variables

```bash
eval "$(shifter env init)"    # generate shell config

SHIFTER_SOURCE=claude-code    # default source agent
SHIFTER_TARGET=codex          # default target agent
SHIFTER_DRY_RUN=1             # always preview first
SHIFTER_ASPECTS=agents,mcp    # default aspect filter
```

[Full environment reference →](MANUAL.md#environment-variables)

## Manual

Full documentation with all commands, environment variables, config file locations, FAQ, and loss severity guide — **[Shifter Manual →](MANUAL.md)**

## Development

```bash
make build          # compile with version info
make test           # run all tests
make check          # fmt + vet + test + build
make release        # cross-compile all platforms
```

---

<p align="center">
  <sub>MIT · <a href="https://github.com/Moximxxx/shifter">GitHub</a> · <a href="https://github.com/Moximxxx/shifter/releases">Releases</a></sub>
</p>
