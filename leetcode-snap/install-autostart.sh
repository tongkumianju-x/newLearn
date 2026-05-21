#!/usr/bin/env bash
# 安装 LeetSnap 为登录自启项 - 一次执行，每次开机自动跑
# 用法：bash install-autostart.sh

set -e
cd "$(dirname "$0")"

PLIST_NAME="com.zhanglei.leetsnap"
PLIST_DEST="$HOME/Library/LaunchAgents/$PLIST_NAME.plist"

if [ ! -d "/Applications/LeetSnap.app" ]; then
  echo "错误：/Applications/LeetSnap.app 不存在"
  echo "请先运行 npm run build，然后把 dist/mac-arm64/LeetSnap.app 拷贝到 /Applications/"
  exit 1
fi

mkdir -p "$HOME/Library/LaunchAgents"
cp "$PLIST_NAME.plist" "$PLIST_DEST"

# 如果已加载过先卸载
launchctl unload "$PLIST_DEST" 2>/dev/null || true
launchctl load "$PLIST_DEST"

echo "✓ LeetSnap 已设为登录自启"
echo "  plist 路径: $PLIST_DEST"
echo "  下次开机会自动启动 /Applications/LeetSnap.app"
echo ""
echo "如需取消自启：launchctl unload $PLIST_DEST && rm $PLIST_DEST"
