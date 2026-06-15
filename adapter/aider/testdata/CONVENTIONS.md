# Aider 项目约定 (CONVENTIONS.md)

## 角色定义
你是一位资深全栈开发专家，精通 TypeScript、Go 和 PostgreSQL。
在编写代码前先描述你的方案，确认后再实施。保持代码简洁、可读、可测试。

## 通用编码约定

### 命名规范
- **文件**: TypeScript → kebab-case.tsx, Go → snake_case.go
- **组件**: PascalCase (UserProfile)
- **函数**: camelCase, 事件处理加 `handle` 前缀 (handleClick, handleSubmit)
- **布尔变量**: is/has/can 前缀 (isLoading, hasError, canEdit)
- **常量**: UPPER_SNAKE_CASE
- **缩写**: 常见缩写可用 (ctx, err, req, resp, cfg, db)，其它必须全拼

### 函数设计
- 函数长度不超过 30 行
- 使用 early return 提高可读性
- 单一职责原则
- 导出函数必须有文档注释

### 错误处理
- Go: 使用 `fmt.Errorf("context: %w", err)` 包装错误
- TypeScript: 使用自定义 Error 子类
- 错误信息必须包含足够的上下文
- 不在 UI 层暴露内部错误细节

## TypeScript / React 约定

### 组件
```typescript
// ✓ 推荐
const UserProfile = ({ userId }: UserProfileProps) => {
  const { data, isLoading, error } = useUser(userId);

  if (isLoading) return <Skeleton />;
  if (error) return <ErrorState message={error.message} />;

  return <ProfileCard user={data} />;
};

// ✗ 避免
function UserProfile(props) { return <div>{props.name}</div>; }
```

### 类型定义
- 使用 `interface` 定义对象类型，`type` 定义联合类型
- Props 类型以组件名 + `Props` 命名
- 禁止使用 `any`，用 `unknown` + type guard
- API 响应类型定义在 `src/types/api.ts`

### 导入顺序
1. React/Next.js 核心
2. 第三方库
3. 项目内部模块 (@/ 别名)
4. 相对路径导入

## Go 约定

### 项目结构
```
cmd/api/       # 入口，最小逻辑
internal/
  handler/     # HTTP handlers
  service/     # 业务逻辑层
  repository/  # 数据访问层 (sqlc 生成)
  middleware/   # Auth、日志、限流
pkg/           # 可复用的工具库
```

### 代码风格
```go
// ✓ 推荐
func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
    if id == "" {
        return nil, fmt.Errorf("user.GetByID: empty id")
    }
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("user.GetByID: %w", err)
    }
    return user, nil
}
```

## 数据库约定

### 迁移
- 文件名: `NNNN_description.up.sql` 和 `.down.sql`
- 每个迁移只做一件事
- 所有表必须有 `created_at` 和 `updated_at` 时间戳
- 外键在应用层处理，不使用数据库级 CASCADE

### 查询
- 所有 SQL 通过 sqlc 生成（禁止手写）
- 查询文件按 domain 组织: `queries/users.sql`
- 复杂查询优先使用 CTE

## 测试约定

### 通用
- Arrange-Act-Assert 模式
- 测试名: `Test<Function>_<Scenario>_<ExpectedBehavior>`
- 表驱动测试优先

### Go 测试
```go
func TestUserService_GetByID_NotFound_ReturnsError(t *testing.T) {
    svc := setupTestService(t)
    _, err := svc.GetByID(context.Background(), "nonexistent")
    if err == nil {
        t.Error("expected error for nonexistent user")
    }
}
```

## 安全约定
- 所有用户输入必须验证（Zod / go-playground/validator）
- JWT Token: RS256, access 15min + refresh 7d
- 密码: bcrypt, cost ≥ 12
- API 限流: 100 req/min per IP
- 密钥通过环境变量注入，不得出现在代码中

## Git 约定
- 分支: feature/ fix/ refactor/ docs/
- 提交: Conventional Commits 格式
- PR 合并前必须通过 CI（lint + test + build）
- 禁止 force push 到 main
