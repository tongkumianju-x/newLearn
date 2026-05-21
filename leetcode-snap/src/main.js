const { app, BrowserWindow, globalShortcut, ipcMain, screen, Tray, Menu, nativeImage, clipboard, Notification, systemPreferences, shell, dialog, desktopCapturer } = require('electron');
const path = require('path');
const { spawn } = require('child_process');
const fs = require('fs');
const os = require('os');
const Store = require('electron-store');

const store = new Store({
  defaults: {
    hotkey: 'F2',
    hotkeyOpacityUp: 'Alt+1',
    hotkeyOpacityDown: 'Alt+2',
    hotkeyToggle: 'Alt+3',
    hotkeyQuit: 'Alt+Q',
    opacity: 1.0,
    llmProvider: 'openai',
    llmApiKey: '',
    llmEndpoint: 'https://api.deepseek.com/v1/chat/completions',
    llmModel: 'deepseek-chat',
    llmTimeoutMs: 300000,
    preferLocalSolver: false,
    language: 'go'
  }
});

const OPACITY_MIN = 0.25;
const OPACITY_MAX = 1.0;
const OPACITY_STEP = 0.1;
const QUIT_CONFIRM_MS = 1500;

let quitPending = false;
let quitTimer = null;

let outputWin = null;
let settingsWin = null;
let tray = null;
let isCapturing = false;
let pendingMessages = [];
let rendererReady = false;

function createOutputWindow() {
  if (outputWin && !outputWin.isDestroyed()) return outputWin;

  const display = screen.getPrimaryDisplay();
  const { width } = display.workArea;

  outputWin = new BrowserWindow({
    width: 560,
    height: 720,
    x: width - 580,
    y: 60,
    title: 'LeetSnap',
    frame: false,
    transparent: false,
    backgroundColor: '#1a1d23',
    resizable: true,
    alwaysOnTop: true,
    skipTaskbar: false,
    show: false,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
      backgroundThrottling: false
    }
  });

  rendererReady = false;
  outputWin.loadFile(path.join(__dirname, 'renderer', 'index.html'));

  const savedOpacity = clampOpacity(Number(store.get('opacity')) || 1.0);
  outputWin.setOpacity(savedOpacity);

  outputWin.on('close', (e) => {
    e.preventDefault();
    outputWin.hide();
  });

  outputWin.on('closed', () => {
    outputWin = null;
    rendererReady = false;
  });

  return outputWin;
}

function showOutput() {
  const win = createOutputWindow();
  if (!win.isVisible()) win.show();
  win.focus();
}

function hideOutput() {
  if (outputWin && !outputWin.isDestroyed() && outputWin.isVisible()) outputWin.hide();
}

function toggleOutput() {
  const win = createOutputWindow();
  if (win.isVisible()) {
    win.hide();
  } else {
    win.show();
    win.focus();
  }
}

function requestQuit() {
  if (quitPending) {
    clearTimeout(quitTimer);
    globalShortcut.unregisterAll();
    app.exit(0);
    return;
  }
  quitPending = true;
  showOutput();
  safeSend('quit-prompt', { message: '再按一次退出快捷键确认退出 LeetSnap' });
  try {
    new Notification({ title: 'LeetSnap', body: '再按一次退出快捷键确认退出' }).show();
  } catch (_) {}
  quitTimer = setTimeout(() => {
    quitPending = false;
    safeSend('quit-prompt-cancel', {});
  }, QUIT_CONFIRM_MS);
}

function clampOpacity(v) {
  if (isNaN(v)) v = 1.0;
  return Math.max(OPACITY_MIN, Math.min(OPACITY_MAX, Number(v.toFixed(2))));
}

function adjustOpacity(delta) {
  const cur = Number(store.get('opacity')) || 1.0;
  const next = clampOpacity(cur + delta);
  store.set('opacity', next);
  if (outputWin && !outputWin.isDestroyed()) {
    if (!outputWin.isVisible()) outputWin.show();
    outputWin.setOpacity(next);
  }
  safeSend('opacity-changed', { value: next, percent: Math.round(next * 100) });
  rebuildTrayMenu();
  return next;
}

