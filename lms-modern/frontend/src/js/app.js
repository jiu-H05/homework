// app.js — 应用引导与外壳：等待后端、恢复登录态、构建导航、路由视图。
// 采用动态导入逐个加载模块，便于定位加载/评估失败的具体模块。
const bootP = () => document.getElementById("boot").querySelector("p");
const MOD = {};
async function load(label, path) {
  try {
    MOD[label] = await import(path);
  } catch (e) {
    bootP().textContent = `模块加载失败 [${label}] ${path}: ${e.message}`;
    throw e;
  }
}
async function loadAll() {
  await load("api", "./api.js");
  await load("net", "./net.js");
  await load("components", "./components.js");
  await load("login", "./login.js");
  await load("dashboard", "./views/dashboard.js");
  await load("books", "./views/books.js");
  await load("readers", "./views/readers.js");
  await load("borrows", "./views/borrows.js");
  await load("logs", "./views/logs.js");
  await load("catalog", "./views/catalog.js");
  await load("myborrows", "./views/myborrows.js");
  await load("profile", "./views/profile.js");
}
let api, setToken, getToken;
let lmsFetch, renderLogin, openModal, toast, formValues;
let renderDashboard, renderBooks, renderReaders, renderBorrows, renderLogs;
let renderCatalog, renderMyBorrows, renderProfile;
function wireRefs() {
  ({ api, setToken, getToken } = MOD.api);
  ({ lmsFetch } = MOD.net);
  ({ renderLogin } = MOD.login);
  ({ openModal, toast, formValues } = MOD.components);
  renderDashboard = MOD.dashboard.renderDashboard;
  renderBooks = MOD.books.renderBooks;
  renderReaders = MOD.readers.renderReaders;
  renderBorrows = MOD.borrows.renderBorrows;
  renderLogs = MOD.logs.renderLogs;
  renderCatalog = MOD.catalog.renderCatalog;
  renderMyBorrows = MOD.myborrows.renderMyBorrows;
  renderProfile = MOD.profile.renderProfile;
}

const NAV = {
  admin: [
    { id: "dashboard", title: "控制台", ico: "▦", render: (c) => renderDashboard(c) },
    { id: "books", title: "图书管理", ico: "📚", render: (c) => renderBooks(c) },
    { id: "readers", title: "读者管理", ico: "👥", render: (c) => renderReaders(c) },
    { id: "borrows", title: "借阅管理", ico: "🔁", render: (c) => renderBorrows(c) },
    { id: "logs", title: "操作日志", ico: "📋", render: (c) => renderLogs(c) },
  ],
  reader: [
    { id: "catalog", title: "馆藏浏览", ico: "📚", render: (c) => renderCatalog(c) },
    { id: "myborrows", title: "我的借阅", ico: "🔖", render: (c) => renderMyBorrows(c) },
    { id: "profile", title: "个人信息", ico: "👤", render: (c) => renderProfile(c) },
  ],
};

let currentUser = null;
let currentNav = null;

async function waitForBackend() {
  for (let i = 0; i < 60; i++) {
    try {
      const r = await lmsFetch("http://172.21.79.249:8765/api/health");
      if (r.ok) return true;
      document.getElementById("boot").querySelector("p").textContent =
        `health status ${r.status}（重试中）`;
    } catch (e) {
      document.getElementById("boot").querySelector("p").textContent =
        `fetch 错误：${e.message}（重试 ${i}）`;
      await new Promise((r) => setTimeout(r, 500));
    }
  }
  return false;
}

function showLogin() {
  document.getElementById("boot").classList.add("hidden");
  document.getElementById("app-root").classList.add("hidden");
  const lr = document.getElementById("login-root");
  lr.classList.remove("hidden");
  renderLogin(async () => {
    await enterApp();
  });
}

async function enterApp() {
  try {
    currentUser = await api.me();
  } catch {
    setToken(null);
    showLogin();
    return;
  }
  document.getElementById("boot").classList.add("hidden");
  document.getElementById("login-root").classList.add("hidden");
  document.getElementById("app-root").classList.remove("hidden");
  buildShell();
  const items = NAV[currentUser.role] || [];
  route(items[0].id);
}

