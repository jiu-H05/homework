// net.js — 统一的 fetch 入口。
// 优先使用 Tauri HTTP 插件（请求由 Rust 核心侧发起，可绕过 WebView2 对
// loopback / 跨域 / 混合内容的限制）；在普通浏览器中回退到原生 fetch。
const tauri = typeof window !== "undefined" ? window.__TAURI__ : null;

// 仅当处于真实 Tauri 运行时（由核心注入 window.__TAURI_INTERNALS__）才使用插件 fetch；
// 普通浏览器里 vendor/http.global.js 虽会定义 window.__TAURI__.http.fetch，但缺少
// __TAURI_INTERNALS__，不能使用，应回退原生 fetch。
const pluginFetch =
  tauri &&
  typeof window !== "undefined" &&
  window.__TAURI_INTERNALS__ &&
  tauri.http &&
  typeof tauri.http.fetch === "function"
    ? tauri.http.fetch.bind(tauri.http)
    : null;

export const lmsFetch = pluginFetch || window.fetch.bind(window);
export const usingPlugin = !!pluginFetch;

// tauriInvoke：调用 Rust 侧注册的 command。仅在真实 Tauri 运行时可用；
// 普通浏览器下 reject，由调用方决定回退。走 __TAURI_INTERNALS__ 与 http 插件同一通道。
export function tauriInvoke(cmd, args) {
  if (
    typeof window !== "undefined" &&
    window.__TAURI_INTERNALS__ &&
    typeof window.__TAURI_INTERNALS__.invoke === "function"
  ) {
    return window.__TAURI_INTERNALS__.invoke(cmd, args || {});
  }
  return Promise.reject(new Error("not running inside Tauri"));
}
