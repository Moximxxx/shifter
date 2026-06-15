# Security Rules
> Group: Mandatory | Severity: Critical

## Rule
所有输入必须经过验证。禁止在代码中硬编码密钥。禁止 SQL 字符串拼接。

## Input Validation
- 所有用户输入使用 Zod (前端) 和 go-playground/validator (后端) 验证
- 文件上传限制：类型白名单、最大 10MB
- URL 参数必须经过 encodeURIComponent 处理

## Secrets Management
- 所有密钥通过环境变量加载，禁止硬编码
- `.env` 文件不得提交到 Git（已在 .gitignore 中）
- 生产环境密钥通过 GitHub Secrets 或 Vault 注入

## Injection Prevention
- 数据库查询使用参数化查询（sqlc 自动处理）
- 禁止使用 `eval()`、`Function()`、`dangerouslySetInnerHTML`
- HTML 渲染使用 DOMPurify 净化

## Dependency Security
- 定期运行 `npm audit` 和 `go mod tidy`
- CI 中启用 Dependabot 或 Renovate
- 禁止安装未审核的第三方包
