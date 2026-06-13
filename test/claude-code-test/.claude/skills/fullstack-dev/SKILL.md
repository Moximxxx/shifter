---
name: fullstack-dev
description: 全栈开发工作流 — 从需求分析到代码实现到部署的完整流程
allowed_tools: [Read, Write, Edit, Glob, Grep, Bash, WebFetch, WebSearch]
---
你正在执行全栈开发工作流。遵循以下阶段：

## Phase 1: 需求分析
- 理解用户需求，提出澄清问题
- 确定影响范围和依赖

## Phase 2: 方案设计
- 输出至少 2 个技术方案
- 每个方案包含优缺点、风险、工作量
- 标明推荐方案

## Phase 3: 实施
- 按照项目编码规范编写代码
- TypeScript strict mode, Go Effective Go
- 所有用户输入使用 Zod 验证
- 数据库查询使用 sqlc（禁止手写 SQL）

## Phase 4: 测试
- 单元测试: Arrange-Act-Assert
- 边界条件测试
- 运行 `bun test && go test ./...` 确认通过

## Phase 5: 审查
- 自行审查代码
- 检查安全漏洞
- 确认没有硬编码密钥

## Phase 6: 提交
- 变更文件列表
- 建议的提交信息 (Conventional Commits 格式)
- 等待用户确认后提交
