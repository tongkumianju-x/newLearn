# LeetSnap — Mac 本地 LeetCode 截图速答插件

按 **fn + F2** 在 Mac 任意界面框选一道 LeetCode 题，应用会 OCR 识别题面，调用你配置的大模型 API 返回**多种**可直接提交的解法（每种带精炼标题、复杂度、思路），并把结果显示在右上角的浮动输出框里。

## 1. 功能一览

| 模块 | 实现 | 说明 |
|---|---|---|
| 全局热键 | `globalShortcut` | 截图 fn+F2、提高透明度 fn+↑、降低透明度 fn+↓ |
| 截图 | `screencapture -i -x` | 调系统原生交互截图 |
| OCR | `tesseract.js` (eng + chi_sim) | 完全离线，第一次运行时下载语言包 |
| 双引擎 | 本地题库 + 兼容 OpenAI/Anthropic 的 LLM | 默认优先 LLM，可在设置切换 |
| 多解法 | LLM 返回 2~4 种递进解法 | 每种带 ≤12 字标题、复杂度、Go/Python/C++ 等 AC 代码 |
| 输出框 | 无边框置顶 BrowserWindow | 暗色风格，可调透明度（25%~100%），可关闭 |
| 系统托盘 | `Tray` | 状态栏常驻，提供截图、显示/隐藏、设置、退出 |

## 2. 快速开始

```bash
git clone <repo>
cd leetcode-snap
npm install
./run.sh        # 或 npm start
```

> 第一次启动时 macOS 会请求"屏幕录制"权限：
> **系统设置 → 隐私与安全性 → 屏幕录制 → 勾选 Electron** 然后**重启应用**（macOS TCC 权限对运行中进程不生效）。

## 3. 配置 LLM API Key

**密钥不存放在仓库代码里**，仅保存到本地：

```
~/Library/Application Support/leetcode-snap/config.json
```

首次启动后，点托盘 **LS** 图标 → **设置…**：

1. 点击预设按钮（DeepSeek / 智谱 GLM / Moonshot 等）一键填充 Endpoint + Model
2. 在 **API Key** 一栏粘贴你的 Key
3. 点 **测试连接** 验证
4. 保存并应用

| 推荐服务商 | 价格/特点 |
|---|---|
| DeepSeek | ¥1/百万 token，写代码强 |
| 智谱 GLM-4-Flash | 免费 |
| Moonshot Kimi | 国内速度快 |
| OpenAI gpt-4o-mini | 需海外网络 |
| Anthropic Claude | 协议不同，已支持 |

> **绝不要**把真实 Key 写到代码、commit、issue 评论或截图里。
> 仓库根目录已配置 `.gitignore` 排除 `.env`、`config.local.json`、`*.key`、`secrets/` 等敏感路径。

## 4. 快捷键

| 快捷键 | 行为 |
|---|---|
| **fn + F2** | 截图框选 → OCR → 解题 |
| **fn + ↑** | 输出窗透明度 +10% |
| **fn + ↓** | 输出窗透明度 -10%（最低 25%） |
| 标题栏 × / 隐藏 | 关闭输出窗，托盘进程仍驻留 |
| 托盘 → 退出 | 完全退出 |

可在设置面板把任一快捷键改成 `CommandOrControl+Shift+L` 这种组合键。

## 5. 双引擎策略

```
截图 → OCR 文本
        │
        ▼
   ┌──────────────┐
   │  本地题库匹配 │  关键词加权 ≥ 4 分 → 直接返回
   └──────┬───────┘
          │ 未命中
          ▼
   ┌──────────────┐
   │  外接 LLM    │  按 system prompt 输出多解法 JSON
   └──────────────┘
```

可在设置切换为"优先 LLM"。

## 6. 本地题库扩展

把新题放到 `src/solver/bank/<id>-<slug>.json`：

```json
{
  "id": 15,
  "titleEn": "3Sum",
  "titleCn": "三数之和",
  "keywords": ["3sum", "三数之和", "triplets that sum to zero"],
  "complexity": "时间 O(n²)，空间 O(1)",
  "solutions": {
    "python": "...",
    "go": "..."
  },
  "explanation": "排序 + 双指针。"
}
```

无需重启即可在下次冷启动时生效。

## 7. 目录结构

```
leetcode-snap/
├── package.json
├── run.sh / restart.sh             # 一键启动 / 重启
├── .gitignore                      # 排除密钥与缓存
├── .env.example                    # 配置项参考（不含真实 Key）
├── build/entitlements.mac.plist
└── src/
    ├── main.js                     # Electron 主进程：热键、截图、OCR、IPC、托盘、透明度
    ├── preload.js                  # 安全桥接 contextBridge
    ├── renderer/
    │   ├── index.html              # 输出框 UI（多 Tab 切换、透明度浮层）
    │   └── settings.html           # 设置面板（含 8 服务商一键预设、测试连接）
    └── solver/
        ├── local.js                # 本地题库匹配
        ├── llm.js                  # OpenAI/Anthropic 协议适配，输出多解法
        └── bank/                   # JSON 题库
```

## 8. 安全说明

- API Key 仅存 `~/Library/Application Support/leetcode-snap/config.json`，永不入库；
- LLM 调用走你配置的 endpoint，应用本身不会代理任何流量；
- 截图仅在本地处理，OCR 完成后立即删除临时 PNG；
- OCR 模型缓存（`*.traineddata`）通过 `.gitignore` 排除。

## 9. 已知限制

- 仅支持 macOS（依赖 `screencapture` 与 macOS TCC 权限模型）；
- 第一次按 fn+F2 时 macOS 才会弹"屏幕录制权限"请求，授权后**必须重启 LeetSnap**；
- 未授权时 `screencapture -i` 会静默降级为只截壁纸，本应用已加大小检测自动跳设置。
