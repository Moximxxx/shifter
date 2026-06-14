# 🔄 Shifter

**一键在不同 Coding Agent 之间迁移配置。** 配置一次，到处使用。

```bash
shifter port claude-code --to codex       # 一行搞定
shifter ui                                # 交互式向导
```

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/Moximxxx/shifter/main/install.sh | sh
```

## 支持的 Agent

| Agent | 子代理 | 技能 | 命令 | MCP | 权限 | 钩子 |
|-------|:---:|:---:|:---:|:---:|:---:|:---:|
| **Claude Code** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Codex CLI** | ✓ | ✓ | 嵌入 | ✓ | ✓ | ✓ |
| **OpenCode** | ✓ | ✓ | ✓ | ✓ | ✓ | — |
| **Qoder** | ✓ | ✓ | — | ✓ | — | — |
| **Gemini CLI** | — | — | — | ✓ | ✓ | — |
| **Cline** | 嵌入 | — | — | — | — | — |
| **Aider** | — | — | — | — | ✓ | — |

> `—` = 原生不支持; `嵌入` = 以项目指令形式嵌入  
> 📖 完整文档见 **[中文操作手册 (MANUAL.zh-CN.md)](MANUAL.zh-CN.md)**

## 快速上手

```bash
shifter detect                              # 扫描已配置的 Agent
shifter port claude-code --to codex         # 迁移配置
shifter port --dry-run                      # 预览变更 (配合环境变量)
shifter ui                                  # 交互式向导
```

### 环境变量（可选）

```bash
eval "$(shifter env init)"
export SHIFTER_SOURCE=claude-code
export SHIFTER_TARGET=codex
shifter port          # 无需参数
```

## 命令一览

| 命令 | 用途 |
|------|------|
| `shifter port <源> --to <目标>` | 单向配置迁移 |
| `shifter export <agent>` | 导出为私有 JSON 格式 |
| `shifter import <文件> --to <agent>` | 从私有格式导入 |
| `shifter sync <A> <B>` | 双向同步 |
| `shifter save <名称> --source <A>` | 保存为全局模板 |
| `shifter load <名称> --to <B>` | 加载全局模板 |
| `shifter detect` | 扫描已配置的 Agent |
| `shifter status` | 查看项目 Agent 状态 |
| `shifter backup <agent>` | 备份/恢复 |
| `shifter env` | 环境变量配置 |
| `shifter ui` / `shifter wizard` | 交互式向导 |

## 核心概念

```
源 Agent ──Read──▶ 私有格式 JSON ──Write──▶ 目标 Agent
(Claude Code)      (Canonical)            (Codex/OpenCode/Qoder)
```

- **语义映射**：理解配置含义，不只是格式转换
- **损失追踪**：不能完美迁移的功能生成明确警告
- **自动备份**：每次写入创建带时间戳的备份，随时可撤销
- **安全写入**：写入前验证语法，绝不写出损坏的配置

## 项目结构

```
shifter/
├── adapter/          # Agent 适配器（7 种，解析中心）
├── canonical/        # 通用配置模型（私有格式）
├── engine/           # 核心引擎（port/sync/detect/backup/profile）
├── registry/         # 适配器工厂（添加新 Agent 只需注册）
├── tui/              # Bubble Tea 交互式终端界面
├── cmd/              # CLI 命令（cobra）
├── pkg/              # 工具库（env/format/paths）
├── test/             # 全部 7 个 Agent 的真实工作流测试固件
├── dist/             # Release 二进制
└── install.sh        # 一键安装脚本
```

## 📖 完整文档

**[中文操作手册 (MANUAL.zh-CN.md)](MANUAL.zh-CN.md)**

## 许可

MIT
