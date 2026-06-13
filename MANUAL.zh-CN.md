# 🔄 Shifter 操作手册

> **一键在不同 Coding Agent 之间迁移配置** — 把你的 agents、skills、MCP、权限配置一次，到处使用。

---

## 目录

1. [安装](#安装)
2. [核心概念](#核心概念)
3. [快速上手](#快速上手)
4. [常用场景](#常用场景)
5. [命令参考](#命令参考)
6. [交互式向导](#交互式向导)
7. [支持的 Agent](#支持的-agent)
8. [配置对照表](#配置对照表)
9. [配置文件存储位置](#配置文件存储位置)
10. [常见问题](#常见问题)

---

## 安装

### 方式一：从源码编译
```bash
cd /home/qin/code/shifter
go build -o shifter .
sudo mv shifter /usr/local/bin/
```

### 方式二：go install
```bash
go install github.com/moximxxx/shifter@latest
```

### 验证安装
```bash
shifter --help
```

---

## 核心概念

```
┌─────────────────┐                  ┌─────────────────┐
│   Claude Code    │                  │   Codex CLI      │
│  .claude/        │                  │  .codex/         │
│  ├── settings.json│    shifter      │  ├── config.toml │
│  ├── agents/     │  ──────────────▶ │  ├── codex.md    │
│  ├── skills/     │    port/sync     │  └── ...         │
│  └── commands/   │                  │                  │
└─────────────────┘                  └─────────────────┘
```

Shifter 在内部使用一个**通用配置模型**（Canonical Model）来桥接不同的 Agent：

1. **读取**：从源 Agent 的原生配置格式读入
2. **转换**：转为通用模型（用户无需关心这个中间层）
3. **写入**：转换为目标 Agent 的原生格式写出
4. **追踪**：不能完美迁移的功能会生成明确的**损失警告**

### 三个核心工作流

| 工作流 | 命令 | 说明 |
|--------|------|------|
| **🔀 迁移** | `shifter port A --to B` | 在当前项目中，把 Agent A 的配置迁移到 Agent B |
| **💾 保存** | `shifter save <名称>` | 把当前项目的 Agent 配置保存为全局模板 |
| **📥 加载** | `shifter load <名称> --to B` | 把全局模板加载到当前项目的某个 Agent 上 |

---

## 快速上手

### 第一步：看看当前项目有哪些 Agent

```bash
cd ~/projects/my-project
shifter detect
```

输出：
```
Scanning for configured coding agents...

  ✗ Aider (not configured)
  ✓ Claude Code  1 settings_file, 3 agents, 5 skills
  ✗ Cline (not configured)
  ✓ Codex CLI  1 config_file
  ✗ Gemini CLI (not configured)
  ✗ OpenCode (not configured)
  ✗ Qoder (not configured)
```

### 第二步：把 Claude Code 的配置迁移到 Codex

```bash
# 先预览会改什么（不会真的写入）
shifter port claude-code --to codex --dry-run

# 确认无误后，执行迁移
shifter port claude-code --to codex
```

输出：
```
✓ Port complete: claude-code → codex

Files written:
  ✓ .codex/config.toml
  ✓ .codex/codex.md

Loss warnings:
  ⚠ [info] skills: Codex skill support is experimental; enable with features.skills=true

Summary: 3 agents, 5 skills, 2 commands, 2 MCP servers, 2 hooks
```

### 第三步：只迁移部分内容

```bash
# 只要 agents 和 MCP 服务器
shifter port claude-code --to codex --aspects agents,mcp

# 可选的 aspect:
#   instructions  - 项目指令文件 (CLAUDE.md, GEMINI.md 等)
#   agents        - 子代理定义
#   skills        - 技能/Skill 定义
#   commands      - 自定义斜杠命令
#   mcp           - MCP 服务器连接
#   permissions   - 权限规则
#   hooks         - 生命周期钩子
#   settings      - 通用设置（模型、沙箱等）
```

---

## 常用场景

### 场景一：把 Claud Code 的工作流一键配到新项目

```bash
# 在老项目中，保存 Claude Code 配置为模板
cd ~/projects/my-service
shifter save backend-dev --source claude-code --desc "后端开发标准配置"

# 在新项目中，加载模板到 Codex
cd ~/projects/new-service
shifter load backend-dev --to codex

# 也可以加载到 OpenCode
shifter load backend-dev --to opencode
```

### 场景二：团队共享配置模板

```bash
# 团队技术负责人导出配置
cd ~/projects/team-standards
shifter save team-setup --source claude-code --desc "团队标准：agents + MCP + hooks"

# 团队成员导入（每个人可以导入到不同的 Agent）
shifter load team-setup --to codex     # 用 Codex 的同学
shifter load team-setup --to opencode  # 用 OpenCode 的同学
shifter load team-setup --to claude-code  # 用 Claude Code 的同学
```

> **提示**：模板保存在 `~/.shifter/profiles/` 目录下，你可以通过 git 来管理和分享这些 JSON 文件。

### 场景三：多工具并行使用，保持配置同步

```bash
# 双向同步 — 保持两个 Agent 配置一致
shifter sync claude-code codex --strategy newer

# 合并策略说明：
#   newer  - 每个字段使用更新时间较新的版本（默认）
#   merge  - 列表取并集，标量取较新的
#   source - 有冲突时以第一个 Agent 为准
```

### 场景四：查看已保存的所有模板

```bash
shifter profiles
```

输出：
```
Saved profiles (3):

  📁 backend-dev
     Source: claude-code
     Contents: 3 agents, 5 skills, 2 MCP, 2 hooks
     Description: 后端开发标准配置
     Updated: 2026-06-13 10:30

  📁 team-setup
     Source: claude-code
     Contents: 3 agents, 4 MCP, 2 hooks
     Description: 团队标准：agents + MCP + hooks
     Updated: 2026-06-12 15:00

  📁 frontend-toolkit
     Source: opencode
     Contents: 2 agents, 3 skills, 1 MCP
     Description: 前端开发配置
     Updated: 2026-06-11 09:00
```

### 场景五：删除不需要的模板

```bash
shifter profiles delete frontend-toolkit
```

### 场景六：安全的配置备份与恢复

```bash
# 每次写入操作会自动备份（默认开启）
shifter port claude-code --to codex  # 自动备份在 ~/.shifter/backups/

# 手动备份某个 Agent 的配置
shifter backup claude-code

# 查看所有备份
shifter backup --list

# 恢复某个备份
shifter backup --restore ~/.shifter/backups/shifter-backup-2026-06-13T103000.tar.gz

# 跳过备份（CI 环境中）
shifter port claude-code --to codex --no-backup
```

---

## 命令参考

### `shifter detect`
扫描并列出当前系统中已配置的 Coding Agent。

```bash
shifter detect                  # 文本格式
shifter detect --json           # JSON 格式（适合脚本）
```

### `shifter status`
显示当前项目的 Agent 配置状态概览。

```bash
shifter status
shifter status --json
shifter status --exit-code      # 如果没有 Agent 配置则退出码为 1
```

### `shifter port <源> --to <目标>`
单向配置迁移。从源 Agent 读取配置，写入目标 Agent。

```bash
shifter port claude-code --to codex                    # 完整迁移
shifter port claude-code --to codex --dry-run          # 预览模式
shifter port claude-code --to codex --aspects agents,mcp  # 只迁移部分
shifter port claude-code --to codex --scope global     # 全局级别配置
shifter port claude-code --to codex --no-backup        # 跳过自动备份
shifter port claude-code --to codex --force            # 跳过确认提示
```

### `shifter sync <Agent-A> <Agent-B>`
双向配置同步。合并两个 Agent 的配置。

```bash
shifter sync claude-code codex                         # 默认：newer 策略
shifter sync claude-code codex --strategy merge        # 合并策略
shifter sync claude-code codex --strategy source        # 以 A 为准
shifter sync claude-code codex --dry-run               # 预览模式
```

### `shifter save <名称> --source <Agent>`
把当前项目的 Agent 配置保存为全局模板。

```bash
shifter save my-config --source claude-code
shifter save team-setup --source codex --desc "团队标准配置"
```

### `shifter load <名称> --to <Agent>`
加载全局模板到当前项目的某个 Agent。

```bash
shifter load my-config --to codex
shifter load team-setup --to opencode
```

### `shifter profiles`
管理全局模板。

```bash
shifter profiles                  # 列出所有模板
shifter profiles delete old-one   # 删除指定模板
```

### `shifter backup`
管理配置备份。

```bash
shifter backup claude-code        # 创建备份
shifter backup --list             # 列出所有备份
shifter backup --restore <路径>   # 恢复备份
```

### `shifter ui` / `shifter wizard`
启动交互式终端向导。

```bash
shifter ui      # 启动向导
shifter wizard  # 同上
```

---

## 交互式向导

运行 `shifter ui` 或 `shifter wizard` 启动交互式界面。

### 向导流程

```
┌──────────────────────────────────────────────────┐
│ 🔄 Shifter — Interactive Wizard                  │
│                                                  │
│ Found 3 configured agent(s):                     │
│   ✓ Claude Code (1 settings_file)               │
│   ✓ Codex CLI (1 config_file)                   │
│   ✓ OpenCode (1 config_file)                    │
│                                                  │
│ What would you like to do?                       │
│                                                  │
│   ❯ 💾 Save — 保存当前配置为全局模板            │
│     📥 Load — 加载模板到当前项目                 │
│     🔀 Port — 在两个 Agent 之间迁移配置         │
│                                                  │
│ ↑↓ navigate • Enter select • q quit              │
└──────────────────────────────────────────────────┘
```

### 操作键位

| 按键 | 功能 |
|------|------|
| `↑` `↓` 或 `j` `k` | 上下移动光标 |
| `Enter` | 确认选择 |
| `Space` | 勾选/取消勾选（选择 aspect 时） |
| `Esc` | 返回上一屏 |
| `q` | 在主菜单或结果页退出 |
| 字母/数字键 | 输入模板名称（保存时） |
| `Backspace` | 删除输入字符 |

### 三个模式详解

#### 💾 Save — 保存模板

1. **选择源 Agent**：从当前项目已配置的 Agent 中选一个
2. **命名模板**：输入一个名字（字母、数字、连字符、下划线）
3. **保存**：按 Enter 确认，配置被保存到 `~/.shifter/profiles/`

#### 📥 Load — 加载模板

1. **选择模板**：从已保存的模板列表中选一个
2. **选择目标 Agent**：选择要应用模板的 Agent
3. **应用**：按 Enter 确认，配置被写入目标 Agent 的项目目录

#### 🔀 Port — 迁移配置

1. **选择源 Agent**：从哪个 Agent 读取配置
2. **选择目标 Agent**：把配置写入哪个 Agent
3. **选择要迁移的内容**：勾选/取消 agents、skills、MCP 等
4. **执行迁移**：按 Enter，配置被写入并显示结果

---

## 支持的 Agent

| Agent | 格式 | 项目指令 | 子代理 | 技能 | 命令 | MCP | 权限 | 钩子 |
|-------|------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Claude Code** | JSON + Markdown | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Codex CLI** | TOML | ✓ | ✓ | ✓ | 嵌入 | ✓ | ✓ | ✓ |
| **OpenCode** | JSON | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | — |
| **Gemini CLI** | JSON | ✓ | — | — | — | ✓ | ✓ | — |
| **Qoder** | Markdown | — | ✓ | ✓ | — | ✓ | — | — |
| **Cline** | Markdown | ✓ | 嵌入 | — | — | — | — | — |
| **Aider** | YAML | ✓ | — | — | — | — | ✓ | — |

> **图例说明**：
> - `✓` = 原生支持
> - `—` = 不支持（迁移时会显示损失警告）
> - `嵌入` = 作为项目指令嵌入（如 Agent 定义写入 Cline 的 `.clinerules/`）

---

## 配置对照表

### 从 Claude Code 迁移到 Codex CLI 时的转换规则

| Claude Code | → | Codex CLI | 损失等级 |
|-------------|---|-----------|:---:|
| `CLAUDE.md` | → | `.codex/codex.md` | 无 |
| `settings.json` → `model` | → | `config.toml` → `model` | 无 |
| `permissions.allow` | → | `approval_policy = "on-request"` | ⚠ 警告 |
| `mcpServers.*` | → | `[mcp_servers.*]` | 无 |
| `agents/*.md` | → | `[agents.*]` TOML 表 | 无 |
| `hooks.PostToolUse` | → | `[[hooks]]` | 无 |
| `skills/*/SKILL.md` | → | `features.skills = true` | ℹ 信息 |
| `commands/*.md` | → | 嵌入 `codex.md` 中 | ℹ 信息 |

### 从 Claude Code 迁移到 OpenCode 时的转换规则

| Claude Code | → | OpenCode | 损失等级 |
|-------------|---|----------|:---:|
| `CLAUDE.md` | → | 嵌入 agent prompt | ⚠ 警告 |
| `mcpServers.*` | → | `mcp.*` (command 格式不同) | 无 |
| `agents/*.md` | → | `agent.*` 对象 | 无 |
| `permissions.allow` | → | `permission.* = "allow"` | 无 |
| `commands/*.md` | → | `.opencode/commands/*.md` | 无 |
| `hooks.*` | → | **丢弃**（OpenCode 用插件系统） | ⚠ 警告 |

### 损失等级说明

| 等级 | 含义 | 示例 |
|:---:|------|------|
| ℹ 信息 | 功能可用但形式不同 | 斜杠命令嵌入为文本指令 |
| ⚠ 警告 | 功能部分丢失 | 子代理无法自动激活 |
| 🔴 严重 | 功能完全丢失 | 钩子系统不兼容 |

---

## 配置文件存储位置

### Shifter 自身

| 路径 | 用途 |
|------|------|
| `~/.shifter/profiles/` | 全局配置模板（JSON 格式） |
| `~/.shifter/backups/` | 自动备份存档（tar.gz 格式） |
| `~/.shifter/sync-state.json` | 同步状态记录（未来支持） |

### 各 Agent 的项目配置

| Agent | 项目级配置 | 全局级配置 |
|-------|-----------|-----------|
| Claude Code | `.claude/` | `~/.claude/` |
| Codex CLI | `.codex/` | `~/.codex/` |
| OpenCode | `.opencode/`, `opencode.json` | `~/.config/opencode/` |
| Gemini CLI | `.gemini/` | `~/.gemini/` |
| Qoder | `.qoder/` | `~/.qoder/`, `~/.qoder-cn/` |
| Cline | `.clinerules`, `.clinerules/` | — |
| Aider | `.aider.conf.yml` | `~/.aider.conf.yml` |

---

## 常见问题

### Q: 迁移会覆盖目标 Agent 的现有配置吗？

默认是**合并模式**。Shifter 会尽量保留目标 Agent 的现有配置，只添加或更新对应字段。你可以先使用 `--dry-run` 预览变更。

```bash
shifter port claude-code --to codex --dry-run
```

### Q: 哪些内容不能被迁移？

每个 Agent 支持的功能不同。不能完美迁移的内容会显示**损失警告**。比如：

- Claude Code 的**钩子**（hooks）无法迁移到 OpenCode（OpenCode 没有钩子系统）
- Claude Code 的**斜杠命令**迁移到 Codex 时会嵌入到项目指令中
- Claude Code 的**子代理**迁移到 Cline/Aider 时会作为文本嵌入规则文件

### Q: 如何恢复到迁移前的状态？

每次写入操作都会自动创建备份。使用以下命令恢复：

```bash
# 查看备份列表
shifter backup --list

# 恢复到最近的备份
shifter backup --restore ~/.shifter/backups/shifter-backup-2026-06-13T103000.tar.gz
```

### Q: 可以在两台机器之间同步模板吗？

模板文件保存在 `~/.shifter/profiles/` 下，是纯 JSON 文件。你可以：

```bash
# 导出
cp -r ~/.shifter/profiles ~/backup/profiles

# 导入到另一台机器
scp -r ~/backup/profiles other-machine:~/.shifter/

# 或者用 git 管理
cd ~/.shifter/profiles
git init && git add -A && git commit -m "my profiles"
```

### Q: 如何添加对新 Agent 的支持？

Shifter 采用适配器架构，添加新的 Agent 只需：

1. 在 `adapter/<agent-name>/` 下实现 `AgentAdapter` 接口
2. 在 `registry/registry.go` 中注册

欢迎提交 PR！

### Q: 支持环境变量吗？

支持。MCP 服务器配置中的环境变量（如 `${GITHUB_TOKEN}`）会被迁移时保留。

---

## 快速参考卡片

```bash
# 查看帮助
shifter --help
shifter port --help

# 日常使用三板斧
shifter detect                          # 看看有哪些 Agent
shifter port claude-code --to codex     # 迁移配置
shifter save my-config --source claude-code  # 保存为模板

# 模板管理
shifter profiles                        # 我的模板
shifter load my-config --to opencode    # 在新项目中加载模板
shifter profiles delete old-one         # 删除旧模板

# 安全操作
shifter port A --to B --dry-run         # 先预览
shifter backup --list                   # 查看备份
shifter backup --restore <文件>         # 恢复备份

# 高级功能
shifter sync claude-code codex         # 双向同步
shifter ui                             # 交互式向导
shifter status --json                  # JSON 状态输出
```
