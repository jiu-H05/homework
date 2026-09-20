# -*- coding: utf-8 -*-
"""管理员控制台：图书管理、读者管理、借阅办理、归还续借、统计报表、操作日志。"""
import tkinter as tk
from tkinter import ttk

from .widgets import (BG, CARD_BG, PRIMARY, MUTED, TEXT, SUCCESS, DANGER, WARNING,
                      FONT_FAMILY, center_window, info, warn, error, confirm)
from .services import LibraryService, ServiceError


# ---------- 通用表格 ----------
def make_tree(parent, columns, headings, widths, anchors=None, height=12):
    container = ttk.Frame(parent)
    tv = ttk.Treeview(container, columns=columns, show="headings", height=height)
    if anchors is None:
        anchors = ["center"] * len(columns)
    for c, h, w, a in zip(columns, headings, widths, anchors):
        tv.heading(c, text=h)
        tv.column(c, width=w, anchor=a, stretch=True)
    vsb = ttk.Scrollbar(container, orient="vertical", command=tv.yview)
    hsb = ttk.Scrollbar(container, orient="horizontal", command=tv.xview)
    tv.configure(yscrollcommand=vsb.set, xscrollcommand=hsb.set)
    tv.grid(row=0, column=0, sticky="nsew")
    vsb.grid(row=0, column=1, sticky="ns")
    hsb.grid(row=1, column=0, sticky="ew")
    container.rowconfigure(0, weight=1)
    container.columnconfigure(0, weight=1)
    return container, tv


def selected_id(tv):
    sel = tv.selection()
    if not sel:
        return None
    return tv.item(sel[0])["values"][0]


# ---------- 图书编辑对话框 ----------
class BookDialog(tk.Toplevel):
    def __init__(self, parent, service, book=None):
        super().__init__(parent)
        self.title("修改图书" if book else "新增图书")
        self.configure(bg=BG)
        self.resizable(False, False)
        self.transient(parent)
        self.grab_set()
        self.service = service
        self.book = book
        self.result = None

        frm = ttk.Frame(self, padding=22)
        frm.pack()
        self.vars = {}
        fields = [("isbn", "ISBN"), ("title", "书名"), ("author", "作者"),
                  ("publisher", "出版社"), ("location", "馆藏位置")]
        for i, (key, label) in enumerate(fields):
            ttk.Label(frm, text=label).grid(row=i, column=0, sticky="w", pady=6, padx=(0, 12))
            v = tk.StringVar(value=book[key] if book else "")
            ttk.Entry(frm, textvariable=v, width=28).grid(row=i, column=1, pady=6)
            self.vars[key] = v

        ttk.Label(frm, text="分类").grid(row=5, column=0, sticky="w", pady=6, padx=(0, 12))
        self.cats = service.list_categories()
        self.cat_combo = ttk.Combobox(frm, width=26, state="readonly",
                                     values=[c["name"] for c in self.cats])
        cur_cat = book["category"] if book else ""
        if cur_cat:
            self.cat_combo.set(cur_cat)
        elif self.cats:
            self.cat_combo.current(0)
        self.cat_combo.grid(row=5, column=1, pady=6)

        ttk.Label(frm, text="馆藏数量").grid(row=6, column=0, sticky="w", pady=6, padx=(0, 12))
        self.var_qty = tk.StringVar(value=str(book["total_qty"] if book else 1))
        ttk.Spinbox(frm, from_=1, to=999, textvariable=self.var_qty, width=10).grid(row=6, column=1, sticky="w", pady=6)

        btns = ttk.Frame(frm)
        btns.grid(row=7, column=0, columnspan=2, pady=(18, 0))
        ttk.Button(btns, text="保存", style="Accent.TButton", command=self._ok).pack(side="left", padx=8)
        ttk.Button(btns, text="取消", command=self.destroy).pack(side="left", padx=8)
        x = parent.winfo_rootx() + 120
        y = parent.winfo_rooty() + 80
        self.geometry(f"+{x}+{y}")
        self.wait_window()

    def _cat_id(self):
        name = self.cat_combo.get()
        for c in self.cats:
            if c["name"] == name:
                return c["id"]
        return None

    def _ok(self):
        data = {k: v.get() for k, v in self.vars.items()}
        if not data["title"].strip():
            error("书名不能为空")
            return
        try:
            qty = int(self.var_qty.get())
        except ValueError:
            error("馆藏数量必须是整数")
            return
        if qty < 1:
            error("馆藏数量必须大于 0")
            return
        self.result = (data["isbn"], data["title"], data["author"], data["publisher"],
                       self._cat_id(), data["location"], qty)
        self.destroy()


