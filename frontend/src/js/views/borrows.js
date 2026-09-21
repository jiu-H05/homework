// borrows.js — 管理员借阅管理：在借/历史查询、办理借阅、办理归还（含逾期罚款）。
import { api } from "../api.js";
import {
  esc, loading, renderTable, openModal, confirmDialog, toast, formValues, badge,
} from "../components.js";

function stateBadge(r) {
  if (r.status === "returned") return badge("已归还", "gray");
  if (r.state && r.state.includes("逾期")) return badge(r.state, "red");
  return badge(r.state || "正常", "green");
}

function borrowForm(readers, books) {
  const form = document.createElement("form");
  form.innerHTML = `
    <div class="field"><label>读者</label>
      <select class="input" name="readerId" required>
        <option value="">请选择读者</option>
        ${readers.filter((r) => r.status === "active").map(
    (r) => `<option value="${r.userId}">${esc(r.readerNo)} ${esc(r.name)}</option>`
  ).join("")}
      </select></div>
    <div class="field"><label>图书</label>
      <select class="input" name="bookId" required>
        <option value="">请选择图书</option>
        ${books.filter((b) => b.availableQty > 0).map(
    (b) => `<option value="${b.id}">${esc(b.title)}（可借 ${b.availableQty}）</option>`
  ).join("")}
      </select></div>`;
  return form;
}

export async function renderBorrows(container) {
  loading(container);
  const wrap = document.createElement("div");
  wrap.innerHTML = `
    <div class="login-tabs" style="max-width:340px">
      <button data-t="active" class="active">在借图书</button>
      <button data-t="history">历史记录</button>
    </div>
    <div class="toolbar">
      <div class="spacer"></div>
      <button class="btn btn-primary btn-sm" id="add-borrow">+ 办理借阅</button>
    </div>
    <div id="table"></div>`;
  container.innerHTML = "";
  container.appendChild(wrap);
  const tableEl = wrap.querySelector("#table");
  let mode = "active";

  const columnsActive = [
    { key: "readerNo", title: "读者编号" },
    { key: "readerName", title: "读者" },
    { key: "title", title: "书名" },
    { key: "borrowDate", title: "借出日期" },
    { key: "dueDate", title: "应还日期" },
    { key: "renewCount", title: "续借", align: "center", render: (r) => `${r.renewCount} 次` },
    { title: "状态", align: "center", render: stateBadge },
    {
      title: "操作", align: "right",
      render: (r) => `<div class="row-actions"><button class="btn btn-primary btn-sm act-return">办理归还</button></div>`,
    },
  ];
  const columnsHistory = [
    { key: "readerNo", title: "读者编号" },
    { key: "readerName", title: "读者" },
    { key: "title", title: "书名" },
    { key: "borrowDate", title: "借出日期" },
    { key: "returnDate", title: "归还日期" },
    { key: "renewCount", title: "续借次数", align: "center" },
    {
      key: "fine", title: "罚款(元)", align: "right",
      render: (r) => (r.fine > 0 ? `<span style="color:var(--danger)">${r.fine.toFixed(2)}</span>` : "0.00"),
    },
    { title: "状态", align: "center", render: stateBadge },
  ];

  const load = async () => {
    loading(tableEl);
    const recs = mode === "active" ? await api.activeBorrows() : await api.historyBorrows();
    tableEl.innerHTML = "";
    renderTable(tableEl, mode === "active" ? columnsActive : columnsHistory, recs);

    tableEl.querySelectorAll(".act-return").forEach((b) =>
      b.addEventListener("click", async () => {
        const r = recs[Number(b.closest("tr").rowIndex) - 1];
        const overdue = r.state && r.state.includes("逾期");
        if (await confirmDialog(`确定为 ${r.readerName} 归还《${r.title}》吗？${overdue ? "该图书已逾期，将计算罚款。" : ""}`,
          { okText: "归还" })) {
          try {
            const res = await api.returnBook(r.id);
            if (res.fine > 0) {
              openModal({
                title: "归还成功",
                body: `<p>《${esc(r.title)}》已归还。</p>
                  <p>逾期罚款：<span style="color:var(--danger);font-size:18px;font-weight:700">${Number(res.fine).toFixed(2)} 元</span></p>`,
                footer: "",
              });
            } else {
              toast("归还成功", "success");
            }
            load();
          } catch (err) { toast(err.message, "error"); }
        }
      })
    );
  };

  wrap.querySelectorAll(".login-tabs button").forEach((b) =>
    b.addEventListener("click", () => {
      wrap.querySelectorAll(".login-tabs button").forEach((x) => x.classList.remove("active"));
      b.classList.add("active");
      mode = b.dataset.t;
      load();
    })
  );

  wrap.querySelector("#add-borrow").onclick = async () => {
    const [readers, books] = await Promise.all([api.listReaders(), api.listBooks()]);
    const form = borrowForm(readers, books);
    const footer = document.createElement("div");
    const btnCancel = document.createElement("button");
    btnCancel.className = "btn btn-ghost"; btnCancel.textContent = "取消";
    const btnSave = document.createElement("button");
    btnSave.className = "btn btn-primary"; btnSave.textContent = "办理";
    footer.append(btnCancel, btnSave);
    const m = openModal({ title: "办理借阅", body: form, footer });
    btnCancel.onclick = m.close;
    btnSave.onclick = async () => {
      const v = formValues(form);
      if (!v.readerId || !v.bookId) { toast("请选择读者与图书", "error"); return; }
      try {
        await api.createBorrow(Number(v.readerId), Number(v.bookId));
        toast("借阅成功", "success");
        m.close();
        mode = "active";
        wrap.querySelectorAll(".login-tabs button").forEach((x) => x.classList.toggle("active", x.dataset.t === "active"));
        load();
      } catch (err) { toast(err.message, "error"); }
    };
  };

  load();
}
