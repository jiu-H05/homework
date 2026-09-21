// dashboard.js — 管理员控制台：核心指标、分类馆藏、热门图书。
import { api, downloadBackup, uploadBackup } from "../api.js";
import { esc, loading, renderTable, toast } from "../components.js";

function statCard(label, value, ico, icoClass) {
  return `<div class="stat-card">
    <div class="stat-ico ${icoClass}">${ico}</div>
    <div class="stat-label">${esc(label)}</div>
    <div class="stat-value">${value}</div>
  </div>`;
}

export async function renderDashboard(container) {
  loading(container);
  const [dash, catStats, hot] = await Promise.all([
    api.dashboard(),
    api.categoryStats(),
    api.hotBooks(),
  ]);

  container.innerHTML = `
    <div class="stat-grid">
      ${statCard("图书种类", dash.bookKinds, "📚", "ico-blue")}
      ${statCard("图书总册数", dash.bookCopies, "📖", "ico-green")}
      ${statCard("注册读者", dash.readers, "👤", "ico-orange")}
      ${statCard("当前借出", dash.borrowed, "🔖", "ico-blue")}
      ${statCard("逾期未还", dash.overdue, "⚠", "ico-red")}
    </div>
    <div class="grid-2">
      <div class="card">
        <h3 class="card-title">分类馆藏分布</h3>
        <div id="cat-bars"></div>
      </div>
      <div class="card">
        <h3 class="card-title">热门图书（按借阅次数）</h3>
        <div id="hot-table"></div>
      </div>
    </div>`;

  // 分类条形图
  const maxCopies = Math.max(1, ...catStats.map((c) => c.copies));
  const bars = container.querySelector("#cat-bars");
  if (!catStats.length) {
    bars.innerHTML = `<p class="muted">暂无数据</p>`;
  } else {
    bars.innerHTML = catStats
      .map(
        (c) => `
      <div class="bar-row">
        <div class="bar-label" title="${esc(c.name)}">${esc(c.name)}</div>
        <div class="bar-track"><div class="bar-fill" style="width:${(c.copies / maxCopies) * 100}%"></div></div>
        <div class="bar-value">${c.copies} 册</div>
      </div>`
      )
      .join("");
  }

  renderTable(
    container.querySelector("#hot-table"),
    [
      { key: "title", title: "书名" },
      { key: "author", title: "作者" },
      { key: "times", title: "借阅次数", align: "right" },
    ],
    hot
  );

  // 数据备份与恢复（应用内导入导出，换机/备份不用手动找文件夹）。
  const box = document.createElement("div");
  box.className = "card";
  box.style.marginTop = "16px";
  box.innerHTML = `
    <h3 class="card-title">数据备份与恢复</h3>
    <p class="muted" style="margin:0 0 12px">导出备份可保存到 U 盘/网盘；在另一台电脑装好本系统后，用“导入备份”选择该文件即可还原全部图书与账号。</p>
    <div style="display:flex;gap:10px;align-items:center">
      <button class="btn btn-primary" id="btn-export" type="button">导出备份（.db）</button>
      <button class="btn btn-ghost" id="btn-import" type="button">导入备份…</button>
      <input type="file" id="import-file" accept=".db" style="display:none" />
    </div>`;
  container.appendChild(box);

  box.querySelector("#btn-export").onclick = async (e) => {
    e.target.disabled = true;
    try {
      await downloadBackup();
      toast("备份已导出", "success");
    } catch (err) { toast(err.message, "error"); }
    e.target.disabled = false;
  };
  box.querySelector("#btn-import").onclick = () =>
    box.querySelector("#import-file").click();
  box.querySelector("#import-file").onchange = async (ev) => {
    const f = ev.target.files && ev.target.files[0];
    if (!f) return;
    if (!confirm(`确定用备份文件「${f.name}」覆盖当前全部数据吗？此操作不可撤销。`)) {
      ev.target.value = "";
      return;
    }
    try {
      await uploadBackup(f);
      toast("导入成功，正在刷新…", "success");
      setTimeout(() => location.reload(), 800);
    } catch (err) { toast(err.message, "error"); }
    ev.target.value = "";
  };
}
