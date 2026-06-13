# Coding Standards
> Group: Mandatory | Severity: High

## Naming Conventions
- **文件**: TypeScript: kebab-case.tsx, Go: snake_case.go
- **组件**: PascalCase (UserProfile, OrderList)
- **函数**: camelCase, 布尔值加 is/has 前缀 (isLoading, hasError)
- **常量**: UPPER_SNAKE_CASE
- **接口/类型**: PascalCase, Props 类型以 Props 结尾

## Code Style
- 缩进: 2 spaces
- 行长度: 最大 100 字符
- 引号: TypeScript 用单引号, Go 用双引号
- 分号: TypeScript 不需要, Go 自动处理
- 多行结构始终加尾随逗号

## React 规范
- 优先使用函数组件 + Hooks
- 使用 `const` 定义组件，避免 `function` 声明
- Props 类型定义与组件同文件
- 使用 early return 提高可读性
- 避免在渲染中创建新对象/函数（使用 useMemo/useCallback）

## Go 规范
- 遵循 Effective Go
- context.Context 作为第一个参数
- 错误处理: `if err != nil` 立即返回或包装
- 使用依赖注入，避免全局状态
- 导出函数必须有注释