# ---------- 标签页：图书管理 ----------
class BookTab(ttk.Frame):
    def __init__(self, parent, service):
        super().__init__(parent, padding=14)
        self.service = service

        bar = ttk.Frame(self)
        bar.pack(fill="x", pady=(0, 10))
        ttk.Label(bar, text="关键字：").pack(side="left")
        self.var_kw = tk.StringVar()
        e = ttk.Entry(bar, textvariable=self.var_kw, width=24)
        e.pack(side="left", padx=(0, 10))
        e.bind("<Return>", lambda ev: self.refresh())
        ttk.Label(bar, text="分类：").pack(side="left")
        self.cats = service.list_categories()
        self.cat_combo = ttk.Combobox(bar, width=12, state="readonly",
                                      values=["全部"] + [c["name"] for c in self.cats])
        self.cat_combo.current(0)
        self.cat_combo.pack(side="left", padx=(0, 12))
        self.cat_combo.bind("<<ComboboxSelected>>", lambda ev: self.refresh())
        ttk.Button(bar, text="查询", style="Accent.TButton", command=self.refresh).pack(side="left", padx=4)
        ttk.Button(bar, text="新增", command=self.add).pack(side="left", padx=4)
        ttk.Button(bar, text="修改", command=self.edit).pack(side="left", padx=4)
        ttk.Button(bar, text="删除", style="Danger.TButton", command=self.delete).pack(side="left", padx=4)

        container, self.tv = make_tree(
            self,
            ("id", "isbn", "title", "author", "publisher", "cat", "loc", "total", "avail"),
            ["ID", "ISBN", "书名", "作者", "出版社", "分类", "馆藏位置", "总藏", "在馆"],
            [40, 120, 220, 130, 150, 80, 90, 50, 50],
            ["center", "center", "w", "w", "w", "center", "center", "center", "center"],
            height=15,
        )
        container.pack(fill="both", expand=True)
        self.tv.column("id", width=0, stretch=False)
        self.refresh()

    def _selected_cat_id(self):
        name = self.cat_combo.get()
        for c in self.cats:
            if c["name"] == name:
                return c["id"]
        return None

    def refresh(self):
        for item in self.tv.get_children():
            self.tv.delete(item)
        rows = self.service.list_books(self.var_kw.get().strip(), self._selected_cat_id())
        for r in rows:
            self.tv.insert("", "end", values=(r["id"], r["isbn"], r["title"], r["author"],
                                              r["publisher"], r["category"], r["location"],
                                              r["total_qty"], r["available_qty"]))

    def add(self):
        dlg = BookDialog(self.winfo_toplevel(), self.service)
        if dlg.result:
            self.service.add_book(*dlg.result)
            self.service.log("图书管理", f"新增图书：{dlg.result[1]}")
            self.refresh()

    def edit(self):
        bid = selected_id(self.tv)
        if bid is None:
            warn("请先选择要修改的图书")
            return
        book = next((r for r in self.service.list_books() if r["id"] == bid), None)
        dlg = BookDialog(self.winfo_toplevel(), self.service, book=book)
        if dlg.result:
            self.service.update_book(bid, *dlg.result)
            self.service.log("图书管理", f"修改图书：{dlg.result[1]}")
            self.refresh()

    def delete(self):
        bid = selected_id(self.tv)
        if bid is None:
            warn("请先选择要删除的图书")
            return
        if not confirm("确定删除选中的图书吗？"):
            return
        try:
            self.service.delete_book(bid)
        except ServiceError as e:
            warn(str(e))
            return
        self.service.log("图书管理", f"删除图书ID {bid}")
        self.refresh()


