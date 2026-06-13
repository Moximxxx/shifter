# Project Context: Fullstack SaaS Platform

## 1. Project Overview
**目标**: 全栈 SaaS 平台，提供用户管理、订阅计费、数据分析仪表盘。
**核心功能**: 用户认证 (OAuth + JWT)、订阅管理 (Stripe)、实时数据仪表盘、API 限流和配额管理。

## 2. Tech Stack
| 层级 | 技术 |
|------|------|
| **前端** | Next.js 14 + TypeScript + Tailwind CSS + shadcn/ui |
| **后端** | Go 1.22 + Chi Router |
| **数据库** | PostgreSQL 16 + Redis |
| **测试** | Vitest (前端) + Go testing (后端) + Playwright (E2E) |
| **CI/CD** | GitHub Actions + Docker + Kubernetes |
| **监控** | Prometheus + Grafana + OpenTelemetry |

## 3. Project Structure
```
src/
  app/          # Next.js App Router (页面路由)
  components/
    ui/         # shadcn/ui 基础组件
    features/   # 业务功能组件
  lib/          # 工具函数 (api client, auth, utils)
  hooks/        # 自定义 React Hooks
cmd/api/        # Go API 入口
internal/
  handler/      # HTTP 处理器 (路由 → Service)
  service/      # 业务逻辑层
  repository/   # 数据访问层 (sqlc 生成)
  middleware/    # Auth、CORS、限流、日志
  model/        # 领域模型定义
migrations/     # PostgreSQL 迁移文件
```

## 4. Key Commands
| 命令 | 说明 |
|------|------|
| `bun dev` | 启动前端开发服务器 (:3000) |
| `go run ./cmd/api` | 启动后端 API (:8080) |
| `bun test` | 运行前端测试 |
| `go test ./...` | 运行后端测试 |
| `bun run build` | 构建前端 |
| `make docker-build` | 构建 Docker 镜像 |
| `make lint` | 运行所有语言的 lint |
| `make migrate-up` | 运行数据库迁移 |

## 5. Coding Conventions
- **TypeScript**: strict mode, 单引号, 无分号, trailing commas
- **Go**: Effective Go, context-first, error wrapping, no global state
- **命名**: 文件 kebab-case / snake_case, 组件 PascalCase, 函数 camelCase
- **导入**: React → 第三方 → @/ 别名 → 相对路径
- **括号**: JSX 属性使用双引号, 字符串使用单引号

## 6. Architecture Constraints
- API 遵循 RESTful 设计
- 数据库访问只能通过 sqlc（禁止手写 SQL）
- 前端不直接调用外部 API（必须通过后端代理）
- 所有状态变更通过服务层，控制器只做路由和验证
- 微服务之间的通信通过 gRPC（未来）
- 事件驱动架构用于异步任务（RabbitMQ/Redis Streams）

## 7. Security Checklist
- [ ] 所有用户输入验证 (Zod + go-playground/validator)
- [ ] JWT RS256 + httpOnly cookies
- [ ] API 限流 (100 req/min per IP)
- [ ] CORS 白名单
- [ ] SQL 参数化查询 (sqlc)
- [ ] 密码 bcrypt cost ≥ 12
- [ ] HTTPS only (生产环境)
- [ ] CSP + HSTS 安全头

## 8. Current Sprint Goals
- 实现用户仪表盘实时数据更新 (WebSocket)
- 添加团队协作功能（多用户共享项目）
- 优化数据库查询性能（添加索引、查询缓存）
- 迁移测试从 Jest 到 Vitest
