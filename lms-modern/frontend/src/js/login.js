// login.js — 登录 / 注册视图。
import { api } from "./api.js";
import { esc, toast } from "./components.js";

export function renderLogin(onSuccess) {
  const root = document.getElementById("login-root");
  root.innerHTML = `
    <div class="login-card">
      <div class="login-logo">图</div>
      <h2>图书管理系统</h2>
      <p class="login-sub">Library Management System</p>
      <div class="login-tabs">
        <button data-tab="login" class="active">登录</button>
        <button data-tab="register">读者注册</button>
      </div>
      <div id="login-pane">
        <form id="login-form">
          <div class="field">
            <label>用户名</label>
            <input class="input" name="username" required autocomplete="username" placeholder="请输入用户名" />
          </div>
          <div class="field">
            <label>口令</label>
            <input class="input" type="password" name="password" required autocomplete="current-password" placeholder="请输入口令" />
          </div>
          <button class="btn btn-primary btn-block" type="submit">登 录</button>
        </form>
        <p class="form-hint">默认管理员 admin / admin123；读者 reader / reader123</p>
      </div>
      <div id="register-pane" class="hidden">
        <form id="register-form">
          <div class="form-row">
            <div class="field">
              <label>真实姓名</label>
              <input class="input" name="realName" required placeholder="姓名" />
            </div>
            <div class="field">
              <label>手机号</label>
              <input class="input" name="phone" required placeholder="手机号" />
            </div>
          </div>
          <div class="field">
            <label>用户名</label>
            <input class="input" name="username" required placeholder="设置登录用户名" />
          </div>
          <div class="form-row">
            <div class="field">
              <label>口令</label>
              <input class="input" type="password" name="password" required placeholder="设置口令" />
            </div>
            <div class="field">
              <label>确认口令</label>
              <input class="input" type="password" name="password2" required placeholder="再次输入" />
            </div>
          </div>
          <button class="btn btn-primary btn-block" type="submit">注 册</button>
        </form>
      </div>
    </div>`;

  const tabs = root.querySelectorAll(".login-tabs button");
  tabs.forEach((b) =>
    b.addEventListener("click", () => {
      tabs.forEach((x) => x.classList.remove("active"));
      b.classList.add("active");
      const tab = b.dataset.tab;
      root.querySelector("#login-pane").classList.toggle("hidden", tab !== "login");
      root.querySelector("#register-pane").classList.toggle("hidden", tab !== "register");
    })
  );

  root.querySelector("#login-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const f = e.target;
    const btn = f.querySelector('button[type="submit"]');
    btn.disabled = true;
    try {
      const data = await api.login(f.username.value.trim(), f.password.value);
      toast("登录成功", "success");
      onSuccess(data);
    } catch (err) {
      toast(err.message, "error");
      btn.disabled = false;
    }
  });

  root.querySelector("#register-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const f = e.target;
    const btn = f.querySelector('button[type="submit"]');
    if (f.password.value !== f.password2.value) {
      toast("两次输入的口令不一致", "error");
      return;
    }
    if (f.password.value.length < 6) {
      toast("口令长度至少 6 位", "error");
      return;
    }
    btn.disabled = true;
    try {
      const data = await api.register({
        username: f.username.value.trim(),
        password: f.password.value,
        realName: f.realName.value.trim(),
        phone: f.phone.value.trim(),
      });
      toast(`注册成功，读者编号 ${data.readerNo}，请登录`, "success", 3600);
      tabs[0].click();
      root.querySelector("#login-form").username.value = f.username.value.trim();
    } catch (err) {
      toast(err.message, "error");
    } finally {
      btn.disabled = false;
    }
  });
}
