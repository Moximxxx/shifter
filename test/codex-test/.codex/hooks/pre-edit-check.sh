#!/bin/bash
# 编辑前检查：验证目标文件在允许范围内
set -euo pipefail

echo "🔍 编辑前检查..."

# 禁止编辑的文件模式
FORBIDDEN=(
  '.env'
  '.env.*'
  '*.pem'
  '*.key'
  'credentials.json'
  'secrets.yaml'
)

for pattern in "${FORBIDDEN[@]}"; do
  if find . -name "$pattern" -newer /tmp/codex_session_start 2>/dev/null | grep -q .; then
    echo "🛑 禁止编辑敏感文件: $pattern"
    exit 1
  fi
done

echo "✅ 编辑前检查通过"
exit 0
