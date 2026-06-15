# Testing Standards
> Group: Mandatory | Severity: High

## Test Structure
- 使用 Arrange-Act-Assert (AAA) 模式
- 测试名格式: `should_<预期行为>_when_<条件>`
- 每个测试只验证一个行为

## Coverage Requirements
- 关键业务逻辑: 最低 80% 覆盖率
- 安全相关代码: 最低 90% 覆盖率
- 工具函数: 最低 95% 覆盖率
- 新代码必须包含测试

## Test Types
- **单元测试**: 隔离、快速、无外部依赖 (Vitest / go test)
- **集成测试**: API 端点测试 (supertest / httptest)
- **E2E 测试**: 关键用户流程 (Playwright)

## Mock 策略
- Mock 外部依赖（API 调用、数据库连接）
- 不 Mock 内部逻辑
- 使用依赖注入提高可测试性
- 测试间重置 mock 状态

## Bug 修复流程
1. 先写一个能复现 bug 的测试
2. 确认测试失败
3. 修复代码
4. 确认测试通过
5. 不能只说"修好了"而测试仍在失败

## Definition of Done
- 所有测试通过
- Lint 检查通过 (`ruff check`, `eslint`)
- 没有死代码或注释掉的代码
- 相关文档已更新
