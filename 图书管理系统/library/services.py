# -*- coding: utf-8 -*-
"""
业务逻辑层：图书、读者、借阅/归还/续借、逾期、统计等全部业务规则。
UI 只调用本层方法，保证业务规则集中、可测试。
"""
from datetime import datetime, date, timedelta

from .database import get_conn
from .auth import hash_password, verify_password

# 业务常量
LOAN_DAYS = 30          # 借期（天）
MAX_BORROW = 5          # 每位读者最大在借数量
MAX_RENEW = 1           # 最大续借次数
RENEW_DAYS = 30         # 每次续借延长天数
FINE_PER_DAY = 0.5      # 逾期罚款（元/天）


class ServiceError(Exception):
    """业务规则校验失败时抛出，UI 捕获后提示。"""


class LibraryService:
    def __init__(self):
        self.current_user = None  # 登录后保存用户行

    # ---------- 基础 ----------
    @staticmethod
    def _now():
        return datetime.now().strftime("%Y-%m-%d %H:%M:%S")

    def log(self, action: str, detail: str = ""):
        conn = get_conn()
        try:
            uid = self.current_user["id"] if self.current_user else None
            uname = self.current_user["username"] if self.current_user else "system"
            conn.execute(
                "INSERT INTO operation_logs(user_id,username,action,detail,created_at) VALUES(?,?,?,?,?)",
                (uid, uname, action, detail, self._now()),
            )
            conn.commit()
        finally:
            conn.close()

    # ---------- 登录注册 ----------
    def authenticate(self, username: str, password: str):
        conn = get_conn()
        try:
            row = conn.execute("SELECT * FROM users WHERE username=?", (username,)).fetchone()
            if row is None:
                raise ServiceError("用户名不存在")
            if row["status"] != "active":
                raise ServiceError("该账号已被停用，请联系管理员")
            if not verify_password(password, row["password_hash"]):
                raise ServiceError("口令错误")
            self.current_user = row
            self.log("登录", f"用户 {username} 登录系统")
            return row
        finally:
            conn.close()

    def register_reader(self, username: str, password: str, real_name: str, phone: str = ""):
        username = username.strip()
        real_name = real_name.strip()
        if not username or not password or not real_name:
            raise ServiceError("用户名、口令、姓名不能为空")
        if len(password) < 6:
            raise ServiceError("口令长度至少 6 位")
        conn = get_conn()
        try:
            exists = conn.execute("SELECT id FROM users WHERE username=?", (username,)).fetchone()
            if exists:
                raise ServiceError("用户名已存在")
            cur = conn.execute(
                "INSERT INTO users(username,password_hash,role,real_name,status,created_at) VALUES(?,?,?,?,?,?)",
                (username, hash_password(password), "reader", real_name, "active", self._now()),
            )
            # 读者编号：R + 年份 + 3位序号
            n = conn.execute("SELECT COUNT(*) AS n FROM readers").fetchone()["n"] + 1
            reader_no = f"R{datetime.now().year}{n:03d}"
            conn.execute(
                "INSERT INTO readers(user_id,reader_no,name,phone,created_at) VALUES(?,?,?,?,?)",
                (cur.lastrowid, reader_no, real_name, phone.strip(), self._now()),
            )
            conn.commit()
            return reader_no
        finally:
            conn.close()

    def change_password(self, user_id: int, old_pwd: str, new_pwd: str):
        if len(new_pwd) < 6:
            raise ServiceError("新口令长度至少 6 位")
        conn = get_conn()
        try:
            row = conn.execute("SELECT * FROM users WHERE id=?", (user_id,)).fetchone()
            if not verify_password(old_pwd, row["password_hash"]):
                raise ServiceError("原口令错误")
            conn.execute("UPDATE users SET password_hash=? WHERE id=?", (hash_password(new_pwd), user_id))
            conn.commit()
        finally:
            conn.close()

    # ---------- 读者管理 ----------
    def list_readers(self, keyword: str = ""):
        conn = get_conn()
        try:
            sql = (
                "SELECT r.id, r.reader_no, r.name, r.phone, u.username, u.status, u.id AS user_id, r.created_at "
                "FROM readers r JOIN users u ON r.user_id=u.id "
            )
            if keyword:
                sql += "WHERE r.name LIKE ? OR r.reader_no LIKE ? OR u.username LIKE ? "
                kw = f"%{keyword}%"
                return conn.execute(sql, (kw, kw, kw)).fetchall()
            return conn.execute(sql + "ORDER BY r.id").fetchall()
        finally:
            conn.close()

    def get_reader_by_user(self, user_id: int):
        conn = get_conn()
        try:
            return conn.execute("SELECT * FROM readers WHERE user_id=?", (user_id,)).fetchone()
        finally:
            conn.close()

    def set_user_status(self, user_id: int, status: str):
        conn = get_conn()
        try:
            conn.execute("UPDATE users SET status=? WHERE id=?", (status, user_id))
            conn.commit()
        finally:
            conn.close()

    def reset_password(self, user_id: int, new_pwd: str = "123456"):
        conn = get_conn()
        try:
            conn.execute("UPDATE users SET password_hash=? WHERE id=?", (hash_password(new_pwd), user_id))
            conn.commit()
        finally:
            conn.close()

    # ---------- 分类 ----------
    def list_categories(self):
        conn = get_conn()
        try:
            return conn.execute("SELECT * FROM categories ORDER BY id").fetchall()
        finally:
            conn.close()

    def add_category(self, name: str):
        name = name.strip()
        if not name:
            raise ServiceError("分类名称不能为空")
        conn = get_conn()
        try:
            exists = conn.execute("SELECT id FROM categories WHERE name=?", (name,)).fetchone()
            if exists:
                raise ServiceError("分类已存在")
            conn.execute("INSERT INTO categories(name) VALUES(?)", (name,))
            conn.commit()
        finally:
            conn.close()

    # ---------- 图书管理 ----------
    def list_books(self, keyword: str = "", category_id=None):
        conn = get_conn()
        try:
            sql = (
                "SELECT b.*, c.name AS category FROM books b "
                "LEFT JOIN categories c ON b.category_id=c.id WHERE 1=1 "
            )
            params = []
            if category_id:
                sql += "AND b.category_id=? "
                params.append(category_id)
            if keyword:
                sql += "AND (b.title LIKE ? OR b.author LIKE ? OR b.isbn LIKE ?) "
                kw = f"%{keyword}%"
                params.extend([kw, kw, kw])
            sql += "ORDER BY b.id"
            return conn.execute(sql, params).fetchall()
        finally:
            conn.close()

    def add_book(self, isbn, title, author, publisher, category_id, location, total_qty):
        title = title.strip()
        if not title:
            raise ServiceError("书名不能为空")
        total_qty = int(total_qty)
        if total_qty < 1:
            raise ServiceError("馆藏数量必须大于 0")
        conn = get_conn()
        try:
            conn.execute(
                "INSERT INTO books(isbn,title,author,publisher,category_id,location,total_qty,available_qty,created_at) "
                "VALUES(?,?,?,?,?,?,?,?,?)",
                (isbn.strip(), title, author.strip(), publisher.strip(), category_id, location.strip(),
                 total_qty, total_qty, self._now()),
            )
            conn.commit()
        finally:
            conn.close()

    def update_book(self, book_id, isbn, title, author, publisher, category_id, location, total_qty):
        total_qty = int(total_qty)
        conn = get_conn()
        try:
            row = conn.execute("SELECT * FROM books WHERE id=?", (book_id,)).fetchone()
            borrowed = row["total_qty"] - row["available_qty"]
            if total_qty < borrowed:
                raise ServiceError(f"馆藏数量不能小于当前已借出数量（{borrowed} 本）")
            available = total_qty - borrowed
            conn.execute(
                "UPDATE books SET isbn=?,title=?,author=?,publisher=?,category_id=?,location=?,"
                "total_qty=?,available_qty=? WHERE id=?",
                (isbn.strip(), title.strip(), author.strip(), publisher.strip(), category_id,
                 location.strip(), total_qty, available, book_id),
            )
            conn.commit()
        finally:
            conn.close()

    def delete_book(self, book_id):
        conn = get_conn()
        try:
            active = conn.execute(
                "SELECT COUNT(*) AS n FROM borrow_records WHERE book_id=? AND status='borrowed'", (book_id,)
            ).fetchone()
            if active["n"] > 0:
                raise ServiceError("该图书存在未归还的借阅记录，无法删除")
            conn.execute("DELETE FROM books WHERE id=?", (book_id,))
            conn.commit()
        finally:
            conn.close()

    # ---------- 借阅 ----------
    def _reader_active_count(self, conn, reader_id):
        return conn.execute(
            "SELECT COUNT(*) AS n FROM borrow_records WHERE reader_id=? AND status='borrowed'", (reader_id,)
        ).fetchone()["n"]

    def _reader_has_overdue(self, conn, reader_id):
        return conn.execute(
            "SELECT COUNT(*) AS n FROM borrow_records WHERE reader_id=? AND status='borrowed' AND due_date<?",
            (reader_id, date.today().isoformat()),
        ).fetchone()["n"]

    def borrow(self, reader_id, book_id):
        conn = get_conn()
        try:
            book = conn.execute("SELECT * FROM books WHERE id=?", (book_id,)).fetchone()
            if book is None:
                raise ServiceError("图书不存在")
            if book["available_qty"] <= 0:
                raise ServiceError("该图书暂无在馆库存")
            if self._reader_active_count(conn, reader_id) >= MAX_BORROW:
                raise ServiceError(f"在借图书已达上限（{MAX_BORROW} 本），请先归还")
            if self._reader_has_overdue(conn, reader_id) > 0:
                raise ServiceError("您有逾期未还的图书，请先归还后再借阅")
            today_ = date.today()
            conn.execute(
                "INSERT INTO borrow_records(reader_id,book_id,borrow_date,due_date,renew_count,status,fine) "
                "VALUES(?,?,?,?,?,?,?)",
                (reader_id, book_id, today_.isoformat(), (today_ + timedelta(days=LOAN_DAYS)).isoformat(),
                 0, "borrowed", 0),
            )
            conn.execute("UPDATE books SET available_qty=available_qty-1 WHERE id=?", (book_id,))
            conn.commit()
        finally:
            conn.close()
        self.log("借阅", f"读者ID {reader_id} 借阅图书ID {book_id}")

    # ---------- 续借 ----------
    def renew(self, record_id):
        conn = get_conn()
        try:
            rec = conn.execute("SELECT * FROM borrow_records WHERE id=?", (record_id,)).fetchone()
            if rec is None or rec["status"] != "borrowed":
                raise ServiceError("借阅记录不存在或已归还")
            if rec["renew_count"] >= MAX_RENEW:
                raise ServiceError(f"该图书已续借 {MAX_RENEW} 次，无法再次续借")
            if rec["due_date"] < date.today().isoformat():
                raise ServiceError("图书已逾期，无法续借，请先归还")
            new_due = date.fromisoformat(rec["due_date"]) + timedelta(days=RENEW_DAYS)
            conn.execute(
                "UPDATE borrow_records SET renew_count=renew_count+1, due_date=? WHERE id=?",
                (new_due.isoformat(), record_id),
            )
            conn.commit()
        finally:
            conn.close()
        self.log("续借", f"借阅记录 {record_id} 续借成功")

    # ---------- 归还 ----------
    def return_book(self, record_id):
        conn = get_conn()
        try:
            rec = conn.execute("SELECT * FROM borrow_records WHERE id=?", (record_id,)).fetchone()
            if rec is None or rec["status"] != "borrowed":
                raise ServiceError("借阅记录不存在或已归还")
            today_ = date.today()
            due = date.fromisoformat(rec["due_date"])
            fine = 0.0
            if today_ > due:
                fine = (today_ - due).days * FINE_PER_DAY
            conn.execute(
                "UPDATE borrow_records SET status='returned', return_date=?, fine=? WHERE id=?",
                (today_.isoformat(), fine, record_id),
            )
            conn.execute("UPDATE books SET available_qty=available_qty+1 WHERE id=?", (rec["book_id"],))
            conn.commit()
            return fine
        finally:
            conn.close()

    # ---------- 查询 ----------
    def active_borrows(self, reader_id=None):
        conn = get_conn()
        try:
            sql = (
                "SELECT br.*, b.title, b.author, b.isbn, r.name AS reader_name, r.reader_no "
                "FROM borrow_records br JOIN books b ON br.book_id=b.id "
                "JOIN readers r ON br.reader_id=r.id WHERE br.status='borrowed' "
            )
            if reader_id:
                sql += "AND br.reader_id=? ORDER BY br.due_date"
                return conn.execute(sql, (reader_id,)).fetchall()
            return conn.execute(sql + "ORDER BY br.due_date").fetchall()
        finally:
            conn.close()

    def history(self, reader_id=None):
        conn = get_conn()
        try:
            sql = (
                "SELECT br.*, b.title, b.author, r.name AS reader_name, r.reader_no "
                "FROM borrow_records br JOIN books b ON br.book_id=b.id "
                "JOIN readers r ON br.reader_id=r.id WHERE br.status='returned' "
            )
            if reader_id:
                sql += "AND br.reader_id=? ORDER BY br.return_date DESC"
                return conn.execute(sql, (reader_id,)).fetchall()
            return conn.execute(sql + "ORDER BY br.return_date DESC").fetchall()
        finally:
            conn.close()

    @staticmethod
    def record_state(rec):
        """根据应还日期动态判定借阅状态文本。"""
        if rec["status"] == "returned":
            return "已归还"
        if rec["due_date"] < date.today().isoformat():
            days = (date.today() - date.fromisoformat(rec["due_date"])).days
            return f"已逾期 {days} 天"
        days = (date.fromisoformat(rec["due_date"]) - date.today()).days
        return f"借阅中（剩 {days} 天）"

    # ---------- 统计 ----------
    def dashboard(self):
        conn = get_conn()
        try:
            today_ = date.today().isoformat()
            stats = {
                "book_kinds": conn.execute("SELECT COUNT(*) AS n FROM books").fetchone()["n"],
                "book_copies": conn.execute("SELECT COALESCE(SUM(total_qty),0) AS n FROM books").fetchone()["n"],
                "readers": conn.execute("SELECT COUNT(*) AS n FROM readers").fetchone()["n"],
                "borrowed": conn.execute(
                    "SELECT COUNT(*) AS n FROM borrow_records WHERE status='borrowed'"
                ).fetchone()["n"],
                "overdue": conn.execute(
                    "SELECT COUNT(*) AS n FROM borrow_records WHERE status='borrowed' AND due_date<?", (today_,)
                ).fetchone()["n"],
            }
            return stats
        finally:
            conn.close()

    def category_stats(self):
        conn = get_conn()
        try:
            return conn.execute(
                "SELECT c.name, COUNT(b.id) AS kinds, COALESCE(SUM(b.total_qty),0) AS copies "
                "FROM categories c LEFT JOIN books b ON b.category_id=c.id GROUP BY c.id ORDER BY copies DESC"
            ).fetchall()
        finally:
            conn.close()

    def hot_books(self, limit=8):
        conn = get_conn()
        try:
            return conn.execute(
                "SELECT b.title, b.author, COUNT(br.id) AS times FROM books b "
                "LEFT JOIN borrow_records br ON br.book_id=b.id GROUP BY b.id ORDER BY times DESC, b.id LIMIT ?",
                (limit,),
            ).fetchall()
        finally:
            conn.close()

    def recent_logs(self, limit=50):
        conn = get_conn()
        try:
            return conn.execute(
                "SELECT * FROM operation_logs ORDER BY id DESC LIMIT ?", (limit,)
            ).fetchall()
        finally:
            conn.close()
