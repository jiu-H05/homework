# -*- coding: utf-8 -*-
"""
数据库模块：SQLite 连接、建表与初始数据。
数据库文件 library_data.db 位于程序（.exe 或 main.py）所在目录，便于便携使用。
"""
import os
import sys
import sqlite3
from datetime import datetime, timedelta


def app_dir():
    """返回程序所在目录；打包后为 .exe 目录，开发时为项目根目录。"""
    if getattr(sys, "frozen", False):
        return os.path.dirname(sys.executable)
    return os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))


DB_PATH = os.path.join(app_dir(), "library_data.db")

SCHEMA = """
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('admin','reader')),
    real_name TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','disabled')),
    created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS readers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reader_no TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    phone TEXT,
    created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    isbn TEXT,
    title TEXT NOT NULL,
    author TEXT,
    publisher TEXT,
    category_id INTEGER REFERENCES categories(id),
    location TEXT,
    total_qty INTEGER NOT NULL DEFAULT 1,
    available_qty INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS borrow_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    reader_id INTEGER NOT NULL REFERENCES readers(id),
    book_id INTEGER NOT NULL REFERENCES books(id),
    borrow_date TEXT NOT NULL,
    due_date TEXT NOT NULL,
    return_date TEXT,
    renew_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'borrowed' CHECK(status IN ('borrowed','returned')),
    fine REAL NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS operation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    username TEXT,
    action TEXT NOT NULL,
    detail TEXT,
    created_at TEXT NOT NULL
);
"""


def get_conn():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    conn.execute("PRAGMA foreign_keys = ON")
    return conn


def _seed(conn):
    """写入初始账号、分类、图书与示例借阅记录。"""
    from .auth import hash_password

    now = datetime.now()
    today = now.date()

    # 默认管理员
    conn.execute(
        "INSERT OR IGNORE INTO users(username,password_hash,role,real_name,status,created_at) "
        "VALUES(?,?,?,?,?,?)",
        ("admin", hash_password("admin123"), "admin", "系统管理员", "active", now.strftime("%Y-%m-%d %H:%M:%S")),
    )

    # 示例读者
    conn.execute(
        "INSERT OR IGNORE INTO users(username,password_hash,role,real_name,status,created_at) "
        "VALUES(?,?,?,?,?,?)",
        ("reader", hash_password("reader123"), "reader", "张同学", "active", now.strftime("%Y-%m-%d %H:%M:%S")),
    )
    user_row = conn.execute("SELECT id FROM users WHERE username='reader'").fetchone()
    conn.execute(
        "INSERT OR IGNORE INTO readers(user_id,reader_no,name,phone,created_at) VALUES(?,?,?,?,?)",
        (user_row["id"], "R2023001", "张同学", "13800000000", now.strftime("%Y-%m-%d %H:%M:%S")),
    )

    # 分类
    cats = ["计算机", "文学", "历史", "自然科学", "艺术"]
    for c in cats:
        conn.execute("INSERT OR IGNORE INTO categories(name) VALUES(?)", (c,))

    # 图书
    reader_row = conn.execute("SELECT id FROM readers WHERE reader_no='R2023001'").fetchone()
    books = [
        ("9787111213826", "软件工程：实践者的研究方法", "Roger S. Pressman", "机械工业出版社", "计算机", "N502-A01", 5),
        ("9787115428028", "Python编程：从入门到实践", "Eric Matthes", "人民邮电出版社", "计算机", "N502-A02", 4),
        ("9787302259862", "算法导论", "Thomas H. Cormen", "清华大学出版社", "计算机", "N502-A03", 3),
        ("9787020008735", "红楼梦", "曹雪芹", "人民文学出版社", "文学", "N503-B01", 6),
        ("9787020008728", "西游记", "吴承恩", "人民文学出版社", "文学", "N503-B02", 6),
        ("9787101003048", "史记", "司马迁", "中华书局", "历史", "N504-C01", 3),
        ("9787535732224", "时间简史", "Stephen Hawking", "湖南科学技术出版社", "自然科学", "N505-D01", 4),
        ("9787544270878", "万物简史", "Bill Bryson", "接力出版社", "自然科学", "N505-D02", 3),
        ("9787102048161", "中国艺术史", "苏立文", "上海人民出版社", "艺术", "N506-E01", 2),
        ("9787544720885", "西方哲学史", "Bertrand Russell", "译林出版社", "哲学", "N505-D03", 3),
    ]
    # 哲学分类不在初始列表中，补充
    conn.execute("INSERT OR IGNORE INTO categories(name) VALUES(?)", ("哲学",))

    for isbn, title, author, publisher, cat, location, qty in books:
        exists = conn.execute("SELECT id FROM books WHERE isbn=?", (isbn,)).fetchone()
        if exists:
            continue
        crow = conn.execute("SELECT id FROM categories WHERE name=?", (cat,)).fetchone()
        conn.execute(
            "INSERT INTO books(isbn,title,author,publisher,category_id,location,total_qty,available_qty,created_at) "
            "VALUES(?,?,?,?,?,?,?,?,?)",
            (isbn, title, author, publisher, crow["id"], location, qty, qty, now.strftime("%Y-%m-%d %H:%M:%S")),
        )

    # 示例借阅记录：一本正常借阅，一本已逾期（用于演示逾期与罚款）
    book1 = conn.execute("SELECT id FROM books WHERE title='Python编程：从入门到实践'").fetchone()
    book2 = conn.execute("SELECT id FROM books WHERE title='时间简史'").fetchone()
    has_rec = conn.execute("SELECT COUNT(*) AS n FROM borrow_records").fetchone()
    if has_rec["n"] == 0 and reader_row:
        # 正常借阅：5天前借出，应还 25 天后
        b_date = today - timedelta(days=5)
        due = b_date + timedelta(days=30)
        conn.execute(
            "INSERT INTO borrow_records(reader_id,book_id,borrow_date,due_date,renew_count,status,fine) "
            "VALUES(?,?,?,?,?,?,?)",
            (reader_row["id"], book1["id"], b_date.isoformat(), due.isoformat(), 0, "borrowed", 0),
        )
        conn.execute("UPDATE books SET available_qty=available_qty-1 WHERE id=?", (book1["id"],))
        # 逾期借阅：40天前借出，应还 10 天前
        b_date2 = today - timedelta(days=40)
        due2 = b_date2 + timedelta(days=30)
        conn.execute(
            "INSERT INTO borrow_records(reader_id,book_id,borrow_date,due_date,renew_count,status,fine) "
            "VALUES(?,?,?,?,?,?,?)",
            (reader_row["id"], book2["id"], b_date2.isoformat(), due2.isoformat(), 0, "borrowed", 0),
        )
        conn.execute("UPDATE books SET available_qty=available_qty-1 WHERE id=?", (book2["id"],))


def init_db():
    """建表并在首次运行时写入种子数据。"""
    conn = get_conn()
    try:
        conn.executescript(SCHEMA)
        conn.commit()
        empty = conn.execute("SELECT COUNT(*) AS n FROM users").fetchone()
        if empty["n"] == 0:
            _seed(conn)
        conn.commit()
    finally:
        conn.close()
