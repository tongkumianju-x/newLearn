#!/usr/bin/env bash
# LeetSnap 启动脚本 — 在你自己的终端里跑
# 用法: ./run.sh

set -e
cd "$(dirname "$0")"

if [ ! -d "node_modules" ]; then
  echo "首次运行，安装依赖..."
  npm install --no-audit --no-fund
fi

EXEC="$(pwd)/node_modules/electron/dist/Electron.app/Contents/MacOS/Electron"

if [ ! -x "$EXEC" ]; then
  echo "Electron 二进制未找到: $EXEC"
  echo "请重新运行 npm install"
  exit 1
fi

unset ELECTRON_RUN_AS_NODE

echo ""
echo "=========================================="
echo "  LeetSnap 启动中..."
echo "  快捷键:  fn + F2"
echo "  停止:    Ctrl + C 或关闭此终端"
echo "=========================================="
echo ""

exec "$EXEC" "$(pwd)" --no-sandbox