function flushPending() {
  if (!rendererReady || !outputWin || outputWin.isDestroyed()) return;
  const wc = outputWin.webContents;
  while (pendingMessages.length) {
    const { channel, payload } = pendingMessages.shift();
    try { wc.send(channel, payload); } catch (_) {}
  }
}

function safeSend(channel, payload) {
  if (!outputWin || outputWin.isDestroyed()) {
    pendingMessages.push({ channel, payload });
    createOutputWindow();
    return;
  }
  if (!rendererReady) {
    pendingMessages.push({ channel, payload });
    return;
  }
  const wc = outputWin.webContents;
  if (!wc || wc.isDestroyed()) return;
  try { wc.send(channel, payload); } catch (e) {
    pendingMessages.push({ channel, payload });
  }
}

function checkScreenPermission() {
  if (process.platform !== 'darwin') return 'granted';
  try {
    return systemPreferences.getMediaAccessStatus('screen');
  } catch (_) { return 'unknown'; }
}

async function probeCanCaptureWindows() {
  if (process.platform !== 'darwin') return true;
  try {
    const sources = await desktopCapturer.getSources({ types: ['window'], thumbnailSize: { width: 0, height: 0 } });
    return sources.length > 0;
  } catch (_) {
    return false;
  }
}

function openScreenRecordingPrefs() {
  shell.openExternal('x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture');
}

function captureScreenshot() {
  return new Promise((resolve, reject) => {
    const tmp = path.join(os.tmpdir(), `leetsnap-${Date.now()}.png`);
    const proc = spawn('/usr/sbin/screencapture', ['-i', '-x', tmp]);
    let cancelled = false;
    proc.on('close', (code) => {
      if (code !== 0) {
        return reject(new Error(`screencapture 退出码 ${code}`));
      }
      if (!fs.existsSync(tmp)) {
        return reject(new Error('用户取消截图'));
      }
      const stat = fs.statSync(tmp);
      if (stat.size < 200) {
        fs.unlink(tmp, () => {});
        return reject(new Error('截图失败：文件过小，可能未授予屏幕录制权限'));
      }
      const buf = fs.readFileSync(tmp);
      fs.unlink(tmp, () => {});
      resolve(buf);
    });
    proc.on('error', (err) => reject(new Error('调起截图进程失败：' + err.message)));
  });
}

async function runOcr(buffer) {
  const { createWorker } = require('tesseract.js');
  let worker;
  try {
    worker = await createWorker(['eng', 'chi_sim']);
    const { data } = await worker.recognize(buffer);
    return data.text;
  } finally {
    if (worker) {
      try { await worker.terminate(); } catch (_) {}
    }
  }
}

// 清洗 OCR 文本：去除空格噪声、压成单行完整文本喂给 LLM
function cleanOcrText(raw) {
  if (!raw) return '';
  let s = raw;

  // 1. 去 BOM / 零宽字符 / 软连字符
  s = s.replace(/[\u200B-\u200F\u202A-\u202E\u2060-\u206F\uFEFF\u00AD]/g, '');

  // 2. 全角空格 → 半角
  s = s.replace(/\u3000/g, ' ');

  // 3. 英文连字符回行 "abc-\ndef" → "abcdef"
  s = s.replace(/-\n/g, '');

  // 4. 所有换行 → 空格（LLM 只需要完整文本，不依赖换行结构）
  s = s.replace(/[\r\n]+/g, ' ');

  // 5. CJK 字符集合：CJK 统一汉字 + 假名 + CJK 符号标点（。，：；！？「」『』《》〈〉【】等）+ 全角 ASCII（含 ， 。 ！ ？ 等）
  //    \u3000-\u303F: CJK 符号和标点
  //    \uFF00-\uFFEF: 半/全角 ASCII（含全角逗号 ， 全角句号 。 等）
  const CJK = '\u4E00-\u9FFF\u3040-\u309F\u30A0-\u30FF\u3000-\u303F\uFF00-\uFFEF';

  // 多次循环清洗，直到收敛（每轮：去 CJK 间空格 → 中英间留单空格 → 合并多空格）
  const cjkSpaceRe = new RegExp(`([${CJK}])\\s+([${CJK}])`, 'g');
  const cjkToAsciiRe = new RegExp(`([${CJK}])\\s*([A-Za-z0-9])`, 'g');
  const asciiToCjkRe = new RegExp(`([A-Za-z0-9])\\s*([${CJK}])`, 'g');

  for (let i = 0; i < 5; i++) {
    const before = s;

    // 6. CJK 字符（含中文标点）之间的空格全部删除
    let prev;
    do { prev = s; s = s.replace(cjkSpaceRe, '$1$2'); } while (s !== prev);

    // 7. 中文与英文/数字之间保留单空格
    s = s.replace(cjkToAsciiRe, '$1 $2');
    s = s.replace(asciiToCjkRe, '$1 $2');

    // 8. 多个空格 → 单空格
    s = s.replace(/[ \t]{2,}/g, ' ');

    if (s === before) break;
  }

  return s.trim();
}

