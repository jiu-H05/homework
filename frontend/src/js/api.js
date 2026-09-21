// api.js — 后端 REST API 客户端。
// 职责单一：封装请求、令牌注入、错误归一化；不包含任何视图逻辑。

import { lmsFetch, tauriInvoke } from "./net.js";

// 前后端分离：前端是 Tauri 客户端，后端是独立服务器。
// 服务器地址存 localStorage，由用户在"服务器设置"里填写一次。
// 浏览器直接访问后端页面时，base 留空走同源。
const SERVER_KEY = "lms_server_url";
const inTauri = typeof window !== "undefined" && !!window.__TAURI_INTERNALS__;
export const isTauri = inTauri;
let base = "";

function readStored() {
  try { return localStorage.getItem(SERVER_KEY) || ""; } catch { return ""; }
}

// getBase 返回当前 API 前缀。
export function getBase() { return base; }

export function getServerUrl() { return base; }
export function setServerUrl(u) {
  base = (u || "").trim().replace(/\/+$/, "");
  try { base ? localStorage.setItem(SERVER_KEY, base) : localStorage.removeItem(SERVER_KEY); } catch {}
}

// hasServerConfigured：是否已经配置过服务器地址。
export function hasServerConfigured() { return !!base; }

// resolveApiBase 启动时确定 base：浏览器同源留空；Tauri/移动端读本地配置。
export async function resolveApiBase() {
  if (!inTauri) {
    base = "";
    return base;
  }
  base = readStored();
  return base;
}

// testConnection：用给定地址探测 /api/health，用于设置页"测试连接"。
export async function testConnection(url) {
  const u = (url || "").trim().replace(/\/+$/, "");
  if (!u) throw new ApiError("请输入服务器地址", 0);
  const res = await lmsFetch(u + "/api/health");
  if (!res.ok) throw new ApiError("服务器返回 " + res.status, res.status);
  return true;
}
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