# ---------- 标签页：读者管理 ----------
class ReaderTab(ttk.Frame):
    def __init__(self, parent, service):
        super().__init__(parent, padding=14)
        self.service = service

        bar = ttk.Frame(self)
        bar.pack(fill="x", pady=(0, 10))
        ttk.Label(bar, text="关键字：").pack(side="left")
        self.var_kw = tk.StringVar()
        e = ttk.Entry(bar, textvariable=self.var_kw, width=22)
        e.pack(side="left", padx=(0, 10))
        e.bind("<Return>", lambda ev: self.refresh())
        ttk.Button(bar, text="查询", style="Accent.TButton", command=self.refresh).pack(side="left", padx=4)
        ttk.Button(bar, text="新增读者", command=self.add).pack(side="left", padx=4)
        ttk.Button(bar, text="启用/禁用", command=self.toggle).pack(side="left", padx=4)
        ttk.Button(bar, text="重置口令", command=self.reset).pack(side="left", padx=4)

        container, self.tv = make_tree(
            self,
            ("id", "reader_no", "name", "phone", "username", "status", "created"),
            ["ID", "读者编号", "姓名", "联系电话", "登录用户名", "状态", "注册时间"],
            [40, 110, 120, 130, 130, 80, 160],
            ["center", "center", "center", "center", "center", "center", "center"],
            height=15,
        )
        container.pack(fill="both", expand=True)
        self.tv.column("id", width=0, stretch=False)
        self.refresh()

    def refresh(self):
        for item in self.tv.get_children():
            self.tv.delete(item)
        for r in self.service.list_readers(self.var_kw.get().strip()):
            self.tv.insert("", "end", values=(r["user_id"], r["reader_no"], r["name"], r["phone"],
                                              r["username"],
                                              "正常" if r["status"] == "active" else "禁用",
                                              r["created_at"]))

    def add(self):
        from .widgets import FormDialog
        dlg = FormDialog(self.winfo_toplevel(), "新增读者",
                         [("登录用户名", "", False), ("初始口令（至少6位）", "", True),
                          ("姓名", "", False), ("联系电话", "", False)])
        if not dlg.result:
            return
        username, pwd, name, phone = dlg.result
        try:
            no = self.service.register_reader(username, pwd, name, phone)
        except ServiceError as e:
            error(str(e))
            return
        self.service.log("读者管理", f"新增读者 {name}（{no}）")
        self.refresh()
        info(f"读者添加成功，读者编号：{no}")

    def _selected(self):
        sel = self.tv.selection()
        if not sel:
            return None
        return self.tv.item(sel[0])["values"]

    def toggle(self):
        row = self._selected()
        if row is None:
            warn("请先选择读者")
            return
        uid, _, name, _, _, status, _ = row
        new_status = "disabled" if status == "正常" else "active"
        if not confirm(f"确定{('禁用' if new_status=='disabled' else '启用')}读者 {name} 吗？"):
            return
        self.service.set_user_status(uid, new_status)
        self.service.log("读者管理", f"{name} 账号{new_status}")
        self.refresh()

    def reset(self):
        row = self._selected()
        if row is None:
            warn("请先选择读者")
            return
        uid, _, name, *_ = row
        if not confirm(f"确定将 {name} 的口令重置为 123456 吗？"):
            return
        self.service.reset_password(uid)
        self.service.log("读者管理", f"重置 {name} 的口令")
        info("口令已重置为 123456")


