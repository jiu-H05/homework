// readers.js — 管理员读者管理：检索、新增、启用/停用、重置口令。
import { api } from "../api.js";
import {
  esc, loading, renderTable, openModal, confirmDialog, toast, formValues, badge,
} from "../components.js";

function readerForm() {
  const form = document.createElement("form");
  form.innerHTML = `
    <div class="form-row">
      <div class="field"><label>真实姓名</label><input class="input" name="realName" required /></div>
      <div class="field"><label>手机号</label><input class="input" name="phone" /></div>
    </div>
    <div class="form-row">
      <div class="field"><label>登录用户名</label><input class="input" name="username" required /></div>
      <div class="field"><label>初始口令</label><input class="input" name="password" value="reader123" required /></div>
    </div>
    <p class="form-hint">默认初始口令 reader123，读者登录后可自行修改。</p>`;
  return form;
}

export async function renderReaders(container) {
  loading(container);
  const wrap = document.createElement("div");
  wrap.innerHTML = `
    <div class="toolbar">
      <input class="input" id="kw" placeholder="搜索姓名 / 读者编号 / 用户名" />
      <button class="btn btn-primary btn-sm" id="search">查询</button>
      <div class="spacer"></div>
      <button class="btn btn-primary btn-sm" id="add-reader">+ 新增读者</button>
    </div>
    <div id="table"></div>`;
  container.innerHTML = "";
  container.appendChild(wrap);
  const tableEl = wrap.querySelector("#table");

  const load = async () => {
    loading(tableEl);
    const keyword = wrap.querySelector("#kw").value.trim();
    const readers = await api.listReaders(keyword);
    tableEl.innerHTML = "";
    renderTable(
      tableEl,
      [
        { key: "readerNo", title: "读者编号" },
        { key: "name", title: "姓名" },
        { key: "phone", title: "手机号" },
        { key: "username", title: "登录用户名" },
        {
          key: "status", title: "状态", align: "center",
          render: (r) => badge(r.status === "active" ? "正常" : "已停用", r.status === "active" ? "green" : "gray"),
        },
        { key: "createdAt", title: "注册时间" },
        {
          title: "操作", align: "right",
          render: (r) => `
            <div class="row-actions">
              <button class="btn btn-ghost btn-sm act-pwd">重置口令</button>
              <button class="btn ${r.status === "active" ? "btn-danger" : "btn-outline"} btn-sm act-toggle">
                ${r.status === "active" ? "停用" : "启用"}</button>
            </div>`,
        },
      ],
      readers
    );

    tableEl.querySelectorAll(".act-pwd").forEach((b) =>
      b.addEventListener("click", async () => {
        const r = readers[Number(b.closest("tr").rowIndex) - 1];
        if (await confirmDialog(`确定将 ${r.name} 的口令重置为临时口令吗？`, { okText: "重置" })) {
          try {
            const res = await api.resetReaderPassword(r.userId);
            openModal({
              title: "口令已重置",
              body: `<p>读者 <b>${esc(r.name)}</b> 的新临时口令为：</p>
                <p style="font-size:20px;font-weight:700;letter-spacing:1px">${esc(res.tempPassword)}</p>
                <p class="form-hint">请告知读者尽快登录修改。</p>`,
              footer: "",
            });
          } catch (err) { toast(err.message, "error"); }
        }
      })
    );

    tableEl.querySelectorAll(".act-toggle").forEach((b) =>
      b.addEventListener("click", async () => {
        const r = readers[Number(b.closest("tr").rowIndex) - 1];
        const toActive = r.status !== "active";
        if (await confirmDialog(`确定${toActive ? "启用" : "停用"}读者 ${r.name} 吗？`, { okText: "确定" })) {
          try {
            await api.setReaderStatus(r.userId, toActive ? "active" : "disabled");
            toast("状态已更新", "success");
            load();
          } catch (err) { toast(err.message, "error"); }
        }
      })
    );
  };

  wrap.querySelector("#search").onclick = load;
  wrap.querySelector("#kw").addEventListener("keydown", (e) => { if (e.key === "Enter") load(); });
  wrap.querySelector("#add-reader").onclick = () => {
    const form = readerForm();
    const footer = document.createElement("div");
    const btnCancel = document.createElement("button");
    btnCancel.className = "btn btn-ghost"; btnCancel.textContent = "取消";
    const btnSave = document.createElement("button");
    btnSave.className = "btn btn-primary"; btnSave.textContent = "保存";
    footer.append(btnCancel, btnSave);
    const m = openModal({ title: "新增读者", body: form, footer });
    btnCancel.onclick = m.close;
    btnSave.onclick = async () => {
      const v = formValues(form);
      btnSave.disabled = true;
      try {
        const res = await api.createReader(v);
        toast(`读者已新增，编号 ${res.readerNo}`, "success");
        m.close();
        load();
      } catch (err) { toast(err.message, "error"); btnSave.disabled = false; }
    };
  };

  load();
}
