#!/bin/bash
# 拦截危险命令
set -euo pipefail

cmd="$1"

# 阻止模式
BLOCKED=(
  'rm -rf /'
  'rm -rf --no-preserve-root'
  'git push --force origin main'
  'git push --force origin master'
  'DROP DATABASE'
  'DROP TABLE.*CASCADE'
  'shutdown'
  ':(){ :|:& };:'  # fork bomb
)

for pattern in "${BLOCKED[@]}"; do
  if echo "$cmd" | grep -qE "$pattern"; then
    echo "🛑 拦截：危险命令被阻止 - $cmd"
    exit 1
  fi
done

# 警告模式
WARNED=(
  'rm -rf \.'
  'chmod -R 777'
  'git reset --hard'
  'git clean -fd'
)

for pattern in "${WARNED[@]}"; do
  if echo "$cmd" | grep -qE "$pattern"; then
    echo "⚠ 警告：命令可能危险 - $cmd"
  fi
done

exit 0
