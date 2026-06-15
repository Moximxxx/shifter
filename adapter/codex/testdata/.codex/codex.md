# Project: Fullstack SaaS Platform

## Tech Stack
- **Frontend**: Next.js 14 + TypeScript + Tailwind CSS + shadcn/ui
- **Backend**: Go 1.22 + Chi Router + sqlc
- **Database**: PostgreSQL 16 + Redis
- **Testing**: Vitest (frontend) + Go testing (backend)
- **CI/CD**: GitHub Actions + Docker

## Project Structure
```
src/
  app/          # Next.js App Router pages
  components/   # React components (ui/ and features/)
  lib/          # Shared utilities
  server/       # API route handlers
cmd/
  api/          # Go API server entrypoint
internal/
  handler/      # HTTP handlers
  service/      # Business logic
  repository/   # Data access (sqlc generated)
  middleware/   # Auth, logging, rate limiting
migrations/     # PostgreSQL migrations
```

## Coding Conventions

### TypeScript / React
- Use functional components with hooks
- Prefer `const` over `function` for component definitions
- Use early returns for readability
- All API calls through `src/lib/api/client.ts`
- Component files: PascalCase.tsx
- Utility files: kebab-case.ts

### Go
- Follow Effective Go and standard project layout
- Use `context.Context` as first parameter
- Error wrapping: `fmt.Errorf("context: %w", err)`
- No global state — use dependency injection
- Test files: `*_test.go` co-located with source

### Database
- All queries through sqlc (no raw SQL in code)
- Migration files: `NNNN_description.up.sql` / `.down.sql`
- No foreign key cascades in production (handle in app layer)

## Security Rules
- Validate all user input with Zod (frontend) and go-playground/validator (backend)
- JWT tokens: RS256, 15min access + 7d refresh, httpOnly cookies
- Rate limiting: 100 req/min per IP, 1000 req/min per user
- Secrets in environment variables only — never commit `.env` files
- All external API calls through backend proxy (no direct frontend calls)

## Testing Standards
- Unit tests for all business logic (80%+ coverage)
- Integration tests for API endpoints
- E2E tests for critical user journeys (login, checkout, dashboard)
- Run `bun test` and `go test ./...` before committing

## Git Workflow
- Branch: `feature/`, `fix/`, `refactor/` prefixes
- Conventional Commits: `type(scope): description`
- Squash merge to main
- Never force push to main/master

## Agent Workflow (Codex Multi-Agent)
```
1. 接收任务 → architect 代理分析技术方案（不少于2个方案对比）
2. 方案确认 → task-executor 代理实施（按合同规范修改指定文件）
3. 代码完成 → code-reviewer 代理审查（安全+性能+可维护性+类型安全）
4. 审查通过 → test-writer 代理生成/补充测试
5. 测试通过 → builder 代理构建验证
6. 构建成功 → 代码合并
```
