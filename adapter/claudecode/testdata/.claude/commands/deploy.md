---
description: 执行完整的部署前检查和部署流程
argument-hint: [environment]
---

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
