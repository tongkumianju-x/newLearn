#!/usr/bin/env bash
# LeetSnap 重启脚本：彻底杀掉旧进程后干净启动
# 用法: ./restart.sh

set -e
cd "$(dirname "$0")"

echo "[1/3] 杀掉所有 LeetSnap 相关进程..."
pkill -9 -f "leetcode-snap/node_modules/electron" 2>/dev/null || true
sleep 1

REMAINING=$(pgrep -f "leetcode-snap/node_modules/electron" | wc -l | tr -d ' ')
if [ "$REMAINING" != "0" ]; then
  echo "  仍有 $REMAINING 个孤儿 helper 进程（无害，会被新主进程接管）"
fi

echo "[2/3] 检查 Electron 二进制..."
EXEC="$(pwd)/node_modules/electron/dist/Electron.app/Contents/MacOS/Electron"
if [ ! -x "$EXEC" ]; then
  echo "  Electron 未安装，运行 npm install ..."
  npm install --no-audit --no-fund
fi

echo "[3/3] 启动 LeetSnap（带权限自检）..."
unset ELECTRON_RUN_AS_NODE

echo ""
echo "=========================================="
echo "  LeetSnap 已启动"
echo "  快捷键:  fn + F2"
echo "  停止:    Ctrl + C 或关闭此终端"
echo ""
echo "  如果首次按 fn+F2 时弹出权限请求，请到："
echo "  系统设置 → 隐私与安全性 → 屏幕录制"
echo "  勾选 Electron，然后再次运行 ./restart.sh"
echo "=========================================="
echo ""

exec "$EXEC" "$(pwd)" --no-sandbox
