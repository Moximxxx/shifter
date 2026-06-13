# Git Workflow
> Group: Process | Severity: Medium

## Branch Strategy
- `main` — 生产就绪代码，受保护
- `feature/<name>` — 新功能
- `fix/<name>` — Bug 修复
- `refactor/<name>` — 重构

## Commit Convention
使用 Conventional Commits 格式:
```
type(scope): description

[optional body]
[optional footer]
```

类型: feat, fix, docs, style, refactor, test, chore, ci, perf
范围: 受影响的模块名 (auth, api, ui, db 等)

示例:
```
feat(auth): add JWT refresh token rotation
fix(api): handle empty result set in user query
refactor(db): migrate from raw SQL to sqlc
```

## Push 规则
- 禁止 force push 到 main/master
- 提交前运行 `bun test` 和 `go test ./...`
- 不要提交未完成的工作
- WIP 提交使用 `git commit --amend` 整理后再推送

## 提交时机
- 等待用户明确确认后再提交
- 不要自动提交，让用户审查变更
- 完成后建议提交信息，由用户决定
