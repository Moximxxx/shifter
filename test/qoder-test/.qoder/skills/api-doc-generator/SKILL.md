---
name: api-doc-generator
description: |
  当开发者添加、修改或删除 API 端点时，自动执行 API 文档同步、向后兼容性检查和单元测试生成。
  触发短语: API 文档、API spec、接口文档、Swagger、OpenAPI、兼容性检查、API 变更
license: MIT
metadata:
  author: Shifter Team
  version: 1.0.0
  category: development
  tags: [api, documentation, testing, openapi]
references:
  - references/openapi-guide.md
---

## 目标
自动从代码中提取 API 定义，生成/更新 OpenAPI 3.0 文档，并检查向后兼容性。

## 执行步骤

### Step 1: 扫描 API 端点
扫描项目中的 API 路由定义：
- Go: 扫描 `internal/handler/` 下的路由注册
- TypeScript: 扫描 `src/app/api/` 下的 route handlers

### Step 2: 提取接口信息
从代码中提取：
- HTTP 方法和路径
- 请求参数 (query, path, body)
- 响应格式和状态码
- 认证要求

### Step 3: 生成/更新 OpenAPI 文档
生成符合 OpenAPI 3.0 规范的文档：
```bash
python scripts/generate_openapi.py --output docs/api/openapi.yaml
```

### Step 4: 兼容性检查
比较新旧 API 文档，检查：
- 删除的端点
- 修改的请求参数
- 修改的响应格式
- 新增的必填字段

## 错误处理
- **错误**: 无法解析路由定义
  - **原因**: 非标准的路由注册方式
  - **修复**: 检查 `scripts/routes-parser.py` 中的路由模式匹配规则