# ---------- 标签页：借阅办理 ----------
class BorrowTab(ttk.Frame):
    def __init__(self, parent, service):
        super().__init__(parent, padding=14)
        self.service = service

        card = tk.Frame(self, bg=CARD_BG, highlightbackground="#D9DEE4", highlightthickness=1)
        card.pack(fill="x", pady=(0, 12))
        frm = ttk.Frame(card, padding=14)
        frm.pack(fill="x")
        ttk.Label(frm, text="读者：").grid(row=0, column=0, padx=(0, 6))
        self.reader_combo = ttk.Combobox(frm, width=30, state="readonly")
        self.reader_combo.grid(row=0, column=1, padx=(0, 24))
        ttk.Label(frm, text="图书：").grid(row=0, column=2, padx=(0, 6))
        self.book_combo = ttk.Combobox(frm, width=42, state="readonly")
        self.book_combo.grid(row=0, column=3, padx=(0, 16))
        ttk.Button(frm, text="办理借出", style="Accent.TButton", command=self.do_borrow).grid(row=0, column=4)

        ttk.Label(self, text="当前全部在借记录", font=(FONT_FAMILY, 10, "bold")).pack(anchor="w", pady=(6, 6))
        container, self.tv = make_tree(
            self,
            ("id", "reader_no", "reader", "title", "borrow", "due", "renew", "state"),
            ["记录ID", "读者编号", "读者姓名", "图书名称", "借阅日期", "应还日期", "续借次数", "状态"],
            [60, 100, 90, 260, 110, 110, 70, 130],
            ["center", "center", "center", "w", "center", "center", "center", "center"],
            height=12,
        )
        container.pack(fill="both", expand=True)
        self.refresh()

    def refresh(self):
        self.readers = self.service.list_readers()
        self.reader_map = {f"{r['reader_no']} - {r['name']}": r["id"] for r in self.readers}
        self.reader_combo["values"] = list(self.reader_map.keys())
        self.books = self.service.list_books()
        self.book_map = {f"{b['title']}（在馆 {b['available_qty']}/{b['total_qty']}）": b["id"]
                         for b in self.books}
        self.book_combo["values"] = list(self.book_map.keys())
        for item in self.tv.get_children():
            self.tv.delete(item)
        for r in self.service.active_borrows():
            self.tv.insert("", "end", values=(r["id"], r["reader_no"], r["reader_name"], r["title"],
                                              r["borrow_date"], r["due_date"], r["renew_count"],
                                              self.service.record_state(r)))

    def do_borrow(self):
        r_text = self.reader_combo.get()
        b_text = self.book_combo.get()
        if not r_text or not b_text:
            warn("请选择读者与图书")
            return
        rid = self.reader_map[r_text]
        bid = self.book_map[b_text]
        try:
            self.service.borrow(rid, bid)
        except ServiceError as e:
            warn(str(e))
            return
        info("借阅办理成功")
        self.refresh()


