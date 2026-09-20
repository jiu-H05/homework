# -*- coding: utf-8 -*-
"""
图书管理系统（Library Management System）
程序入口：初始化数据库 -> 启动登录界面 -> 按角色进入管理员控制台 / 读者中心。

运行：python main.py
"""
import tkinter as tk

from library.database import init_db
from library.services import LibraryService
from library.widgets import apply_style
from library.login_window import LoginFrame
from library.admin_window import AdminPanel
from library.reader_window import ReaderPanel


class App(tk.Tk):
    def __init__(self):
        super().__init__()
        apply_style(self)
        self.service = LibraryService()
        self.show_login()

    def _clear(self):
        for w in self.winfo_children():
            w.destroy()

    def show_login(self):
        self._clear()
        LoginFrame(self, self.service, self.on_login)

    def on_login(self):
        self._clear()
        if self.service.current_user["role"] == "admin":
            AdminPanel(self, self.service, self.show_login)
        else:
            ReaderPanel(self, self.service, self.show_login)


def main():
    init_db()
    app = App()
    app.mainloop()


if __name__ == "__main__":
    main()
