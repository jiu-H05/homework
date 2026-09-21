// books.js — 管理员图书管理：检索、题材/地区筛选、新增、编辑、删除，以及分类维护。
import { api } from "../api.js";
import { sortBooks, BOOK_SORTS } from "../sort.js";
import {
  esc, loading, renderTable, openModal, confirmDialog, toast, formValues, fillForm, badge,
} from "../components.js";

let categories = [];
let regions = [];

function categoryName(id) {
  if (!id) return "未分类";
  const c = categories.find((x) => x.id === id);
  return c ? c.name : "未分类";
}

function bookForm(book) {
  const form = document.createElement("form");
  const catOptions = categories
    .map((c) => `<option value="${c.id}">${esc(c.name)}</option>`)
    .join("");
  const regionOptions = regions
    .map((r) => `<option value="${esc(r.name)}">${esc(r.name)}</option>`)
    .join("");
  form.innerHTML = `
    <div class="form-row">
      <div class="field"><label>ISBN</label><input class="input" name="isbn" /></div>
      <div class="field"><label>分类</label>
        <select class="input" name="categoryId"><option value="">未分类</option>${catOptions}</select></div>
    </div>
    <div class="field"><label>书名</label><input class="input" name="title" required /></div>
    <div class="form-row">
      <div class="field"><label>作者</label><input class="input" name="author" /></div>
      <div class="field"><label>出版社</label><input class="input" name="publisher" /></div>
    </div>
    <div class="form-row">
      <div class="field"><label>地区</label>
        <select class="input" name="region"><option value="">未指定</option>${regionOptions}</select></div>
      <div class="field"><label>馆藏位置</label><input class="input" name="location" /></div>
    </div>
    <div class="field"><label>总册数</label><input class="input" type="number" name="totalQty" min="1" value="1" required /></div>`;
  if (book) {
    fillForm(form, {
      ...book,
      categoryId: book.categoryId ?? "",
      region: book.region ?? "",
    });
  }
  return form;
}

async function openBookEditor(book, reload) {
  const isEdit = !!book;
  const form = bookForm(book);
  const footer = document.createElement("div");
  const btnCancel = document.createElement("button");
  btnCancel.className = "btn btn-ghost";
  btnCancel.textContent = "取消";
  const btnSave = document.createElement("button");
  btnSave.className = "btn btn-primary";
  btnSave.textContent = "保存";
  footer.append(btnCancel, btnSave);

  const m = openModal({ title: isEdit ? "编辑图书" : "新增图书", body: form, footer });
  btnCancel.onclick = m.close;
  btnSave.onclick = async () => {
    const v = formValues(form);
    if (!v.title) { toast("请填写书名", "error"); return; }
    v.totalQty = Number(v.totalQty) || 1;
    v.categoryId = v.categoryId ? Number(v.categoryId) : null;
    btnSave.disabled = true;
    try {
      if (isEdit) {
        v.availableQty = book.availableQty;
        await api.updateBook(book.id, v);
        toast("图书已更新", "success");
      } else {
        v.availableQty = v.totalQty;
        await api.createBook(v);
        toast("图书已新增", "success");
      }
      m.close();
      reload();
    } catch (err) {
      toast(err.message, "error");
      btnSave.disabled = false;
    }
  };
}

async function openCategoryManager(reload) {
  const body = document.createElement("div");
  body.innerHTML = `
    <div class="toolbar">
      <input class="input" id="new-cat" placeholder="输入新分类名称" />
      <button class="btn btn-primary btn-sm" id="add-cat">新增分类</button>
    </div>
    <div class="table-wrap" id="cat-list"></div>`;
  const listEl = body.querySelector("#cat-list");
  listEl.innerHTML = `<table class="data"><thead><tr><th>分类名称</th><th style="text-align:right">编号</th></tr></thead>
    <tbody>${categories.map((c) => `<tr><td>${esc(c.name)}</td><td style="text-align:right">${c.id}</td></tr>`).join("")}</tbody></table>`;

  const m = openModal({ title: "分类管理", body, size: "lg" });
  body.querySelector("#add-cat").onclick = async () => {
    const name = body.querySelector("#new-cat").value.trim();
    if (!name) { toast("请填写分类名称", "error"); return; }
    try {
      await api.addCategory(name);
      toast("分类已新增", "success");
      m.close();
      reload();
    } catch (err) { toast(err.message, "error"); }
  };
}

export async function renderBooks(container) {
  loading(container);
  categories = await api.listCategories();
  regions = await api.listRegions();

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
      <div class="spacer"></div>
      <button class="btn btn-outline btn-sm" id="manage-cat">分类管理</button>
      <button class="btn btn-primary btn-sm" id="add-book">+ 新增图书</button>
    </div>
    <div id="table"></div>`;
  container.innerHTML = "";
  container.appendChild(wrap);

  const tableEl = wrap.querySelector("#table");
  const load = async () => {
    loading(tableEl);
    const keyword = wrap.querySelector("#kw").value.trim();
    const category = wrap.querySelector("#cat").value;
    const region = wrap.querySelector("#region").value;
    let books = await api.listBooks({ keyword, category, region });
    books = sortBooks(books, wrap.querySelector("#sort").value);
    tableEl.innerHTML = "";
    renderTable(
      tableEl,
      [
        { key: "title", title: "书名", render: (r) =>
          `<div>${esc(r.title)}</div><div class="muted" style="font-size:12px">${esc(r.isbn)}</div>` },
        { key: "author", title: "作者" },
        { key: "category", title: "分类", render: (r) => badge(categoryName(r.categoryId), "blue") },
        { key: "region", title: "地区", render: (r) => esc(r.region) || "—" },
        { key: "location", title: "位置" },
        {
          key: "availableQty", title: "可借/总册", align: "center",
          render: (r) => `${r.availableQty} / ${r.totalQty}`,
        },
        {
          title: "操作", align: "right",
          render: (r) => `
            <div class="row-actions">
              <button class="btn btn-ghost btn-sm act-edit">编辑</button>
              <button class="btn btn-danger btn-sm act-del">删除</button>
            </div>`,
        },
      ],
      books
    );

    tableEl.querySelectorAll(".act-edit").forEach((b) =>
      b.addEventListener("click", () => {
        const book = books[Number(b.closest("tr").rowIndex) - 1];
        openBookEditor(book, () => renderBooks(container));
      })
    );
    tableEl.querySelectorAll(".act-del").forEach((b) =>
      b.addEventListener("click", async () => {
        const book = books[Number(b.closest("tr").rowIndex) - 1];
        if (await confirmDialog(`确定删除《${book.title}》吗？存在未还借阅时将无法删除。`)) {
          try {
            await api.deleteBook(book.id);
            toast("图书已删除", "success");
            renderBooks(container);
          } catch (err) { toast(err.message, "error"); }
        }
      })
    );
  };

  wrap.querySelector("#search").onclick = load;
  wrap.querySelector("#sort").onchange = load;
  wrap.querySelector("#region").onchange = load;
  wrap.querySelector("#cat").onchange = load;
  wrap.querySelector("#kw").addEventListener("keydown", (e) => { if (e.key === "Enter") load(); });
  wrap.querySelector("#add-book").onclick = () => openBookEditor(null, () => renderBooks(container));
  wrap.querySelector("#manage-cat").onclick = () => openCategoryManager(() => renderBooks(container));

  load();
}
