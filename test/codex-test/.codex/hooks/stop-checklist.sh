#!/bin/bash
# 会话停止检查清单
set -euo pipefail

echo "📋 会话停止检查清单..."

# 检查是否有未提交的变更
if ! git diff --quiet 2>/dev/null; then
  echo "⚠ 有未暂存的变更"
  git diff --stat
fi

# 检查是否有未推送的提交
UNPUSHED=$(git log @{u}.. 2>/dev/null | wc -l || echo 0)
if [ "$UNPUSHED" -gt 0 ]; then
  echo "⚠ 有 $UNPUSHED 个未推送的提交"
fi

# 检查测试是否通过
if [ -f "package.json" ]; then
  echo "  → 运行测试..."
  bun test --run 2>&1 | tail -5 || echo "⚠ 测试未全部通过"
fi

echo "✅ 检查清单完成"
exit 0
