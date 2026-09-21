// setup.js — 服务器地址配置：首次启动或点击"服务器设置"时，填写并测试后端地址。
import { setServerUrl, testConnection, getServerUrl } from "./api.js";
import { esc, toast } from "./components.js";

// renderServerSetup：在 login-root 里渲染配置页，保存并测试通过后调用 onDone。
export function renderServerSetup(onDone) {
  const root = document.getElementById("login-root");
  document.getElementById("boot").classList.add("hidden");
  document.getElementById("app-root").classList.add("hidden");
  root.classList.remove("hidden");

  const prev = getServerUrl();
  root.innerHTML = `
    <div class="login-card">
      <div class="login-logo">图</div>
      <h2>连接服务器</h2>
      <p class="login-sub">请输入图书系统后端的访问地址</p>
      <div class="field">
        <label>服务器地址</label>
        <input class="input" id="srv-url" placeholder="例如 https://example.trycloudflare.com" value="${esc(prev)}" />
      </div>
      <p class="form-hint">公网部署填 https 地址；局域网部署填 http://服务器IP:8765。</p>
      <button class="btn btn-primary btn-block" id="srv-test" type="button">测试连接并保存</button>
      <p style="margin-top:10px"><a href="javascript:void(0)" id="srv-cancel" style="color:#888">返回</a></p>
    </div>`;

  const input = root.querySelector("#srv-url");
  const btn = root.querySelector("#srv-test");
  btn.onclick = async () => {
    btn.disabled = true;
    try {
      await testConnection(input.value);
      setServerUrl(input.value);
      toast("连接成功，已保存", "success");
      root.classList.add("hidden");
      onDone();
    } catch (err) {
      toast(err.message || "连接失败，请检查地址和网络", "error");
    }
    btn.disabled = false;
  };
  root.querySelector("#srv-cancel").onclick = () => {
    if (getServerUrl()) { root.classList.add("hidden"); onDone(); }
  };
  input.focus();
}
