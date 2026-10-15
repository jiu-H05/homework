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
  const [dash, catStats, hot, trend] = await Promise.all([
    api.dashboard(),
    api.categoryStats(),
    api.hotBooks(),
    api.borrowTrend().catch(() => []),
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
        <h3 class="card-title">分类馆藏占比</h3>
        <div id="cat-pie" class="pie-wrap"></div>
      </div>
      <div class="card">
        <h3 class="card-title">近 30 天借阅趋势</h3>
        <div id="trend-chart"></div>
      </div>
    </div>
    <div class="grid-2" style="margin-top:16px">
      <div class="card">
        <h3 class="card-title">分类馆藏分布</h3>
        <div id="cat-bars"></div>
      </div>
      <div class="card">
        <h3 class="card-title">热门图书（按借阅次数）</h3>
        <div id="hot-table"></div>
      </div>
    </div>`;

  // 分类占比饼图（纯 CSS conic-gradient，无外部依赖）
  const pie = container.querySelector("#cat-pie");
  if (!catStats.length) {
    pie.innerHTML = `<p class="muted">暂无数据</p>`;
  } else {
    const total = catStats.reduce((s, c) => s + c.copies, 0) || 1;
    const palette = ["#1f6fb2", "#1f9d6b", "#d98a1f", "#b24a9f", "#6b5bd8", "#d64545", "#3aa6b9", "#8a8f98"];
    let acc = 0;
    const stops = catStats.map((c, i) => {
      const start = (acc / total) * 100;
      acc += c.copies;
      const end = (acc / total) * 100;
      const color = palette[i % palette.length];
      return `${color} ${start}% ${end}%`;
    }).join(", ");
    const legend = catStats.map((c, i) =>
      `<div class="pie-legend-item"><span class="dot" style="background:${palette[i % palette.length]}"></span>${esc(c.name)} <span class="muted">${Math.round(c.copies/total*100)}%</span></div>`
    ).join("");
    pie.innerHTML = `<div class="pie" style="background:conic-gradient(${stops})"></div>
      <div class="pie-legend">${legend}</div>`;
  }

  // 借阅趋势折线图（纯 SVG）
  const trendEl = container.querySelector("#trend-chart");
  if (!trend || !trend.length) {
    trendEl.innerHTML = `<p class="muted">近 30 天暂无借阅记录</p>`;
  } else {
    const W = 460, H = 180, pad = 24;
    const max = Math.max(1, ...trend.map((t) => t.count));
    const step = (W - pad * 2) / Math.max(1, trend.length - 1);
    const pts = trend.map((t, i) => [pad + i * step, H - pad - (t.count / max) * (H - pad * 2)]);
    const path = pts.map((p, i) => (i ? "L" : "M") + p[0].toFixed(1) + "," + p[1].toFixed(1)).join(" ");
    trendEl.innerHTML = `<svg viewBox="0 0 ${W} ${H}" class="trend-svg">
      ${[0, 0.5, 1].map((r) => `<line x1="${pad}" x2="${W-pad}" y1="${pad+r*(H-2*pad)}" y2="${pad+r*(H-2*pad)}" stroke="var(--border)"/>`).join("")}
      <path d="${path}" fill="none" stroke="var(--primary)" stroke-width="2"/>
      ${pts.map((p) => `<circle cx="${p[0]}" cy="${p[1]}" r="3" fill="var(--primary)"/>`).join("")}
    </svg>
    <div class="muted" style="margin-top:4px;font-size:12px">共 ${trend.reduce((s,t)=>s+t.count,0)} 次借阅</div>`;
  }

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