const localSolver = require('./solver/local');
const llmSolver = require('./solver/llm');

async function solve(text) {
  const preferLocal = store.get('preferLocalSolver');
  let result = null;

  if (preferLocal) {
    result = localSolver.match(text, store.get('language'));
    if (result) { result.engine = 'local'; return result; }
  }

  const apiKey = store.get('llmApiKey');
  if (apiKey) {
    try {
      const llmRes = await llmSolver.solve(text, {
        provider: store.get('llmProvider'),
        endpoint: store.get('llmEndpoint'),
        model: store.get('llmModel'),
        apiKey,
        language: store.get('language'),
        timeoutMs: Number(store.get('llmTimeoutMs')) || 300000,
        onProgress: ({ elapsed, remain, timeoutSec }) => {
          safeSend('status', {
            phase: 'solving',
            message: `正在调用 LLM 生成 AC 代码…已等待 ${elapsed}s / 上限 ${timeoutSec}s（剩余 ${remain}s）`
          });
        }
      });
      llmRes.engine = 'llm';
      return llmRes;
    } catch (e) {
      if (!preferLocal) {
        result = localSolver.match(text, store.get('language'));
        if (result) { result.engine = 'local'; return result; }
      }
      throw e;
    }
  }

  if (!preferLocal) {
    result = localSolver.match(text, store.get('language'));
    if (result) { result.engine = 'local'; return result; }
  }

  return {
    engine: 'none',
    title: '未配置 LLM API Key',
    methods: [{
      name: '请先填 API Key',
      complexity: '-',
      code: '# 请到设置中填写 LLM API Key 才能使用在线解题：\n#\n#   托盘菜单 LS → 设置… → API Key 处填入\n#\n# 推荐配置（任选其一）：\n#\n# [DeepSeek]    Endpoint: https://api.deepseek.com/v1/chat/completions   Model: deepseek-chat\n# [智谱 GLM]   Endpoint: https://open.bigmodel.cn/api/paas/v4/chat/completions   Model: glm-4-flash\n# [Moonshot]   Endpoint: https://api.moonshot.cn/v1/chat/completions   Model: moonshot-v1-8k\n# [OpenAI]     Endpoint: https://api.openai.com/v1/chat/completions   Model: gpt-4o-mini\n#\n# 填好后保存并按 fn+F2 重试即可。',
      explanation: 'OCR 文本（已识别成功，等待 API 解题）：\n' + text.slice(0, 400)
    }],
    language: store.get('language')
  };
}

