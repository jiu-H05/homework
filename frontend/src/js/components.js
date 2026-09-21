// components.js — 可复用的 UI 原语：提示、模态框、确认框、表格、表单取值。
// 所有视图共用，避免重复实现，保持视图模块高内聚。

export function esc(s) {
  if (s === null || s === undefined) return "";
  return String(s)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

/* ---------- Toast ---------- */
export function toast(message, type = "info", duration = 2800) {
  const root = document.getElementById("toast-root");
  const el = document.createElement("div");
  el.className = `toast ${type}`;
  el.innerHTML = `<span class="toast-msg">${esc(message)}</span>`;
  root.appendChild(el);
  setTimeout(() => {
    el.style.transition = "opacity .25s";
    el.style.opacity = "0";
    setTimeout(() => el.remove(), 260);
  }, duration);
}

/* ---------- Modal ---------- */
export function openModal({ title, body, footer, size, onClose }) {
  const root = document.getElementById("modal-root");
  const backdrop = document.createElement("div");
  backdrop.className = "modal-backdrop";
  const modal = document.createElement("div");
  modal.className = "modal" + (size === "lg" ? " modal-lg" : "");
  modal.innerHTML = `
    <div class="modal-header">
      <h3>${esc(title)}</h3>
      <button class="modal-close" type="button">&times;</button>
    </div>
    <div class="modal-body"></div>
    <div class="modal-footer"></div>`;
  const bodyEl = modal.querySelector(".modal-body");
  const footerEl = modal.querySelector(".modal-footer");
  if (typeof body === "string") bodyEl.innerHTML = body;
  else if (body instanceof Node) bodyEl.appendChild(body);
  if (footer instanceof Node) footerEl.appendChild(footer);
  else if (typeof footer === "string") footerEl.innerHTML = footer;
  else footerEl.remove();

  backdrop.appendChild(modal);
  root.appendChild(backdrop);

  const close = () => {
    backdrop.remove();
    if (onClose) onClose();
  };
  modal.querySelector(".modal-close").addEventListener("click", close);
  backdrop.addEventListener("click", (e) => { if (e.target === backdrop) close(); });
  return { close, bodyEl, footerEl, modal };
}

/* ---------- Confirm ---------- */
export function confirmDialog(message, { title = "请确认", okText = "确定", danger = true } = {}) {
  return new Promise((resolve) => {
    const footer = document.createElement("div");
    const btnCancel = document.createElement("button");
    btnCancel.className = "btn btn-ghost";
    btnCancel.textContent = "取消";
    const btnOk = document.createElement("button");
    btnOk.className = "btn " + (danger ? "btn-danger" : "btn-primary");
    btnOk.textContent = okText;
    footer.append(btnCancel, btnOk);
    const m = openModal({ title, body: `<p style="margin:0">${esc(message)}</p>`, footer });
    btnCancel.onclick = () => { m.close(); resolve(false); };
    btnOk.onclick = () => { m.close(); resolve(true); };
  });
}

/* ---------- Form helpers ---------- */
export function formValues(formEl) {
  const out = {};
  formEl.querySelectorAll("input, select, textarea").forEach((el) => {
    if (!el.name) return;
    out[el.name] = el.type === "checkbox" ? el.checked : el.value.trim();
  });
  return out;
}

export function fillForm(formEl, values) {
  formEl.querySelectorAll("input, select, textarea").forEach((el) => {
    if (!el.name || values[el.name] === undefined || values[el.name] === null) return;
    if (el.type === "checkbox") el.checked = !!values[el.name];
    else el.value = values[el.name];
  });
}

/* ---------- Table ---------- */
// columns: [{ key, title, render?(row), align, className }]
export function renderTable(container, columns, rows) {
  const wrap = document.createElement("div");
  wrap.className = "table-wrap";
  if (!rows || rows.length === 0) {
    wrap.innerHTML = `<div class="table-empty">暂无数据</div>`;
    container.appendChild(wrap);
    return wrap;
  }
  const table = document.createElement("table");
  table.className = "data";
  const thead = `<thead><tr>${columns
    .map((c) => `<th style="${c.align ? `text-align:${c.align}` : ""}">${esc(c.title)}</th>`)
    .join("")}</tr></thead>`;
  const tbody = document.createElement("tbody");
  rows.forEach((row) => {
    const tr = document.createElement("tr");
    columns.forEach((c) => {
      const td = document.createElement("td");
      if (c.align) td.style.textAlign = c.align;
      if (c.render) td.innerHTML = c.render(row);
      else td.textContent = row[c.key] ?? "";
      tr.appendChild(td);
    });
    tbody.appendChild(tr);
  });
  table.innerHTML = thead;
  table.appendChild(tbody);
  wrap.appendChild(table);
  container.appendChild(wrap);
  return wrap;
}

export function badge(text, variant) {
  return `<span class="badge badge-${variant}">${esc(text)}</span>`;
}

export function loading(container, text = "加载中…") {
  container.innerHTML = `<div class="table-empty">${esc(text)}</div>`;
}