# ---------- 标签页：归还/续借 ----------
class ReturnTab(ttk.Frame):
    def __init__(self, parent, service):
        super().__init__(parent, padding=14)
        self.service = service

        bar = ttk.Frame(self)
        bar.pack(fill="x", pady=(0, 10))
        ttk.Label(bar, text="读者筛选：").pack(side="left")
        self.reader_combo = ttk.Combobox(bar, width=32, state="readonly")
        self.reader_combo.pack(side="left", padx=(0, 12))
        self.reader_combo.bind("<<ComboboxSelected>>", lambda ev: self.refresh())
        ttk.Button(bar, text="显示全部", command=self.show_all).pack(side="left", padx=4)
        ttk.Button(bar, text="办理归还", style="Success.TButton", command=self.do_return).pack(side="left", padx=8)
        ttk.Button(bar, text="办理续借", style="Accent.TButton", command=self.do_renew).pack(side="left", padx=4)

        container, self.tv = make_tree(
            self,
            ("id", "reader_no", "reader", "title", "borrow", "due", "renew", "state"),
            ["记录ID", "读者编号", "读者姓名", "图书名称", "借阅日期", "应还日期", "续借次数", "状态"],
            [60, 100, 90, 260, 110, 110, 70, 140],
            ["center", "center", "center", "w", "center", "center", "center", "center"],
            height=15,
        )
        container.pack(fill="both", expand=True)
        self.readers = service.list_readers()
        self.reader_map = {f"{r['reader_no']} - {r['name']}": r["id"] for r in self.readers}
        self.reader_combo["values"] = ["全部读者"] + list(self.reader_map.keys())
        self.reader_combo.current(0)
        self.refresh()

    def show_all(self):
        self.reader_combo.current(0)
        self.refresh()

    def refresh(self):
        sel = self.reader_combo.get()
        rid = None
        if sel and sel != "全部读者":
            rid = self.reader_map.get(sel)
        for item in self.tv.get_children():
            self.tv.delete(item)
        for r in self.service.active_borrows(rid):
            self.tv.insert("", "end", values=(r["id"], r["reader_no"], r["reader_name"], r["title"],
                                              r["borrow_date"], r["due_date"], r["renew_count"],
                                              self.service.record_state(r)))

    def _record_id(self):
        return selected_id(self.tv)

    def do_return(self):
        rec_id = self._record_id()
        if rec_id is None:
            warn("请先选择要归还的借阅记录")
            return
        if not confirm("确定办理该图书的归还吗？"):
            return
        try:
            fine = self.service.return_book(rec_id)
        except ServiceError as e:
            warn(str(e))
            return
        self.service.log("归还", f"记录 {rec_id} 归还，罚款 {fine:.1f} 元")
        if fine > 0:
            warn(f"图书已逾期，归还成功，应缴罚款 {fine:.1f} 元")
        else:
            info("归还成功，无逾期罚款")
        self.refresh()

    def do_renew(self):
        rec_id = self._record_id()
        if rec_id is None:
            warn("请先选择要续借的借阅记录")
            return
        try:
            self.service.renew(rec_id)
        except ServiceError as e:
            warn(str(e))
            return
        info("续借成功，应还日期已延长 30 天")
        self.refresh()


# ---------- 标签页：统计报表 ----------
class DashboardTab(ttk.Frame):
    def __init__(self, parent, service):
        super().__init__(parent, padding=14)
        self.service = service

        cards = ttk.Frame(self)
        cards.pack(fill="x", pady=(0, 14))
        self.stat_labels = {}
        defs = [("book_kinds", "图书种类"), ("book_copies", "图书总册数"), ("readers", "注册读者"),
                ("borrowed", "当前借出"), ("overdue", "逾期数量")]
        for i, (key, name) in enumerate(defs):
            c = tk.Frame(cards, bg=CARD_BG, highlightbackground="#D9DEE4", highlightthickness=1)
            c.grid(row=0, column=i, padx=(0 if i == 0 else 10, 0), sticky="nsew")
            cards.columnconfigure(i, weight=1)
            v = tk.Label(c, text="0", bg=CARD_BG, fg=PRIMARY if key != "overdue" else DANGER,
                         font=(FONT_FAMILY, 20, "bold"))
            v.pack(pady=(14, 2))
            tk.Label(c, text=name, bg=CARD_BG, fg=MUTED, font=(FONT_FAMILY, 9)).pack(pady=(0, 12))
            self.stat_labels[key] = v

        body = ttk.Frame(self)
        body.pack(fill="both", expand=True)

        # 左：分类统计图
        left = tk.Frame(body, bg=CARD_BG, highlightbackground="#D9DEE4", highlightthickness=1)
        left.pack(side="left", fill="both", expand=True, padx=(0, 10))
        tk.Label(left, text="各分类馆藏册数", bg=CARD_BG, fg=TEXT,
                 font=(FONT_FAMILY, 11, "bold")).pack(anchor="w", padx=14, pady=(12, 4))
        self.canvas = tk.Canvas(left, bg=CARD_BG, highlightthickness=0, height=240)
        self.canvas.pack(fill="both", expand=True, padx=10, pady=(0, 10))
        self.canvas.bind("<Configure>", lambda e: self._draw_chart())

        # 右：热门图书
        right = tk.Frame(body, bg=CARD_BG, highlightbackground="#D9DEE4", highlightthickness=1)
        right.pack(side="left", fill="both", expand=True)
        tk.Label(right, text="图书借阅热度", bg=CARD_BG, fg=TEXT,
                 font=(FONT_FAMILY, 11, "bold")).pack(anchor="w", padx=14, pady=(12, 4))
        c2, self.tv = make_tree(
            right, ("title", "author", "times"), ["书名", "作者", "借阅次数"],
            [240, 140, 80], ["w", "w", "center"], height=10,
        )
        c2.pack(fill="both", expand=True, padx=10, pady=(0, 10))
        self.refresh()

    def refresh(self):
        stats = self.service.dashboard()
        for k, v in stats.items():
            self.stat_labels[k].configure(text=str(v))
        for i in self.tv.get_children():
            self.tv.delete(i)
        for r in self.service.hot_books():
            self.tv.insert("", "end", values=(r["title"], r["author"], r["times"]))
        self._draw_chart()

    def _draw_chart(self):
        c = self.canvas
        c.delete("all")
        rows = self.service.category_stats()
        max_v = max([r["copies"] for r in rows] + [1])
        c.update_idletasks()
        w = c.winfo_width()
        if w < 120:
            return  # 画布尚未布局，等待 <Configure> 后重绘
        bar_h = 22
        gap = 12
        top = 10
        label_w = 70
        for i, r in enumerate(rows):
            y = top + i * (bar_h + gap)
            c.create_text(8, y + bar_h / 2, text=r["name"], anchor="w",
                          fill=TEXT, font=(FONT_FAMILY, 9))
            bw = max(int((w - label_w - 46) * r["copies"] / max_v), 2)
            c.create_rectangle(label_w, y, label_w + bw, y + bar_h,
                               fill=PRIMARY, outline="")
            c.create_text(label_w + bw + 6, y + bar_h / 2, text=str(r["copies"]),
                          anchor="w", fill=MUTED, font=(FONT_FAMILY, 9))


