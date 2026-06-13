# Agent Workflow & Delegation
> Group: Process | Severity: High

## 子代理使用规范

### 主代理 (Main Agent)
- 唯一拥有 Write/Edit/Bash 权限的代理
- 负责任务分解、委派、验证
- 维护 `.cline/memory.md` 作为跨会话记忆

### 子代理 (Subagents)
- 仅限只读操作（Read、Glob、Grep）
- 用于扩展上下文窗口的调研任务
- 子代理完成后的输出必须由主代理审核

### 委派触发条件
- **广泛探索**: 映射大型/不熟悉目录的依赖关系
- **模式匹配**: 复杂正则或符号搜索
- **文档合成**: 阅读和总结外部文档或 README

### 命名规范
每个子代理按调研目标命名:
- `feature-gate-audit`
- `schema-dependency-mapper`
- `api-route-validator`

### 交接流程
1. 子代理完成后，输出"调研报告"
2. 主代理确认调研发现是否需要改变当前技术方案
3. 确认后才进行编辑操作

## 记忆管理
- 每次任务执行前读取 `.cline/memory.md`
- 任务执行中持续更新 `.cline/memory.md`
- 有价值的长期知识保存到 `.cline/memory_data/` 目录
- 维护 `.cline/memory_data/index.md` 索引

## 知识来源优先级
1. `.cline/memory_data/` > 项目文档 > 浏览器搜索
2. 不确定时先询问用户，不要猜测
