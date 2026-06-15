# CLAUDE.md — Shifter Project

## Project Overview

**Shifter** 是一个跨 Coding Agent 项目级工作流无缝迁移工具。解决的核心场景：

- **多工具并行开发**：同一项目需要在 OpenCode、Claude Code、Codex 等多个 Agent 之间切换，配置（agents/skills/MCP/hooks/permissions）需要在不同工具间保持一致
- **团队协作**：团队成员使用不同的 Coding Agent（有人用 Claude Code，有人用 Codex，有人用 OpenCode），但需要共享同一套项目工作流配置
- **配置复用**：在一个项目中调好的 Agent 配置，一键迁移到新项目或分享给团队

## Architecture

```
Source Agent ──Read──▶ Canonical JSON ──Write──▶ Target Agent
(Claude Code)          (Private Format)          (Codex/OpenCode/Qoder)
```

核心设计：**解析中心 + 私有格式 + 适配器工厂**
- `adapter/` — 7 个 Agent 适配器，每个实现 `Read()` 和 `Write()`
- `canonical/` — 通用配置模型（JSON 超集，覆盖所有 Agent 能力）
- `registry/` — 适配器工厂，添加新 Agent 只需注册一行
- `engine/` — 业务引擎（port/sync/detect/backup/profile）

## Tech Stack

- **Language**: Go 1.22+
- **CLI**: cobra
- **TUI**: Bubble Tea + Lipgloss
- **Config parsing**: YAML (gopkg.in/yaml.v3), TOML (BurntSushi/toml), JSONC
- **Testing**: go test, table-driven, golden files

## Key Commands

```bash
go build -ldflags "-X main.version=$(ver) -X main.commit=$(git rev-parse --short HEAD)" -o shifter .
go test ./...                      # All tests
go test ./test/golden/... -update  # Update golden files
go test ./... -coverprofile=c.out  # Coverage
make build test release            # Build/test/release
```

## Development Workflow

```bash
# 1. Create feature branch from develop
git checkout develop
git checkout -b feature/xxx

# 2. Test-driven development
# Write test → Run (fail) → Implement → Run (pass)

# 3. PR to release branch
gh pr create --base release --head feature/xxx

# 4. After merge, tag and release
git checkout release && git pull
make release VERSION=x.y.z
gh release create vx.y.z dist/*.tar.gz
```

## Branch Model

```
release   ← 默认分支，打 tag 发布
  ↑ PR
develop   ← 日常开发
  ↑ branch
feature/* ← 新功能
```

## Testing Strategy

- **单元测试**: 表驱动 + golden files（期望输出快照对比）
- **集成测试**: adapter round-trip（Read→Write→Read，验证内容一致性）
- **E2E 测试**: 真实 CLI 调用 `exec.Command("shifter", ...)`
- **Golden files**: `adapter/*/testdata/` 下 63 个快照文件，CI 中对比防止回归

## Key Design Decisions

1. **Canonical model 为内部格式**（用户不需要关心）— 区别于 AgentsMesh 等工具
2. **Go 单二进制分发**（非 Node.js）— 零依赖，`curl | sh` 安装
3. **每次写入自动备份** — 信任基础
4. **损失追踪是一等公民** — 不能迁移的功能生成明确警告，绝不静默丢弃
5. **项目级配置优先**（非全局）— TUI 只显示当前项目配置的 Agent

## File Structure

```
shifter/
├── adapter/           # Agent adapters (one per coding agent)
│   ├── claudecode/    # Claude Code
│   ├── codex/         # Codex CLI
│   ├── opencode/      # OpenCode
│   ├── qoder/         # Qoder
│   ├── gemini/        # Gemini CLI
│   ├── cline/         # Cline
│   └── aider/         # Aider
├── canonical/         # Universal config model + merge logic
├── engine/            # Business logic (port, sync, detect, backup, profile)
├── registry/          # Adapter factory
├── tui/               # Bubble Tea interactive terminal UI
├── cmd/               # CLI commands (cobra)
├── pkg/               # Utilities (i18n, paths, format, log, convert, env, logo, settings)
├── test/              # Fixtures + E2E tests
└── dist/              # Release binaries
```

## Conventions

- 所有 exported symbols 必须有 GoDoc 注释
- 错误处理：`fmt.Errorf("context: %w", err)` 包装，不吞错误
- 权限常量：使用 `paths.DirPerm` / `paths.FilePerm`，不用 magic number
- 目录：使用 `paths.MustHomeDir()` 替代 `os.UserHomeDir()`
- i18n：所有 TUI 字符串通过 `i18n.T("key")` 获取，中英文在 `pkg/i18n/locales/`
- Logging：`--log` 标志启用，写入 `~/.shifter/logs/shifter-YYYYMMDD.log`

## Rules

### R-01: i18n Required for All User-Facing Strings

所有面向用户的字符串必须通过 `i18n.T("key")` 获取，中英文翻译文件分别在 `pkg/i18n/locales/en.json` 和 `zh.json`。

- 新增 TUI 文字 → 必须添加中英文翻译
- 新增 CLI 输出 → 考虑使用 i18n 翻译
- 提交前检查 `pkg/i18n/locales/` 两个文件是否同步更新

详见: [docs/rules/R-01-i18n.md](docs/rules/R-01-i18n.md)

## Rules Index

- [R-01: i18n Required](docs/rules/R-01-i18n.md) — 所有 UI 字符串必须国际化

## Incidents Index

- [INC-2026-06-16: Help Bar Duplication](docs/incidents/INC-2026-06-16-helpbar-duplication.md) — i18n 与 Go 代码按键前缀重复
