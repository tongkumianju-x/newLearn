const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('api', {
  ready: () => ipcRenderer.send('renderer-ready'),
  onStatus: (cb) => ipcRenderer.on('status', (_, p) => cb(p)),
  onResult: (cb) => ipcRenderer.on('result', (_, p) => cb(p)),
  onPreview: (cb) => ipcRenderer.on('preview', (_, p) => cb(p)),
  onOcrText: (cb) => ipcRenderer.on('ocr-text', (_, p) => cb(p)),
  onOcrStats: (cb) => ipcRenderer.on('ocr-stats', (_, p) => cb(p)),
  onOpacityChanged: (cb) => ipcRenderer.on('opacity-changed', (_, p) => cb(p)),
  onQuitPrompt: (cb) => ipcRenderer.on('quit-prompt', (_, p) => cb(p)),
  onQuitPromptCancel: (cb) => ipcRenderer.on('quit-prompt-cancel', (_, p) => cb(p)),
  hide: () => ipcRenderer.send('window-hide'),
  close: () => ipcRenderer.send('window-close'),
  manualCapture: () => ipcRenderer.send('manual-capture'),
  openSettings: () => ipcRenderer.send('open-settings'),
  openScreenPrefs: () => ipcRenderer.send('open-screen-prefs'),
  copy: (text) => ipcRenderer.send('copy-text', text),
  reloadHotkey: () => ipcRenderer.send('reload-hotkey'),
  storeGet: (k) => ipcRenderer.invoke('store-get', k),
  storeSet: (k, v) => ipcRenderer.invoke('store-set', k, v),
  storeAll: () => ipcRenderer.invoke('store-all')
});
