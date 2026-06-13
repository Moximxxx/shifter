#!/bin/bash
# 编辑后验证：检查代码格式和基本语法
set -euo pipefail

echo "🔍 编辑后验证..."

# 如果有 TypeScript 文件被修改，运行类型检查
if git diff --name-only | grep -q '\.tsx\?$'; then
  echo "  → 运行 TypeScript 类型检查..."
  npx tsc --noEmit 2>&1 || echo "⚠ 类型检查有警告"
fi

# 如果有 Go 文件被修改，运行 vet
if git diff --name-only | grep -q '\.go$'; then
  echo "  → 运行 go vet..."
  go vet ./... 2>&1 || echo "⚠ go vet 有警告"
fi

echo "✅ 编辑后验证完成"
exit 0
