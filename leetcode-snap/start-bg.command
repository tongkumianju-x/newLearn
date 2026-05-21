#!/usr/bin/env bash
# LeetSnap 双击后台启动 - 关掉终端不会停
# 用法：在 Finder 双击 start-bg.command

cd "$(dirname "$0")"

# 已在跑就不重复启动
if pgrep -f "leetcode-snap/node_modules/electron/dist/Electron.app/Contents/MacOS/Electron" > /dev/null; then
  osascript -e 'display notification "LeetSnap 已在运行" with title "LeetSnap"'
  exit 0
fi

if [ ! -d "node_modules" ]; then
  osascript -e 'display dialog "请先在终端运行 npm install"'
  exit 1
fi

EXEC="$(pwd)/node_modules/electron/dist/Electron.app/Contents/MacOS/Electron"
LOG="$HOME/Library/Logs/leetsnap.log"

# nohup + setsid + 重定向，让进程脱离终端 controlling tty
unset ELECTRON_RUN_AS_NODE
nohup "$EXEC" "$(pwd)" --no-sandbox > "$LOG" 2>&1 < /dev/null &
disown

osascript -e 'display notification "LeetSnap 已在后台启动，菜单栏可见 LS 图标" with title "LeetSnap"'
exit 0