# ---------- 标签页：操作日志 ----------
class LogsTab(ttk.Frame):
    def __init__(self, parent, service):
        super().__init__(parent, padding=14)
        self.service = service
        bar = ttk.Frame(self)
        bar.pack(fill="x", pady=(0, 10))
        ttk.Button(bar, text="刷新", style="Accent.TButton", command=self.refresh).pack(side="left")
        container, self.tv = make_tree(
            self, ("id", "time", "username", "action", "detail"),
            ["ID", "时间", "操作人", "操作", "详情"],
            [50, 150, 100, 120, 360],
            ["center", "center", "center", "center", "w"], height=16,
        )
        container.pack(fill="both", expand=True)
        self.refresh()

    def refresh(self):
        for i in self.tv.get_children():
            self.tv.delete(i)
        for r in self.service.recent_logs():
            self.tv.insert("", "end", values=(r["id"], r["created_at"], r["username"],
                                              r["action"], r["detail"]))


# ---------- 管理员主面板 ----------
class AdminPanel(ttk.Frame):
    def __init__(self, root, service, on_logout):
        super().__init__(root)
        self.pack(fill="both", expand=True)
        root.title("图书管理系统 - 管理员控制台")
        center_window(root, 1100, 700)

        top = ttk.Frame(self, padding=(16, 10))
        top.pack(fill="x")
        ttk.Label(top, text="管理员控制台", font=(FONT_FAMILY, 14, "bold")).pack(side="left")
        ttk.Label(top, text=f"欢迎您，{service.current_user['real_name'] or service.current_user['username']}",
                  foreground=MUTED).pack(side="left", padx=16)
        ttk.Button(top, text="退出登录", command=on_logout).pack(side="right")

        nb = ttk.Notebook(self)
        nb.pack(fill="both", expand=True, padx=12, pady=(0, 12))
        nb.add(BookTab(nb, service), text="图书管理")
        nb.add(ReaderTab(nb, service), text="读者管理")
        nb.add(BorrowTab(nb, service), text="借阅办理")
        nb.add(ReturnTab(nb, service), text="归还 / 续借")
        nb.add(DashboardTab(nb, service), text="统计报表")
        nb.add(LogsTab(nb, service), text="操作日志")
