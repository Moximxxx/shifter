#!/bin/bash
# Claude Code PreCommit Hook — 提交前检查
set -euo pipefail

echo "🔍 提交前检查..."

# 检查是否有 .env 文件被暂存
if git diff --cached --name-only | grep -q '\.env$'; then
  echo "🛑 阻止提交: .env 文件不能提交到仓库"
  exit 1
fi

# 检查密钥泄露
STAGED_FILES=$(git diff --cached --name-only)
if echo "$STAGED_FILES" | grep -qE '\.(ts|tsx|js|jsx|go|py|sh)$'; then
  echo "  → 检查密钥泄露..."
  if git diff --cached | grep -qE '(sk-[a-zA-Z0-9]{20,})|(ghp_[a-zA-Z0-9]{36})|(AKIA[A-Z0-9]{16})'; then
    echo "🛑 检测到可能的密钥泄露，请检查后再提交"
    exit 1
  fi
fi

# 运行 lint
if echo "$STAGED_FILES" | grep -qE '\.(ts|tsx)$'; then
  echo "  → 运行 ESLint..."
  npx eslint $STAGED_FILES --quiet 2>&1 || echo "⚠ ESLint 有警告"
fi

echo "✅ 提交前检查通过"
exit 0