async function handleCapture() {
  if (isCapturing) {
    new Notification({ title: 'LeetSnap', body: '正在处理上一次截图，请稍候…' }).show();
    return;
  }
  isCapturing = true;

  const perm = checkScreenPermission();
  const canCapture = await probeCanCaptureWindows();
  if ((perm === 'denied' || perm === 'restricted') && !canCapture) {
    isCapturing = false;
    showOutput();
    safeSend('status', { phase: 'error', message: '未授予屏幕录制权限：将打开系统设置→隐私→屏幕录制，请勾选 Electron / LeetSnap 后重启本应用。' });
    new Notification({ title: 'LeetSnap', body: '请在系统设置 → 隐私与安全性 → 屏幕录制 中勾选 LeetSnap，然后重启' }).show();
    openScreenRecordingPrefs();
    return;
  }

  const wasVisible = outputWin && !outputWin.isDestroyed() && outputWin.isVisible();
  if (wasVisible) {
    outputWin.hide();
    await new Promise(r => setTimeout(r, 250));
  }

  try {
    const buf = await captureScreenshot();

    showOutput();
    safeSend('status', { phase: 'ocr', message: '截图完成，正在 OCR 识别…' });
    safeSend('preview', `data:image/png;base64,${buf.toString('base64')}`);

    const rawText = await runOcr(buf);
    const text = cleanOcrText(rawText);
    const rawLen = (rawText || '').length;
    const cleanLen = text.length;
    const reduced = rawLen > 0 ? Math.round((1 - cleanLen / rawLen) * 100) : 0;
    safeSend('ocr-text', text || '(未识别到任何文字)');
    safeSend('ocr-stats', { rawLen, cleanLen, reduced });

    if (!text || text.length < 5) {
      throw new Error('OCR 结果为空，可能截图内容无文字或权限不足只截到桌面');
    }

    safeSend('status', { phase: 'solving', message: `正在生成 AC 代码…（已清洗 OCR：${rawLen} → ${cleanLen} 字符，压缩 ${reduced}%）` });
    const result = await solve(text);

    safeSend('result', result);
    safeSend('status', { phase: 'done', message: `完成（来源：${result.engine}）` });
  } catch (err) {
    showOutput();
    const msg = err.message || String(err);
    safeSend('status', { phase: 'error', message: msg });
    new Notification({ title: 'LeetSnap 出错', body: msg }).show();
    if (msg.includes('权限') || msg.includes('文件过小')) {
      openScreenRecordingPrefs();
    }
  } finally {
    isCapturing = false;
  }
}

function registerHotkey() {
  globalShortcut.unregisterAll();

  const captureKey = store.get('hotkey');
  const okCap = globalShortcut.register(captureKey, handleCapture);
  console.log(`[hotkey] 截图 ${captureKey}: ${okCap ? '✓' : '✗ 注册失败'}`);
  if (!okCap) {
    new Notification({ title: 'LeetSnap', body: `截图快捷键 ${captureKey} 注册失败，可能被占用` }).show();
  }

  const upKey = store.get('hotkeyOpacityUp') || 'Alt+1';
  const downKey = store.get('hotkeyOpacityDown') || 'Alt+2';
  const okUp = globalShortcut.register(upKey, () => adjustOpacity(+OPACITY_STEP));
  const okDown = globalShortcut.register(downKey, () => adjustOpacity(-OPACITY_STEP));
  console.log(`[hotkey] 提高透明度 ${upKey}: ${okUp ? '✓' : '✗'}    降低透明度 ${downKey}: ${okDown ? '✓' : '✗'}`);

  const toggleKey = store.get('hotkeyToggle') || 'Alt+3';
  const okToggle = globalShortcut.register(toggleKey, toggleOutput);
  console.log(`[hotkey] 显示/隐藏切换 ${toggleKey}: ${okToggle ? '✓' : '✗'}`);

  const quitKey = store.get('hotkeyQuit') || 'Alt+Q';
  const okQuit = globalShortcut.register(quitKey, requestQuit);
  console.log(`[hotkey] 退出程序 ${quitKey}: ${okQuit ? '✓' : '✗'}`);

  if (!okUp || !okDown || !okToggle || !okQuit) {
    new Notification({
      title: 'LeetSnap 部分快捷键注册失败',
      body: `${!okUp ? upKey + ' ' : ''}${!okDown ? downKey + ' ' : ''}${!okToggle ? toggleKey + ' ' : ''}${!okQuit ? quitKey + ' ' : ''}被占用，请到设置改成其它组合键。`
    }).show();
  }

  console.log(`[hotkey] 注意：macOS 不支持单独的 fn+数字键作为全局快捷键，fn 是 Apple 私有键。请使用带 Cmd/Alt/Shift 修饰的组合键。`);
}

function createTray() {
  const icon = nativeImage.createFromPath(path.join(__dirname, '..', 'assets', 'tray.png'));
  tray = new Tray(icon.isEmpty() ? nativeImage.createEmpty() : icon);
  tray.setTitle('LS');
  rebuildTrayMenu();
  tray.setToolTip('LeetSnap — fn+F2 截图解题');
}

