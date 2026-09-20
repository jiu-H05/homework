// net.js — 统一的 fetch 入口。
// 优先使用 Tauri HTTP 插件（请求由 Rust 核心侧发起，可绕过 WebView2 对
// loopback / 跨域 / 混合内容的限制）；在普通浏览器中回退到原生 fetch。
const tauri = typeof window !== "undefined" ? window.__TAURI__ : null;

const pluginFetch = tauri && tauri.http && typeof tauri.http.fetch === "function"
  ? tauri.http.fetch.bind(tauri.http)
  : null;

export const lmsFetch = pluginFetch || window.fetch.bind(window);
export const usingPlugin = !!pluginFetch;
