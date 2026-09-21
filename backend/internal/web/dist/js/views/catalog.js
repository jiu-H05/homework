// catalog.js — 读者馆藏浏览：检索、题材/地区筛选、自助借阅。
import { api } from "../api.js";
import { sortBooks, BOOK_SORTS } from "../sort.js";
import { esc, loading, renderTable, toast } from "../components.js";

export async function renderCatalog(container) {
  loading(container);
  const categories = await api.listCategories();
  const regions = await api.listRegions();
  const wrap = document.createElement("div");
  wrap.innerHTML = `
    <div class="toolbar">
      <input class="input" id="kw" placeholder="搜索书名 / 作者 / ISBN" />
      <select class="input" id="cat" style="min-width:140px">
        <option value="">全部分类</option>
        ${categories.map((c) => `<option value="${c.id}">${esc(c.name)}</option>`).join("")}
      </select>
      <select class="input" id="region" style="min-width:170px">
        <option value="">全部地区</option>
        ${regions.map((r) => `<option value="${esc(r.name)}">${esc(r.name)}（${r.kinds}）</option>`).join("")}
      </select>
      <select class="input" id="sort" style="min-width:170px">
        ${BOOK_SORTS.map((s) => `<option value="${s.value}">${esc(s.label)}</option>`).join("")}
      </select>
      <button class="btn btn-primary btn-sm" id="search">查询</button>
    </div>
    <div id="table"></div>`;
  container.innerHTML = "";
  container.appendChild(wrap);
  const tableEl = wrap.querySelector("#table");

  const load = async () => {
    loading(tableEl);
    let books = await api.listBooks({
      keyword: wrap.querySelector("#kw").value.trim(),
      category: wrap.querySelector("#cat").value,
      region: wrap.querySelector("#region").value,
    });
    books = sortBooks(books, wrap.querySelector("#sort").value);
    tableEl.innerHTML = "";
    renderTable(
      tableEl,
      [
        { key: "title", title: "书名", render: (r) =>
          `<div>${esc(r.title)}</div><div class="muted" style="font-size:12px">${esc(r.author)} · ${esc(r.publisher)}</div>` },
        { key: "category", title: "分类" },
        { key: "region", title: "地区" },
        { key: "location", title: "馆藏位置" },
        {
          key: "availableQty", title: "可借/总册", align: "center",
          render: (r) => `${r.availableQty} / ${r.totalQty}`,
        },
        {
          title: "操作", align: "right",
          render: (r) => `<div class="row-actions">
            <button class="btn ${r.availableQty > 0 ? "btn-primary" : "btn-ghost"} btn-sm act-borrow"
              ${r.availableQty > 0 ? "" : "disabled"}>
              ${r.availableQty > 0 ? "借阅" : "已借完"}</button></div>`,
        },
      ],
      books
    );

    tableEl.querySelectorAll(".act-borrow").forEach((b) =>
      b.addEventListener("click", async () => {
        const book = books[Number(b.closest("tr").rowIndex) - 1];
        b.disabled = true;
        try {
          await api.selfBorrow(book.id);
          toast(`《${book.title}》借阅成功`, "success");
          load();
        } catch (err) { toast(err.message, "error"); b.disabled = false; }
      })
    );
  };

  wrap.querySelector("#search").onclick = load;
  wrap.querySelector("#sort").onchange = load;
  wrap.querySelector("#region").onchange = load;
  wrap.querySelector("#cat").onchange = load;
  wrap.querySelector("#kw").addEventListener("keydown", (e) => { if (e.key === "Enter") load(); });
  load();
}
