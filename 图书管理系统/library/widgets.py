# -*- coding: utf-8 -*-
"""UI 共享模块：主题样式、通用对话框与小工具。"""
import tkinter as tk
from tkinter import ttk, messagebox

# 配色（中性蓝灰，专业风格）
PRIMARY = "#1F6FB2"
PRIMARY_DARK = "#175A92"
BG = "#F4F6F8"
CARD_BG = "#FFFFFF"
BORDER = "#D9DEE4"
TEXT = "#1F2933"
MUTED = "#6B7682"
SUCCESS = "#2E8B57"
DANGER = "#C0392B"
WARNING = "#D68910"

FONT_FAMILY = "Microsoft YaHei UI"
FONT_NORMAL = (FONT_FAMILY, 10)
FONT_BOLD = (FONT_FAMILY, 10, "bold")
FONT_TITLE = (FONT_FAMILY, 17, "bold")
FONT_SUBTITLE = (FONT_FAMILY, 11)
FONT_SMALL = (FONT_FAMILY, 9)


def apply_style(root):
    """配置全局 ttk 样式。"""
    style = ttk.Style(root)
    try:
        style.theme_use("clam")
    except tk.TclError:
        pass
    root.configure(bg=BG)

    style.configure(".", font=FONT_NORMAL, background=BG, foreground=TEXT)
    style.configure("TFrame", background=BG)
    style.configure("Card.TFrame", background=CARD_BG)
    style.configure("TLabel", background=BG, foreground=TEXT)
    style.configure("Card.TLabel", background=CARD_BG, foreground=TEXT)
    style.configure("Muted.TLabel", background=BG, foreground=MUTED, font=FONT_SMALL)
    style.configure("Card.Muted.TLabel", background=CARD_BG, foreground=MUTED, font=FONT_SMALL)
    style.configure("Title.TLabel", background=BG, foreground=TEXT, font=FONT_TITLE)
    style.configure("Subtitle.TLabel", background=BG, foreground=MUTED, font=FONT_SUBTITLE)
    style.configure("Stat.TLabel", background=CARD_BG, foreground=PRIMARY, font=(FONT_FAMILY, 20, "bold"))
    style.configure("StatName.TLabel", background=CARD_BG, foreground=MUTED, font=FONT_SMALL)

    # 按钮
    style.configure("TButton", padding=(12, 6), font=FONT_NORMAL)
    style.configure("Accent.TButton", background=PRIMARY, foreground="#FFFFFF",
                    bordercolor=PRIMARY, focuscolor=PRIMARY_DARK, padding=(14, 7), font=FONT_BOLD)
    style.map("Accent.TButton",
              background=[("active", PRIMARY_DARK), ("pressed", PRIMARY_DARK)],
              foreground=[("active", "#FFFFFF")])
    style.configure("Danger.TButton", background=DANGER, foreground="#FFFFFF",
                    bordercolor=DANGER, focuscolor=DANGER, padding=(12, 6))
    style.map("Danger.TButton", background=[("active", "#A93226")], foreground=[("active", "#FFFFFF")])
    style.configure("Success.TButton", background=SUCCESS, foreground="#FFFFFF",
                    bordercolor=SUCCESS, focuscolor=SUCCESS, padding=(12, 6))
    style.map("Success.TButton", background=[("active", "#247147")], foreground=[("active", "#FFFFFF")])

    # 输入框
    style.configure("TEntry", fieldbackground="#FFFFFF", padding=4)
    style.configure("TCombobox", fieldbackground="#FFFFFF", padding=4)

    # 选项卡
    style.configure("TNotebook", background=BG, borderwidth=0)
    style.configure("TNotebook.Tab", padding=(18, 9), font=FONT_NORMAL, background="#E7EBF0", foreground=MUTED)
    style.map("TNotebook.Tab",
              background=[("selected", BG)],
              foreground=[("selected", PRIMARY), ("active", TEXT)])

    # 表格
    style.configure("Treeview", background=CARD_BG, fieldbackground=CARD_BG, foreground=TEXT,
                    rowheight=30, font=FONT_NORMAL, bordercolor=BORDER)
    style.configure("Treeview.Heading", font=FONT_BOLD, background="#E7EBF0", foreground=TEXT, padding=(6, 7))
    style.map("Treeview.Heading", background=[("active", "#D8DEE5")])
    style.map("Treeview", background=[("selected", "#DCEAF6")], foreground=[("selected", TEXT)])

    # 滚动条
    style.configure("Vertical.TScrollbar", background=BORDER, troughcolor=BG)
    style.configure("Horizontal.TScrollbar", background=BORDER, troughcolor=BG)


def center_window(win, width, height):
    win.update_idletasks()
    sw = win.winfo_screenwidth()
    sh = win.winfo_screenheight()
    x = int((sw - width) / 2)
    y = int((sh - height) / 2.4)
    win.geometry(f"{width}x{height}+{max(x,0)}+{max(y,0)}")
    win.minsize(width, height)


def info(msg, title="提示"):
    messagebox.showinfo(title, msg)


def warn(msg, title="注意"):
    messagebox.showwarning(title, msg)


def error(msg, title="错误"):
    messagebox.showerror(title, msg)


def confirm(msg, title="请确认"):
    return messagebox.askyesno(title, msg)


class FormDialog(tk.Toplevel):
    """通用表单对话框：字段列表 [(标签, 默认值, 是否口令)]，确定后返回值列表。"""

    def __init__(self, parent, title, fields, width=380):
        super().__init__(parent)
        self.title(title)
        self.configure(bg=BG)
        self.resizable(False, False)
        self.transient(parent)
        self.grab_set()
        self.result = None
        self.entries = []

        frm = ttk.Frame(self, padding=20)
        frm.pack(fill="both", expand=True)
        for i, (label, default, secret) in enumerate(fields):
            ttk.Label(frm, text=label).grid(row=i, column=0, sticky="w", pady=7, padx=(0, 12))
            show = "*" if secret else ""
            e = ttk.Entry(frm, show=show, width=26, font=FONT_NORMAL)
            e.insert(0, default or "")
            e.grid(row=i, column=1, pady=7)
            self.entries.append(e)
        btns = ttk.Frame(frm)
        btns.grid(row=len(fields), column=0, columnspan=2, pady=(18, 0))
        ttk.Button(btns, text="确定", style="Accent.TButton", command=self._ok).pack(side="left", padx=8)
        ttk.Button(btns, text="取消", command=self.destroy).pack(side="left", padx=8)
        self.update_idletasks()
        x = parent.winfo_rootx() + (parent.winfo_width() - self.winfo_width()) // 2
        y = parent.winfo_rooty() + (parent.winfo_height() - self.winfo_height()) // 3
        self.geometry(f"+{max(x,0)}+{max(y,0)}")
        self.entries[0].focus_set()
        self.bind("<Return>", lambda e: self._ok())
        self.bind("<Escape>", lambda e: self.destroy())
        self.wait_window()

    def _ok(self):
        self.result = [e.get().strip() for e in self.entries]
        self.destroy()
