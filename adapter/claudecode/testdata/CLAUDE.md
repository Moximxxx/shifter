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