async function request(method, path, body, _retried) {
  const headers = {};
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  if (body !== undefined) headers["Content-Type"] = "application/json";

  let res;
  try {
    res = await lmsFetch(base + path, {
      method,
      headers,
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
  } catch (e) {
    // 连接失败多为网络切换或服务地址变更，重新读取一次配置再重试一次。
    if (!_retried) {
      await resolveApiBase();
      return request(method, path, body, true);
    }
    throw new ApiError("无法连接到服务器，请检查网络或服务器地址", 0);
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

// ---- 内存缓存：登录后拉一次共享基础数据，切页面/筛选不再重复请求 ----
const cache = {
  categories: null,
  regions: null,
  books: null, // 全量图书（本地筛选/排序）
};

// invalidateCache：写操作后调用，让受影响的缓存失效。
// scope: "all" | "meta"（分类/地区）| "books"（图书）
export function invalidateCache(scope = "all") {
  if (scope === "all" || scope === "meta") {
    cache.categories = null;
    cache.regions = null;
  }
  if (scope === "all" || scope === "books") {
    cache.books = null;
  }
}

// localFilterBooks：在本地对全量图书做关键词/分类/地区筛选，零网络请求。
function localFilterBooks(books, { keyword, category, region } = {}) {
  let out = books;
  if (category) out = out.filter((b) => b.categoryId === Number(category));
  if (region) out = out.filter((b) => b.region === region);
  if (keyword) {
    const kw = keyword.toLowerCase();
    out = out.filter(
      (b) =>
        (b.title && b.title.toLowerCase().includes(kw)) ||
        (b.author && b.author.toLowerCase().includes(kw)) ||
        (b.isbn && b.isbn.toLowerCase().includes(kw))
    );
  }
  return out;
}

export const api = {
  // 认证
  login: (username, password) => post("/api/auth/login", { username, password }),
  register: (payload) => post("/api/auth/register", payload),
  me: () => get("/api/auth/me"),
  changePassword: (oldPassword, newPassword) =>
    post("/api/auth/password", { oldPassword, newPassword }),

  // 分类（内存缓存）
  listCategories: async () => {
    if (cache.categories) return cache.categories;
    const d = await get("/api/categories");
    cache.categories = d;
    return d;
  },
  addCategory: async (name) => {
    const r = await post("/api/categories", { name });
    invalidateCache("meta");
    return r;
  },
  listRegions: async () => {
    if (cache.regions) return cache.regions;
    const d = await get("/api/regions");
    cache.regions = d;
    return d;
  },

  // 图书：首次拉全量缓存，之后筛选全在本地做，不再请求后端
  listBooks: async ({ keyword, category, region } = {}) => {
    if (cache.books == null) {
      cache.books = await get("/api/books");
    }
    return localFilterBooks(cache.books, { keyword, category, region });
  },
  createBook: async (book) => {
    const r = await post("/api/books", book);
    invalidateCache("books");
    return r;
  },
  updateBook: async (id, book) => {
    const r = await put(`/api/books/${id}`, book);
    invalidateCache("books");
    return r;
  },
  deleteBook: async (id) => {
    const r = await del(`/api/books/${id}`);
    invalidateCache("books");
    return r;
  },

  // 读者管理
  listReaders: (keyword) => get("/api/readers" + qs({ keyword })),
  createReader: (payload) => post("/api/readers", payload),
  setReaderStatus: (id, status) => patch(`/api/readers/${id}/status`, { status }),
  resetReaderPassword: (id) => post(`/api/readers/${id}/reset-password`),

  // 借阅（管理员）
  activeBorrows: (readerId) => get("/api/borrows/active" + qs({ readerId })),
  historyBorrows: (readerId) => get("/api/borrows/history" + qs({ readerId })),
  createBorrow: async (readerId, bookId) => {
    const r = await post("/api/borrows", { readerId, bookId });
    invalidateCache("books");
    return r;
  },
  returnBook: async (id) => {
    const r = await post(`/api/borrows/${id}/return`);
    invalidateCache("books");
    return r;
  },

  // 读者自助
  myReader: () => get("/api/me/reader"),
  myActive: () => get("/api/me/borrows/active"),
  myHistory: () => get("/api/me/borrows/history"),
  selfBorrow: async (bookId) => {
    const r = await post("/api/me/borrows", { bookId });
    invalidateCache("books");
    return r;
  },
  selfRenew: (id) => post(`/api/me/borrows/${id}/renew`),

  // 统计与日志
  dashboard: () => get("/api/stats/dashboard"),
  categoryStats: () => get("/api/stats/categories"),
  hotBooks: () => get("/api/stats/hot-books"),
  logs: () => get("/api/logs"),
};

// downloadBackup 导出当前数据库为一个 .db 备份文件（管理员）。
export async function downloadBackup() {
  const headers = {};
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await lmsFetch(getBase() + "/api/admin/backup", { headers });
  if (!res.ok) throw new ApiError("导出失败（" + res.status + "）", res.status);
  const blob = await res.blob();
  let name = "lms-backup.db";
  const cd = res.headers.get("Content-Disposition") || "";
  const m = cd.match(/filename="?([^"]+)"?/);
  if (m) name = m[1];
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url; a.download = name;
  document.body.appendChild(a); a.click(); a.remove();
  URL.revokeObjectURL(url);
}

// uploadBackup 导入一个 .db 备份文件，后端校验后热替换当前数据（管理员）。
export async function uploadBackup(file) {
  const headers = { "Content-Type": "application/octet-stream" };
  const token = getToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  const buf = await file.arrayBuffer();
  const res = await lmsFetch(getBase() + "/api/admin/restore", {
    method: "POST",
    headers,
    body: buf,
  });
  let data = null;
  const text = await res.text();
  if (text) { try { data = JSON.parse(text); } catch { data = { raw: text }; } }
  if (!res.ok) throw new ApiError((data && data.error) || "导入失败（" + res.status + "）", res.status);
  return data;
}
