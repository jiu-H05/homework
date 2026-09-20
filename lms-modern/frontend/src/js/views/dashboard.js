// dashboard.js — 管理员控制台：核心指标、分类馆藏、热门图书。
import { api } from "../api.js";
import { esc, loading, renderTable } from "../components.js";

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
}
