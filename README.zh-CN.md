# 🔄 Shifter

**一键在不同 Coding Agent 之间迁移配置。**

把你配置好的 agents、skills、MCP 服务器、权限、钩子、项目指令从一个 Agent 搬到另一个。配置一次，到处使用。

```bash
# 一行命令把 Claude Code 的配置迁到 Codex
shifter port claude-code --to codex

# 或用交互式向导
shifter ui
```

## 支持的 Agent

| Agent | 项目指令 | 子代理 | 技能 | 命令 | MCP | 权限 | 钩子 |
|-------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Claude Code** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Codex CLI** | ✓ | ✓ | ✓ | 嵌入 | ✓ | ✓ | ✓ |
| **OpenCode** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | — |
| **Gemini CLI** | ✓ | — | — | — | ✓ | ✓ | — |
| **Qoder** | — | ✓ | ✓ | — | ✓ | — | — |
| **Cline** | ✓ | 嵌入 | — | — | — | — | — |
| **Aider** | ✓ | — | — | — | — | ✓ | — |

> `—` = 原生不支持; `嵌入` = 以项目指令形式嵌入

> 📖 完整文档见 **[中文操作手册 (MANUAL.zh-CN.md)](MANUAL.zh-CN.md)**

## 安装

### 从源码编译
```bash
cd /home/qin/code/shifter
go build -o shifter .
sudo mv shifter /usr/local/bin/
```

### go install
```bash
go install github.com/moximxxx/shifter@latest
```

## 快速上手

```bash
# 看看当前项目有哪些 Agent
shifter detect

# 把 Claude Code 配置迁到 Codex（先预览再执行）
shifter port claude-code --to codex --dry-run
shifter port claude-code --to codex

# 保存当前项目的 Agent 配置为全局模板
shifter save my-setup --source claude-code --desc "我的开发配置"

# 在新项目中加载模板
cd ~/projects/new-project
shifter load my-setup --to codex

# 启动交互式向导（全部操作可视化）
shifter ui
```

## 核心概念

```
源 Agent（如 Claude Code）
    │  读取 .claude/settings.json、agents/、skills/ 等
    ▼
内部通用模型（用户无需关心）
    │  语义映射 + 损失追踪
    ▼
目标 Agent（如 Codex CLI）
    │  写入 .codex/config.toml、codex.md 等
```

- **语义映射**：理解每个配置的含义，不只是格式转换
- **损失追踪**：不能完美迁移的功能会生成明确警告，不会悄悄丢失
- **自动备份**：每次写入自动创建带时间戳的 tar.gz 备份，随时可撤销
- **安全写入**：写入前验证语法正确性，绝不写出损坏的配置

## 命令一览

| 命令 | 用途 |
|------|------|
| `shifter detect` | 扫描系统中已配置的 Agent |
| `shifter status` | 查看当前项目 Agent 配置状态 |
| `shifter port <源> --to <目标>` | 单向配置迁移 |
| `shifter sync <A> <B>` | 双向配置同步 |
| `shifter save <名称> --source <A>` | 保存为全局模板 |
| `shifter load <名称> --to <B>` | 加载全局模板 |
| `shifter profiles` | 管理全局模板 |
| `shifter backup <Agent>` | 管理配置备份 |
| `shifter ui` / `shifter wizard` | 启动交互式向导 |

## 三个核心工作流

### 🔀 迁移（Port）
在当前项目中，把一个 Agent 的配置搬到另一个 Agent。

```bash
shifter port claude-code --to codex
shifter port claude-code --to codex --aspects agents,mcp  # 只要 agents 和 MCP
```

### 💾 保存（Save）
把当前项目的 Agent 配置保存为全局模板，下次在新项目中直接加载。

```bash
shifter save team-standards --source claude-code --desc "团队标准配置"
shifter load team-standards --to codex    # 在新项目中加载
shifter load team-standards --to opencode  # 同一个模板，不同 Agent
```

### 🔄 同步（Sync）
双向同步两个 Agent 的配置，保持一致性。

```bash
shifter sync claude-code codex --strategy newer
```

## 项目结构

```
shifter/
├── adapter/          # Agent 适配器（每种 Agent 一个）
│   ├── claudecode/   # Claude Code
│   ├── codex/        # Codex CLI
│   ├── opencode/     # OpenCode
│   ├── gemini/       # Gemini CLI
│   ├── qoder/        # Qoder
│   ├── cline/        # Cline
│   └── aider/        # Aider
├── canonical/        # 通用配置模型 + 合并逻辑
├── engine/           # 核心业务逻辑
│   ├── port/         # 单向迁移
│   ├── sync/         # 双向同步
│   ├── detect/       # Agent 扫描
│   ├── profile/      # 模板管理
│   └── backup/       # 备份恢复
├── tui/              # Bubble Tea 交互式终端界面
├── cmd/              # CLI 命令（cobra）
└── pkg/format/       # 格式工具（frontmatter、JSONC）
```

## 📖 完整文档

**[中文操作手册 (MANUAL.zh-CN.md)](MANUAL.zh-CN.md)** — 包含所有使用场景、命令参考、配置对照表、FAQ

## 开发

```bash
make build    # 编译
make test     # 测试
make check    # 完整检查
```

## 许可

MIT
