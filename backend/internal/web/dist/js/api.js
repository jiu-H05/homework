// api.js — 后端 REST API 客户端。
// 职责单一：封装请求、令牌注入、错误归一化；不包含任何视图逻辑。

import { lmsFetch } from "./net.js";

// BASE：在 Tauri 桌面壳内用显式私网 IP（绕过 WebView2 限制）；
// 在被 Go 后端以同一端口托管的浏览器中使用同源相对路径——页面从哪个私网 IP 打开，
// API 就请求哪个 IP，天然不走 localhost，且更换网络无需改代码。
// 必须以核心注入的 __TAURI_INTERNALS__ 作为 Tauri 运行时判据；
// vendor/http.global.js 在普通浏览器里也会定义 window.__TAURI__（HTTP 垫片），不能据此判断。
const inTauri = typeof window !== "undefined" && !!window.__TAURI_INTERNALS__;
const BASE = inTauri ? "http://10.45.171.101:8765" : "";
// API_BASE 供启动健康检查等场景复用：浏览器中为空（同源相对，跟随当前私网 IP），Tauri 中为显式地址。
export const API_BASE = BASE;
const TOKEN_KEY = "lms_token";

export class ApiError extends Error {
  constructor(message, status) {
    super(message);
    this.status = status;
  }
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}
export function setToken(token) {
  if (token) localStorage.setItem(TOKEN_KEY, token);
  else localStorage.removeItem(TOKEN_KEY);
}

async function request(method, path, body) {
  const headers = {};
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  if (body !== undefined) headers["Content-Type"] = "application/json";

  let res;
  try {
    res = await lmsFetch(BASE + path, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch (e) {
    throw new ApiError("无法连接到本地服务，请确认后端已启动", 0);
  }

  let data = null;
  const text = await res.text();
  if (text) {
    try { data = JSON.parse(text); } catch { data = { raw: text }; }
  }

  if (!res.ok) {
    const msg = (data && data.error) || `请求失败（${res.status}）`;
    throw new ApiError(msg, res.status);
  }
  return data;
}

const get = (p) => request("GET", p);
const post = (p, b) => request("POST", p, b);
const put = (p, b) => request("PUT", p, b);
const patch = (p, b) => request("PATCH", p, b);
const del = (p) => request("DELETE", p);

function qs(params) {
  const usp = new URLSearchParams();
  Object.entries(params || {}).forEach(([k, v]) => {
    if (v !== undefined && v !== null && v !== "") usp.append(k, v);
  });
  const s = usp.toString();
  return s ? "?" + s : "";
}

export const api = {
  // 认证
  login: (username, password) => post("/api/auth/login", { username, password }),
  register: (payload) => post("/api/auth/register", payload),
  me: () => get("/api/auth/me"),
  changePassword: (oldPassword, newPassword) =>
    post("/api/auth/password", { oldPassword, newPassword }),

  // 分类
  listCategories: () => get("/api/categories"),
  addCategory: (name) => post("/api/categories", { name }),
  listRegions: () => get("/api/regions"),

  // 图书
  listBooks: ({ keyword, category, region } = {}) =>
    get("/api/books" + qs({ keyword, category, region })),
  createBook: (book) => post("/api/books", book),
  updateBook: (id, book) => put(`/api/books/${id}`, book),
  deleteBook: (id) => del(`/api/books/${id}`),

  // 读者管理
  listReaders: (keyword) => get("/api/readers" + qs({ keyword })),
  createReader: (payload) => post("/api/readers", payload),
  setReaderStatus: (id, status) => patch(`/api/readers/${id}/status`, { status }),
  resetReaderPassword: (id) => post(`/api/readers/${id}/reset-password`),

  // 借阅（管理员）
  activeBorrows: (readerId) => get("/api/borrows/active" + qs({ readerId })),
  historyBorrows: (readerId) => get("/api/borrows/history" + qs({ readerId })),
  createBorrow: (readerId, bookId) => post("/api/borrows", { readerId, bookId }),
  returnBook: (id) => post(`/api/borrows/${id}/return`),

  // 读者自助
  myReader: () => get("/api/me/reader"),
  myActive: () => get("/api/me/borrows/active"),
  myHistory: () => get("/api/me/borrows/history"),
  selfBorrow: (bookId) => post("/api/me/borrows", { bookId }),
  selfRenew: (id) => post(`/api/me/borrows/${id}/renew`),

  // 统计与日志
  dashboard: () => get("/api/stats/dashboard"),
  categoryStats: () => get("/api/stats/categories"),
  hotBooks: () => get("/api/stats/hot-books"),
  logs: () => get("/api/logs"),
};