async function rebuildTrayMenu() {
  if (!tray) return;
  const perm = checkScreenPermission();
  const canCapture = await probeCanCaptureWindows();
  let permLabel;
  if (canCapture) permLabel = '✓ 已授予屏幕录制权限';
  else if (perm === 'denied') permLabel = '✗ 未授权（点击打开系统设置）';
  else permLabel = '? 屏幕录制权限：' + perm + '（点击打开系统设置）';

  const opacity = Number(store.get('opacity')) || 1;
  const captureKey = store.get('hotkey') || 'F2';
  const upKey = store.get('hotkeyOpacityUp') || 'Alt+1';
  const downKey = store.get('hotkeyOpacityDown') || 'Alt+2';
  const toggleKey = store.get('hotkeyToggle') || 'Alt+3';
  const quitKey = store.get('hotkeyQuit') || 'Alt+Q';

  const menu = Menu.buildFromTemplate([
    { label: `立即截图解题  (${captureKey})`, click: handleCapture },
    { label: `显示 / 隐藏输出窗  (${toggleKey})`, click: toggleOutput },
    { label: '仅显示输出窗', click: showOutput },
    { label: '仅隐藏输出窗', click: hideOutput },
    { type: 'separator' },
    { label: `透明度：${Math.round(opacity * 100)}%`, enabled: false },
    { label: `提高透明度（更不透明）  ${upKey}`, click: () => adjustOpacity(+OPACITY_STEP) },
    { label: `降低透明度（更透明）    ${downKey}`, click: () => adjustOpacity(-OPACITY_STEP) },
    { label: '重置透明度 100%', click: () => { store.set('opacity', 1.0); if (outputWin && !outputWin.isDestroyed()) outputWin.setOpacity(1.0); rebuildTrayMenu(); } },
    { type: 'separator' },
    { label: permLabel, click: openScreenRecordingPrefs },
    { label: '设置…', click: () => createSettingsWindow() },
    { type: 'separator' },
    { label: `退出 LeetSnap  (${quitKey})`, click: () => { globalShortcut.unregisterAll(); app.exit(0); } }
  ]);
  tray.setContextMenu(menu);
}

function createSettingsWindow() {
  if (settingsWin && !settingsWin.isDestroyed()) {
    settingsWin.show();
    return;
  }
  settingsWin = new BrowserWindow({
    width: 520,
    height: 560,
    title: 'LeetSnap 设置',
    resizable: false,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true
    }
  });
  settingsWin.loadFile(path.join(__dirname, 'renderer', 'settings.html'));
}

ipcMain.handle('store-get', (_, key) => store.get(key));
ipcMain.handle('store-set', (_, key, val) => { store.set(key, val); return true; });
ipcMain.handle('store-all', () => store.store);
ipcMain.on('window-hide', () => hideOutput());
ipcMain.on('window-close', () => hideOutput());
ipcMain.on('reload-hotkey', () => registerHotkey());
ipcMain.on('open-settings', () => createSettingsWindow());
ipcMain.on('manual-capture', () => handleCapture());
ipcMain.on('copy-text', (_, text) => clipboard.writeText(text));
ipcMain.on('renderer-ready', () => {
  rendererReady = true;
  flushPending();
});
ipcMain.on('open-screen-prefs', () => openScreenRecordingPrefs());

app.whenReady().then(() => {
  if (process.platform === 'darwin') app.dock.hide();
  createOutputWindow();
  createTray();
  registerHotkey();

  setTimeout(async () => {
    const perm = checkScreenPermission();
    const canCapture = await probeCanCaptureWindows();
    if ((perm === 'denied' || perm === 'restricted') && !canCapture) {
      new Notification({
        title: 'LeetSnap 需要屏幕录制权限',
        body: '请在系统设置→隐私与安全性→屏幕录制中勾选 Electron / LeetSnap，否则只能截到桌面壁纸。'
      }).show();
      rebuildTrayMenu();
    } else {
      rebuildTrayMenu();
    }
  }, 1500);
});

app.on('will-quit', () => globalShortcut.unregisterAll());
app.on('window-all-closed', (e) => e.preventDefault());