function buildShell() {
  const navEl = document.getElementById("nav");
  const items = NAV[currentUser.role] || [];
  navEl.innerHTML = items
    .map(
      (it) => `
    <button class="nav-item" data-id="${it.id}" type="button">
      <span class="nav-ico">${it.ico}</span><span>${it.title}</span>
    </button>`
    )
    .join("");
  navEl.querySelectorAll(".nav-item").forEach((b) =>
    b.addEventListener("click", () => route(b.dataset.id))
  );

  const roleText = currentUser.role === "admin" ? "管理员" : "读者";
  document.getElementById("user-chip").innerHTML = `
    <span class="uc-name">${currentUser.realName || currentUser.username}</span>
    <span class="uc-role">${roleText} · ${currentUser.username}</span>`;
}

async function route(id) {
  const items = NAV[currentUser.role] || [];
  const item = items.find((x) => x.id === id);
  if (!item) return;
  currentNav = id;
  document.querySelectorAll(".nav-item").forEach((b) =>
    b.classList.toggle("active", b.dataset.id === id)
  );
  document.getElementById("page-title").textContent = item.title;
  const view = document.getElementById("view");
  view.innerHTML = "";
  try {
    await item.render(view);
  } catch (err) {
    if (err.status === 401) {
      setToken(null);
      showLogin();
    } else {
      view.innerHTML = `<div class="card">加载失败：${err.message}</div>`;
    }
  }
}

function openPasswordModal() {
  const form = document.createElement("form");
  form.innerHTML = `
    <div class="field"><label>原口令</label><input class="input" type="password" name="oldPassword" required /></div>
    <div class="field"><label>新口令</label><input class="input" type="password" name="password" required /></div>
    <div class="field"><label>确认新口令</label><input class="input" type="password" name="password2" required /></div>
    <p class="form-hint">新口令长度至少 6 位。</p>`;
  const footer = document.createElement("div");
  const btnCancel = document.createElement("button");
  btnCancel.className = "btn btn-ghost"; btnCancel.textContent = "取消";
  const btnSave = document.createElement("button");
  btnSave.className = "btn btn-primary"; btnSave.textContent = "保存";
  footer.append(btnCancel, btnSave);
  const m = openModal({ title: "修改口令", body: form, footer });
  btnCancel.onclick = m.close;
  btnSave.onclick = async () => {
    const v = formValues(form);
    if (v.password !== v.password2) { toast("两次输入的新口令不一致", "error"); return; }
    if (v.password.length < 6) { toast("新口令长度至少 6 位", "error"); return; }
    btnSave.disabled = true;
    try {
      await api.changePassword(v.oldPassword, v.password);
      toast("口令修改成功", "success");
      m.close();
    } catch (err) { toast(err.message, "error"); btnSave.disabled = false; }
  };
}

function wireGlobal() {
  document.getElementById("btn-logout").addEventListener("click", () => {
    setToken(null);
    currentUser = null;
    showLogin();
  });
  document.getElementById("btn-password").addEventListener("click", openPasswordModal);
}

async function bootstrap() {
  try {
    await loadAll();
    wireRefs();
  } catch (e) {
    return; // 失败信息已显示在启动页
  }
  wireGlobal();
  const ready = await waitForBackend();
  if (!ready) {
    document.getElementById("boot").innerHTML =
      `<p style="color:var(--danger)">后端服务未能启动，请重新打开应用。</p>`;
    return;
  }
  if (getToken()) await enterApp();
  else showLogin();
}

// E2E 自动化钩子：仅在外部通过 webview.eval 显式调用时生效，
// 复用与界面完全相同的登录/进入/路由逻辑（真实网络、真实认证、真实渲染）。
window.__e2e = {
  async login(username, password) {
    const data = await api.login(username, password);
    setToken(data.token);
    await enterApp();
    return currentUser.role;
  },
  async logout() {
    setToken(null);
    currentUser = null;
    showLogin();
    return true;
  },
  async go(id) {
    await route(id);
    return id;
  },
  role: () => currentUser && currentUser.role,
};

bootstrap();
