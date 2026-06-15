# Tech Stack Constraints
> Group: Mandatory | Severity: Medium

## 技术栈锁定

### 前端
- **框架**: Next.js 14 (App Router) + React 18
- **语言**: TypeScript strict mode
- **样式**: Tailwind CSS + shadcn/ui 组件库
- **状态管理**: React Context + useReducer (不用 Redux)
- **包管理**: Bun

### 后端
- **框架**: Go 1.22+ + Chi Router
- **数据库访问**: sqlc (禁止手写 SQL)
- **迁移**: golang-migrate
- **缓存**: Redis (go-redis)

### 禁止使用
- ~~Express/Fastify~~ → 使用 Go Chi
- ~~Prisma~~ → 使用 sqlc
- ~~CSS Modules~~ → 使用 Tailwind
- ~~Redux~~ → 使用 Context + useReducer
- ~~any 类型~~ → 使用 unknown + type guards

### 添加新依赖
- 必须有明确的业务需求
- 检查包的安全性和维护状态
- 评估 bundle size 影响
- 优先选择生态内已有方案
