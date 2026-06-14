# 🔄 Shifter

[📖 English](README.md)

<p align="center">
  <strong>一键在不同 Coding Agent 之间迁移配置。</strong><br>
  配置一次 agents、skills、MCP、hooks，在 Claude Code、Codex、OpenCode、Qoder 中通用。
</p>

<p align="center">
  <img src="https://img.shields.io/badge/version-0.1.0-7C3AED?style=flat-square" alt="version">
  <img src="https://img.shields.io/badge/go-1.24%2B-00ADD8?style=flat-square&logo=go" alt="go">
  <img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="license">
  <img src="https://img.shields.io/badge/agents-7-10B981?style=flat-square" alt="agents">
</p>

---

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/Moximxxx/shifter/release/install.sh | sh
```

## 快速上手

```bash
shifter detect                              # 扫描已配置的 Agent
shifter port claude-code --to codex         # 迁移配置
shifter ui                                  # 交互式向导
```

配合环境变量，连参数都不用写：

```bash
export SHIFTER_SOURCE=claude-code
export SHIFTER_TARGET=codex
shifter port        # 成了
```

## 工作原理

```
┌──────────────┐       ┌──────────────────┐       ┌──────────────┐
│ Claude Code  │  Read │  私有格式 JSON    │ Write │  Codex CLI   │
│  .claude/    │ ────▶ │  (Canonical)     │ ────▶ │  .codex/     │
└──────────────┘       └──────────────────┘       └──────────────┘
```

读取原生配置 → 转为通用中间格式 → 写入任意目标 Agent。语义映射处理格式差异；不能完美迁移的功能生成明确警告 — 绝不悄悄丢失。

## 功能

|     | 功能 | 说明 |
|:---:|------|------|
| 🔀 | **Port** | Agent 间迁移：`shifter port claude-code --to codex --aspects agents,mcp` |
| 💾 | **Profiles** | 保存/加载可复用模板：`shifter save team-setup --source claude-code` |
| 📤 | **Export/Import** | 私有 JSON 可移植格式：`shifter export claude-code \| shifter import - --to codex` |
| 🔄 | **Sync** | 双向同步 + 合并策略：`shifter sync claude-code codex --strategy newer` |
| 🎨 | **TUI 向导** | 交互式引导流程：`shifter ui` |
| ⚠️ | **损失追踪** | 不兼容功能生成明确严重级别警告 |
| 💾 | **自动备份** | 每次写入创建带时间戳的 tar.gz — 随时撤销 |

## 支持的 Agent

| Agent | 子代理 | 技能 | MCP | 钩子 | 格式 |
|-------|:---:|:---:|:---:|:---:|--------|
| **Claude Code** | ✓ | ✓ | ✓ | ✓ | JSON + MD |
| **Codex CLI** | ✓ | ✓ | ✓ | ✓ | TOML |
| **OpenCode** | ✓ | ✓ | ✓ | — | JSONC |
| **Qoder** | ✓ | ✓ | ✓ | — | MD frontmatter |
| **Gemini CLI** | — | — | ✓ | — | JSON |
| **Cline** | 嵌入 | — | — | — | Markdown |
| **Aider** | — | — | — | — | YAML |

> `嵌入` = 内容以指令形式保留。完整能力矩阵见 [操作手册](MANUAL.zh-CN.md)。

## 命令

```bash
shifter port      <源> --to <目标>    # 单向配置迁移
shifter export    <agent>             # 导出为私有 JSON 格式
shifter import    <文件> --to <目标>  # 从私有格式导入
shifter sync      <A> <B>             # 双向同步
shifter save      <名称> --source <A> # 保存为全局模板
shifter load      <名称> --to <B>     # 加载模板
shifter detect                        # 扫描已配置的 Agent
shifter status                        # 查看项目状态
shifter backup    <Agent>             # 创建/恢复备份
shifter profiles                      # 管理模板
shifter env                           # 显示环境配置
shifter ui                            # 交互式 TUI 向导
```

## 架构

```
adapter/           ═══════════ 解析中心 ═══════════
├── claudecode/    Read  .claude/*              → canonical
├── codex/         Read  .codex/config.toml      → canonical
├── opencode/      Read  opencode.jsonc          → canonical
├── qoder/         Read  .qoder/*                → canonical
├── gemini/        Read  .gemini/settings.json   → canonical
├── cline/         Read  .clinerules/*           → canonical
└── aider/         Read  .aider.conf.yml         → canonical

canonical/types.go ═══════ 私有格式 ═══════
                   通用 JSON 超集，覆盖所有 Agent

registry/          ═══════ 适配器工厂 ═══════
                   Get("codex") → Write(canonical) → .codex/config.toml

engine/            ═══════ 业务引擎 ═══════
├── port/          Source → Canonical → Target 流水线
├── sync/          双向合并引擎
├── detect/        并发文件系统扫描
├── backup/        带时间戳的 tar.gz 备份/恢复
└── profile/       ~/.shifter/profiles/ 管理
```

添加新 Agent：实现 `AgentAdapter`（Read + Write），注册到 `registry/`，完成。

## 环境变量

```bash
eval "$(shifter env init)"    # 生成 shell 配置

SHIFTER_SOURCE=claude-code    # 默认源 Agent
SHIFTER_TARGET=codex          # 默认目标 Agent
SHIFTER_DRY_RUN=1             # 始终先预览
SHIFTER_ASPECTS=agents,mcp    # 默认迁移维度
```

## 操作手册

完整文档 — 所有命令详解、环境变量、配置文件位置、FAQ、损失等级说明。**[Shifter 操作手册 →](MANUAL.zh-CN.md)**

## 开发

```bash
make build          # 编译（含版本信息）
make test           # 运行测试
make check          # fmt + vet + test + build
make release        # 全平台交叉编译
```

---

<p align="center">
  <sub>MIT · <a href="https://github.com/Moximxxx/shifter">GitHub</a> · <a href="https://github.com/Moximxxx/shifter/releases">Releases</a></sub>
</p>
