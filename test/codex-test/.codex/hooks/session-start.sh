#!/bin/bash
# Codex SessionStart Hook - 加载项目上下文
set -euo pipefail

echo "📋 加载项目上下文..."
echo "  Branch: $(git branch --show-current 2>/dev/null || echo 'N/A')"
echo "  Last commit: $(git log -1 --oneline 2>/dev/null || echo 'N/A')"
echo "  Node: $(node -v 2>/dev/null || echo 'N/A')"
echo "  Go: $(go version 2>/dev/null || echo 'N/A')"

# 检查环境变量
for var in DATABASE_URL GITHUB_TOKEN; do
  if [ -z "${!var:-}" ]; then
    echo "⚠ 警告: $var 未设置"
  fi
done

exit 0
