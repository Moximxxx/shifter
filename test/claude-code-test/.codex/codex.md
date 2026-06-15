# Project: Fullstack SaaS Platform

## Agent Workflow (多代理协作流程)

### 代理体系
本项目配置了 7 个专用子代理，由 coordinator 主代理统一调度：

```
用户任务 → coordinator (主代理)
  ├── analyzer → 分析任务、对比方案、输出建议
  ├── task-executor → 按合同执行代码修改
  ├── code-reviewer → 审查代码质量和安全性
  ├── test-writer → 生成和补充测试
  ├── builder → 构建和验证
  ├── retro → 复盘和改进
  └── crash-doctor → 故障诊断和修复
```

### 流程规范
1. 所有代码修改必须先通过 analyzer 分析
2. analyzer 输出包含: files_to_modify、constraints、verification、suggested_skills
3. task-executor 只修改 analyzer 指定的文件
4. 代码修改后必须通过 code-reviewer 审查
5. 审查不通过 → 自动修复循环 (最多3次)
6. 所有阶段通过后 → builder 构建验证 → retro 复盘


---

## Custom Commands (ported from another agent)

> **Note:** Codex CLI does not have native slash commands. These are provided as instruction patterns you can reference.

### /deploy

*执行完整的部署前检查和部署流程*

执行部署流程，目标环境: $ARGUMENTS (默认 staging)。

## 部署前检查
1. 运行所有测试: `bun test && go test ./...`
2. 类型检查: `npx tsc --noEmit`
3. Lint: `npx eslint . --quiet`
4. 构建: `bun run build && go build ./...`

## 部署步骤
1. 构建 Docker 镜像: `docker build -t app:$ARGUMENTS .`
2. 推送到镜像仓库: `docker push registry.example.com/app:$ARGUMENTS`
3. 更新 K8s 部署: `kubectl set image deployment/app app=registry.example.com/app:$ARGUMENTS`
4. 等待就绪: `kubectl rollout status deployment/app`
5. 健康检查: `curl -f https://$ARGUMENTS.example.com/health`

任何步骤失败立即停止并报告。

---

### /review

*审查当前变更的代码质量和安全性*

你正在审查代码变更。根据 $ARGUMENTS 聚焦审查范围。

## 审查清单
1. 是否有明显的 bug？
2. 代码是否遵循项目规范？
3. 边界条件是否处理？
4. 是否有测试覆盖变更？
5. 是否可以简化？

## 输出格式
按照 code-reviewer 代理的输出格式，提供包含文件引用的总结。

---



---

## Custom Skills (ported from another agent)

> **Note:** Codex CLI has experimental skill support (features.skills=true). Skills listed below can be installed to `~/.codex/skills/`.

### fullstack-dev

*全栈开发工作流 — 从需求分析到代码实现到部署的完整流程*

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

---

### security-audit

*安全审计 — 全面的代码安全审查，包括 OWASP Top 10 检查*

执行全面的安全审计，覆盖 OWASP Top 10 和常见安全漏洞。

## 审计清单

### 1. 注入攻击 (Injection)
- [ ] SQL 查询使用参数化（检查所有数据库查询）
- [ ] 命令行参数未拼接到 shell 命令
- [ ] LDAP/OS 命令注入防护

### 2. 认证失效 (Broken Authentication)
- [ ] JWT 使用 RS256 算法
- [ ] 密码使用 bcrypt (cost ≥ 12)
- [ ] 会话超时设置合理
- [ ] 无硬编码凭证

### 3. 敏感数据泄露 (Sensitive Data Exposure)
- [ ] 密钥存储在环境变量中
- [ ] 日志不包含敏感信息
- [ ] HTTPS 强制启用
- [ ] 数据库连接使用 TLS

### 4. XML 外部实体 (XXE)
- [ ] XML 解析器禁用外部实体

### 5. 访问控制失效 (Broken Access Control)
- [ ] 每个端点验证用户权限
- [ ] API 限流已配置
- [ ] CORS 配置正确

### 6. 安全配置错误 (Security Misconfiguration)
- [ ] 调试模式在生产环境关闭
- [ ] 默认密码已更改
- [ ] 错误信息不泄露内部细节

### Execute
```bash
bash scripts/scan-secrets.sh
bash scripts/check-dependencies.sh
```

### 审计报告格式
```
## Security Audit Report
- 审计日期: [date]
- 审计范围: [scope]

### 发现的问题
| 严重程度 | 类型 | 位置 | 描述 | 修复建议 |
|----------|------|------|------|----------|

### 通过检查
- [通过的项目列表]

### 总结
- 总问题数: X
- 严重: X, 高危: X, 中危: X, 低危: X
- 安全评分: X/100
```

---

