// 打包入口：把 Tauri HTTP 插件的 fetch 挂到全局，供无打包器的页面使用。
import { fetch } from "@tauri-apps/plugin-http";

window.__TAURI__ = window.__TAURI__ || {};
window.__TAURI__.http = window.__TAURI__.http || {};
window.__TAURI__.http.fetch = fetch;
