// myborrows.js — 读者我的借阅：在借（可续借一次）/ 历史记录。
import { api } from "../api.js";
import { loading, renderTable, toast, badge } from "../components.js";

function stateBadge(r) {
  if (r.status === "returned") return badge("已归还", "gray");
  if (r.state && r.state.includes("逾期")) return badge(r.state, "red");
  return badge(r.state || "正常", "green");
}

export async function renderMyBorrows(container) {
  loading(container);
  const wrap = document.createElement("div");
  wrap.innerHTML = `
    <div class="login-tabs" style="max-width:340px">
      <button data-t="active" class="active">当前借阅</button>
      <button data-t="history">历史记录</button>
    </div>
    <div id="table" style="margin-top:16px"></div>`;
  container.innerHTML = "";
  container.appendChild(wrap);
  const tableEl = wrap.querySelector("#table");
  let mode = "active";

  const columnsActive = [
    { key: "title", title: "书名" },
    { key: "borrowDate", title: "借出日期" },
    { key: "dueDate", title: "应还日期" },
    { key: "renewCount", title: "已续借", align: "center", render: (r) => `${r.renewCount} 次` },
    { title: "状态", align: "center", render: stateBadge },
    {
      title: "操作", align: "right",
      render: (r) => `<div class="row-actions">
        <button class="btn ${r.renewCount === 0 ? "btn-outline" : "btn-ghost"} btn-sm act-renew"
          ${r.renewCount === 0 ? "" : "disabled"}>
          ${r.renewCount === 0 ? "续借" : "不可续借"}</button></div>`,
    },
  ];
  const columnsHistory = [
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
    const recs = mode === "active" ? await api.myActive() : await api.myHistory();
    tableEl.innerHTML = "";

    // 顶部到期提醒：3 天内应还且未逾期的图书。
    let banner = wrap.querySelector(".due-banner");
    if (banner) banner.remove();
    if (mode === "active") {
      const today = new Date(); today.setHours(0,0,0,0);
      const warn = recs.filter((r) => {
        if (r.status !== "borrowed" || !r.dueDate) return false;
        if (r.state && r.state.includes("逾期")) return false;
        const due = new Date(r.dueDate); due.setHours(0,0,0,0);
        const diff = (due - today) / 86400000;
        return diff >= 0 && diff <= 3;
      });
      if (warn.length) {
        const b = document.createElement("div");
        b.className = "due-banner";
        b.innerHTML = `<strong>温馨提醒：</strong>以下 ${warn.length} 本图书将在 3 天内到期，请及时归还或续借：
          ${warn.map((r) => `「${r.title}」（应还 ${r.dueDate}）`).join("、")}`;
        wrap.insertBefore(b, tableEl);
      }
    }

    renderTable(tableEl, mode === "active" ? columnsActive : columnsHistory, recs);
    tableEl.querySelectorAll(".act-renew").forEach((b) =>
      b.addEventListener("click", async () => {
        const r = recs[Number(b.closest("tr").rowIndex) - 1];
        try {
          await api.selfRenew(r.id);
          toast("续借成功，应还日期已延长 30 天", "success");
          load();
        } catch (err) { toast(err.message, "error"); }
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

  load();
}
