# -*- coding: utf-8 -*-
"""读者端：图书检索与借阅、我的借阅（归还状态/续借）、个人信息。"""
import tkinter as tk
from tkinter import ttk

from .widgets import (BG, CARD_BG, PRIMARY, MUTED, TEXT, FONT_FAMILY,
                      info, warn, error, FormDialog)
from .services import LibraryService, ServiceError
from .admin_window import make_tree


class SearchTab(ttk.Frame):
    """图书检索与自助借阅。"""
    def __init__(self, parent, service, reader):
        super().__init__(parent, padding=14)
        self.service = service
        self.reader = reader

        bar = ttk.Frame(self)
        bar.pack(fill="x", pady=(0, 10))
        ttk.Label(bar, text="关键字：").pack(side="left")
        self.var_kw = tk.StringVar()
        e = ttk.Entry(bar, textvariable=self.var_kw, width=26)
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
        ttk.Button(bar, text="借阅此书", style="Success.TButton", command=self.borrow).pack(side="left", padx=10)

        container, self.tv = make_tree(
            self,
            ("id", "isbn", "title", "author", "publisher", "cat", "loc", "avail", "total"),
            ["ID", "ISBN", "书名", "作者", "出版社", "分类", "馆藏位置", "在馆", "总藏"],
            [40, 120, 230, 140, 150, 80, 90, 50, 50],
            ["center", "center", "w", "w", "w", "center", "center", "center", "center"],
            height=15,
        )
        container.pack(fill="both", expand=True)
        self.tv.column("id", width=0, stretch=False)
        self.refresh()

    def _cat_id(self):
        name = self.cat_combo.get()
        for c in self.cats:
            if c["name"] == name:
                return c["id"]
        return None

    def refresh(self):
        for i in self.tv.get_children():
            self.tv.delete(i)
        rows = self.service.list_books(self.var_kw.get().strip(), self._cat_id())
        for r in rows:
            self.tv.insert("", "end", values=(r["id"], r["isbn"], r["title"], r["author"], r["publisher"],
                                              r["category"], r["location"], r["available_qty"], r["total_qty"]))

    def borrow(self):
        sel = self.tv.selection()
        if not sel:
            warn("请先选择要借阅的图书")
            return
        values = self.tv.item(sel[0])["values"]
        bid, title, avail = values[0], values[2], values[7]
        if avail <= 0:
            warn("该图书暂无在馆库存，无法借阅")
            return
        if not confirm(f"确定借阅《{title}》吗？借期 30 天。"):
            return
        try:
            self.service.borrow(self.reader["id"], bid)
        except ServiceError as e:
            warn(str(e))
            return
        info("借阅成功，可在“我的借阅”中查看")
        self.refresh()


class MyBorrowTab(ttk.Frame):
    """当前在借（含续借）与历史记录。"""
    def __init__(self, parent, service, reader):
        super().__init__(parent, padding=14)
        self.service = service
        self.reader = reader

        ttk.Label(self, text="当前在借图书", font=(FONT_FAMILY, 11, "bold")).pack(anchor="w")
        bar = ttk.Frame(self)
        bar.pack(fill="x", pady=6)
        ttk.Button(bar, text="刷新", style="Accent.TButton", command=self.refresh).pack(side="left", padx=4)
        ttk.Button(bar, text="续借此书", command=self.renew).pack(side="left", padx=8)

        c1, self.tv = make_tree(
            self,
            ("id", "title", "author", "borrow", "due", "renew", "state"),
            ["记录ID", "书名", "作者", "借阅日期", "应还日期", "续借次数", "状态"],
            [60, 260, 140, 110, 110, 70, 150],
            ["center", "w", "w", "center", "center", "center", "center"],
            height=8,
        )
        c1.pack(fill="x")

        ttk.Label(self, text="历史借阅记录", font=(FONT_FAMILY, 11, "bold")).pack(anchor="w", pady=(16, 6))
        c2, self.tv2 = make_tree(
            self,
            ("id", "title", "borrow", "returned", "fine"),
            ["记录ID", "书名", "借阅日期", "归还日期", "罚款（元）"],
            [60, 300, 130, 130, 100],
            ["center", "w", "center", "center", "center"],
            height=7,
        )
        c2.pack(fill="both", expand=True)
        self.refresh()

    def refresh(self):
        for i in self.tv.get_children():
            self.tv.delete(i)
        for r in self.service.active_borrows(self.reader["id"]):
            self.tv.insert("", "end", values=(r["id"], r["title"], r["author"], r["borrow_date"],
                                              r["due_date"], r["renew_count"],
                                              self.service.record_state(r)))
        for i in self.tv2.get_children():
            self.tv2.delete(i)
        for r in self.service.history(self.reader["id"]):
            self.tv2.insert("", "end", values=(r["id"], r["title"], r["borrow_date"],
                                               r["return_date"], f"{r['fine']:.1f}"))

    def renew(self):
        sel = self.tv.selection()
        if not sel:
            warn("请先选择要续借的图书")
            return
        rec_id = self.tv.item(sel[0])["values"][0]
        try:
            self.service.renew(rec_id)
        except ServiceError as e:
            warn(str(e))
            return
        info("续借成功，应还日期延长 30 天")
        self.refresh()


