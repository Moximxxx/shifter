---
name: security-scanner
description: 在代码变更时自动扫描安全漏洞，包括密钥泄露、SQL 注入、XSS、不安全的依赖项。
触发短语: 安全检查、漏洞扫描、代码审计、密钥检测、依赖项安全

---

## 目标
在代码变更时自动进行安全扫描，防止安全漏洞引入代码库。

## 执行步骤

### Step 1: 密钥泄露扫描
检查代码中是否包含：
- API 密钥 (OpenAI, GitHub, AWS, Stripe, Slack)
- 私钥 (PEM, SSH private key)
- 数据库连接字符串
- JWT secrets
- 密码硬编码

```bash
bash scripts/scan-secrets.sh
```

### Step 2: SQL 注入检查
检查是否使用了不安全的数据库查询：
- 字符串拼接 SQL
- 未参数化的查询
- 动态表名/列名

```bash
bash scripts/check-sql-injection.sh
```

### Step 3: XSS 防护检查
检查前端代码中的 XSS 风险：
- dangerouslySetInnerHTML 使用
- innerHTML 直接赋值
- 未净化的用户输入渲染

### Step 4: 依赖项安全检查
```bash
npm audit --audit-level=high
go list -m all | nancy sleuth
```

## 检查清单
- [ ] 无新增密钥泄露
- [ ] 数据库查询使用参数化
- [ ] 用户输入经过净化
- [ ] 依赖项无已知高危漏洞
- [ ] 认证逻辑正确

## 输出格式
```
## 安全扫描报告

### 🔴 发现的问题（必须修复）
| 文件 | 行号 | 问题类型 | 描述 |
|------|------|----------|------|
| xxx  | xxx  | xxx      | xxx  |

### 🟢 检查通过
- 密钥泄露扫描: ✅ 通过
- SQL 注入检查: ✅ 通过
- XSS 防护检查: ✅ 通过
- 依赖项安全: ✅ 通过
```
