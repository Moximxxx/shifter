---
name: security-audit
description: 安全审计 — 全面的代码安全审查，包括 OWASP Top 10 检查
allowed_tools: [Read, Glob, Grep, Bash, WebSearch]
---
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
