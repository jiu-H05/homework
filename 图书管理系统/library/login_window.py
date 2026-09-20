# -*- coding: utf-8 -*-
"""登录与读者注册界面。"""
import tkinter as tk
from tkinter import ttk

from .widgets import (BG, CARD_BG, PRIMARY, MUTED, TEXT, FONT_FAMILY,
                      center_window, info, error, FormDialog)
from .services import LibraryService, ServiceError


class LoginFrame(ttk.Frame):
    def __init__(self, root: tk.Tk, service: LibraryService, on_success):
        super().__init__(root)
        self.root = root
        self.service = service
        self.on_success = on_success
        self.pack(fill="both", expand=True)
        root.title("图书管理系统 - 登录")
        center_window(root, 900, 580)
        self._build()

    def _build(self):
        # 左侧品牌区
        left = tk.Frame(self, bg=PRIMARY, width=380)
        left.pack(side="left", fill="y")
        left.pack_propagate(False)
        tk.Label(left, text="图书管理系统", bg=PRIMARY, fg="#FFFFFF",
                 font=(FONT_FAMILY, 26, "bold")).pack(pady=(130, 10), padx=36, anchor="w")
        tk.Label(left, text="Library Management System", bg=PRIMARY, fg="#CFE3F4",
                 font=(FONT_FAMILY, 11)).pack(padx=38, anchor="w")
        tk.Label(left, text="图书检索 · 借阅 · 归还 · 续借 · 统计管理", bg=PRIMARY, fg="#CFE3F4",
                 font=(FONT_FAMILY, 10)).pack(pady=28, padx=38, anchor="w")

        # 右侧登录卡片
        right = ttk.Frame(self)
        right.pack(side="left", fill="both", expand=True, padx=60)

        card = tk.Frame(right, bg=CARD_BG, highlightbackground="#D9DEE4",
                        highlightthickness=1)
        card.place(relx=0.5, rely=0.5, anchor="center", width=360, height=400)

        tk.Label(card, text="欢迎登录", bg=CARD_BG, fg=TEXT,
                 font=(FONT_FAMILY, 18, "bold")).pack(pady=(38, 4))
        tk.Label(card, text="请使用您的账号登录系统", bg=CARD_BG, fg=MUTED,
                 font=(FONT_FAMILY, 9)).pack(pady=(0, 22))

        form = tk.Frame(card, bg=CARD_BG)
        form.pack()
        tk.Label(form, text="用户名", bg=CARD_BG, fg=TEXT, font=(FONT_FAMILY, 9)).grid(row=0, column=0, sticky="w")
        self.var_user = tk.StringVar()
        e_user = ttk.Entry(form, textvariable=self.var_user, width=28, font=(FONT_FAMILY, 11))
        e_user.grid(row=1, column=0, pady=(4, 14), ipady=3)

        tk.Label(form, text="口令", bg=CARD_BG, fg=TEXT, font=(FONT_FAMILY, 9)).grid(row=2, column=0, sticky="w")
        self.var_pwd = tk.StringVar()
        e_pwd = ttk.Entry(form, textvariable=self.var_pwd, show="*", width=28, font=(FONT_FAMILY, 11))
        e_pwd.grid(row=3, column=0, pady=(4, 18), ipady=3)

        ttk.Button(form, text="登 录", style="Accent.TButton", width=26, command=self._login).grid(row=4, column=0)
        ttk.Button(form, text="读者自助注册", width=26, command=self._register).grid(row=5, column=0, pady=(10, 0))

        tk.Label(card, text="默认管理员：admin / admin123    示例读者：reader / reader123",
                 bg=CARD_BG, fg=MUTED, font=(FONT_FAMILY, 8)).pack(side="bottom", pady=14)

        e_user.focus_set()
        e_pwd.bind("<Return>", lambda e: self._login())
        e_user.bind("<Return>", lambda e: e_pwd.focus_set())

    def _login(self):
        username = self.var_user.get().strip()
        password = self.var_pwd.get()
        if not username or not password:
            error("请输入用户名和口令")
            return
        try:
            self.service.authenticate(username, password)
        except ServiceError as e:
            error(str(e))
            return
        self.on_success()

    def _register(self):
        dlg = FormDialog(
            self.root, "读者自助注册",
            [("用户名", "", False), ("口令（至少6位）", "", True),
             ("确认口令", "", True), ("姓名", "", False), ("联系电话", "", False)],
        )
        if not dlg.result:
            return
        username, pwd, pwd2, name, phone = dlg.result
        if pwd != pwd2:
            error("两次输入的口令不一致")
            return
        try:
            reader_no = self.service.register_reader(username, pwd, name, phone)
        except ServiceError as e:
            error(str(e))
            return
        info(f"注册成功！您的读者编号为 {reader_no}\n请使用新账号登录。")
