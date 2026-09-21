// profile.js — 读者个人信息。
import { api } from "../api.js";
import { esc, loading } from "../components.js";

export async function renderProfile(container) {
  loading(container);
  const [me, reader] = await Promise.all([api.me(), api.myReader()]);
  container.innerHTML = `
    <div class="card" style="max-width:560px">
      <h3 class="card-title">个人信息</h3>
      <dl class="desc-list">
        <dt>读者编号</dt><dd>${esc(reader.readerNo)}</dd>
        <dt>姓名</dt><dd>${esc(reader.name)}</dd>
        <dt>手机号</dt><dd>${esc(reader.phone)}</dd>
        <dt>登录用户名</dt><dd>${esc(me.username)}</dd>
        <dt>注册时间</dt><dd>${esc(reader.createdAt)}</dd>
      </dl>
      <p class="form-hint" style="margin-top:18px">如需修改登录口令，请点击右上角“修改口令”。</p>
    </div>`;
}