class ProfileTab(ttk.Frame):
    """个人信息与口令修改。"""
    def __init__(self, parent, service, reader):
        super().__init__(parent, padding=14)
        self.service = service
        self.reader = reader

        card = tk.Frame(self, bg=CARD_BG, highlightbackground="#D9DEE4", highlightthickness=1)
        card.pack(fill="x", pady=10)
        frm = ttk.Frame(card, padding=20)
        frm.pack(fill="x")
        ttk.Label(frm, text="我的信息", style="Card.TLabel",
                  font=(FONT_FAMILY, 13, "bold")).grid(row=0, column=0, columnspan=2, sticky="w", pady=(0, 12))
        info_rows = [("读者编号", reader["reader_no"]), ("姓名", reader["name"]),
                     ("联系电话", reader["phone"] or "未填写"), ("注册时间", reader["created_at"])]
        for i, (k, v) in enumerate(info_rows, start=1):
            tk.Label(frm, text=k, bg=CARD_BG, fg=MUTED, font=(FONT_FAMILY, 10)).grid(
                row=i, column=0, sticky="w", pady=6, padx=(0, 40))
            tk.Label(frm, text=v, bg=CARD_BG, fg=TEXT, font=(FONT_FAMILY, 10)).grid(
                row=i, column=1, sticky="w", pady=6)

        ttk.Button(self, text="修改口令", style="Accent.TButton", command=self.change_pwd).pack(anchor="w", pady=14)

    def change_pwd(self):
        dlg = FormDialog(self.winfo_toplevel(), "修改口令",
                         [("原口令", "", True), ("新口令（至少6位）", "", True),
                          ("确认新口令", "", True)])
        if not dlg.result:
            return
        old, new, new2 = dlg.result
        if new != new2:
            warn("两次输入的新口令不一致")
            return
        try:
            self.service.change_password(self.service.current_user["id"], old, new)
        except ServiceError as e:
            warn(str(e))
            return
        info("口令修改成功")


class ReaderPanel(ttk.Frame):
    def __init__(self, root, service, on_logout):
        super().__init__(root)
        self.pack(fill="both", expand=True)
        root.title("图书管理系统 - 读者中心")
        from .widgets import center_window
        center_window(root, 1000, 660)

        reader = service.get_reader_by_user(service.current_user["id"])

        top = ttk.Frame(self, padding=(16, 10))
        top.pack(fill="x")
        ttk.Label(top, text="读者中心", font=(FONT_FAMILY, 14, "bold")).pack(side="left")
        ttk.Label(top, text=f"欢迎您，{reader['name']}（{reader['reader_no']}）",
                  foreground=MUTED).pack(side="left", padx=16)
        ttk.Button(top, text="退出登录", command=on_logout).pack(side="right")

        nb = ttk.Notebook(self)
        nb.pack(fill="both", expand=True, padx=12, pady=(0, 12))
        nb.add(SearchTab(nb, service, reader), text="图书检索 / 借阅")
        nb.add(MyBorrowTab(nb, service, reader), text="我的借阅")
        nb.add(ProfileTab(nb, service, reader), text="个人信息")
