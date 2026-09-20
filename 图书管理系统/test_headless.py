# -*- coding: utf-8 -*-
"""
图书管理系统 - 业务逻辑无头测试（无需图形界面）。
使用临时数据库，覆盖：登录注册、借阅上限、续借规则、归还、逾期罚款、
逾期阻断借阅、改密、删除保护、统计报表等核心业务规则。
运行：python test_headless.py
"""
import os
import sys
import tempfile
from datetime import date, timedelta

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import library.database as dbmod
from library.database import init_db, get_conn
from library.services import LibraryService, ServiceError, MAX_BORROW, FINE_PER_DAY

PASS = 0
FAIL = 0


def check(name, cond):
    global PASS, FAIL
    if cond:
        PASS += 1
        print(f"  [PASS] {name}")
    else:
        FAIL += 1
        print(f"  [FAIL] {name}")


def expect_error(name, fn):
    global PASS, FAIL
    try:
        fn()
        FAIL += 1
        print(f"  [FAIL] {name}（未抛出预期异常）")
    except ServiceError:
        PASS += 1
        print(f"  [PASS] {name}")


def main():
    tmp = os.path.join(tempfile.gettempdir(), "lms_test.db")
    if os.path.exists(tmp):
        os.remove(tmp)
    dbmod.DB_PATH = tmp
    init_db()

    svc = LibraryService()

    print("== 1. 登录认证 ==")
    admin = svc.authenticate("admin", "admin123")
    check("管理员登录成功且角色为 admin", admin["role"] == "admin")
    expect_error("错误口令被拒绝", lambda: svc.authenticate("admin", "wrong"))
    expect_error("不存在用户被拒绝", lambda: svc.authenticate("nobody", "x"))

    print("== 2. 读者注册 ==")
    no = svc.register_reader("tuser", "test123", "测试读者", "13900000000")
    check("注册返回读者编号", bool(no))
    expect_error("重复用户名被拒绝", lambda: svc.register_reader("tuser", "test123", "x"))
    expect_error("口令过短被拒绝", lambda: svc.register_reader("u2", "123", "x"))
    reader = svc.authenticate("tuser", "test123")
    check("新读者登录成功且角色为 reader", reader["role"] == "reader")
    r = svc.get_reader_by_user(reader["id"])

    print("== 3. 借阅与在借上限 ==")
    books = svc.list_books()
    check("初始馆藏图书种类 >= 10", len(books) >= 10)
    available = [b for b in books if b["available_qty"] > 0]
    for i in range(MAX_BORROW):
        svc.borrow(r["id"], available[i]["id"])
    check(f"成功借阅 {MAX_BORROW} 本达到上限", len(svc.active_borrows(r["id"])) == MAX_BORROW)
    expect_error("第 6 本借阅被上限拒绝", lambda: svc.borrow(r["id"], available[MAX_BORROW]["id"]))
    expect_error("无库存图书借阅被拒绝", lambda: svc.borrow(r["id"], 999999))

    print("== 4. 续借规则 ==")
    recs = svc.active_borrows(r["id"])
    rec0 = recs[0]
    svc.renew(rec0["id"])
    rec0_after = next(x for x in svc.active_borrows(r["id"]) if x["id"] == rec0["id"])
    check("续借后次数为 1", rec0_after["renew_count"] == 1)
    expect_error("同一本二次续借被拒绝", lambda: svc.renew(rec0["id"]))

    print("== 5. 正常归还 ==")
    fine = svc.return_book(recs[1]["id"])
    check("未逾期图书归还罚款为 0", fine == 0)
    check("归还后在借数量减 1", len(svc.active_borrows(r["id"])) == MAX_BORROW - 1)
    expect_error("重复归还被拒绝", lambda: svc.return_book(recs[1]["id"]))

    print("== 6. 逾期、罚款与借阅阻断 ==")
    overdue_book = next(b for b in books if b["id"] not in [x["book_id"] for x in recs])
    overdue_days = 12
    b_date = date.today() - timedelta(days=overdue_days + 30)
    due = b_date + timedelta(days=30)
    conn = get_conn()
    conn.execute(
        "INSERT INTO borrow_records(reader_id,book_id,borrow_date,due_date,renew_count,status,fine) "
        "VALUES(?,?,?,?,?,?,?)",
        (r["id"], overdue_book["id"], b_date.isoformat(), due.isoformat(), 0, "borrowed", 0),
    )
    conn.execute("UPDATE books SET available_qty=available_qty-1 WHERE id=?", (overdue_book["id"],))
    conn.commit()
    conn.close()
    overdue_rec = [x for x in svc.active_borrows(r["id"]) if x["book_id"] == overdue_book["id"]][0]
    check("状态识别为逾期", "逾期" in svc.record_state(overdue_rec))
    expect_error("存在逾期图书时借阅被阻断", lambda: svc.borrow(r["id"], available[MAX_BORROW]["id"]))
    expect_error("逾期图书续借被拒绝", lambda: svc.renew(overdue_rec["id"]))
    fine = svc.return_book(overdue_rec["id"])
    check(f"逾期 {overdue_days} 天罚款 = {overdue_days * FINE_PER_DAY}",
          abs(fine - overdue_days * FINE_PER_DAY) < 0.001)

    print("== 7. 逾期处理后可继续借阅 ==")
    svc.borrow(r["id"], available[MAX_BORROW]["id"])
    check("逾期归还后借阅恢复正常", len(svc.active_borrows(r["id"])) == MAX_BORROW)

    print("== 8. 图书删除保护 ==")
    active_book_id = svc.active_borrows(r["id"])[0]["book_id"]
    expect_error("有在借记录的图书不可删除", lambda: svc.delete_book(active_book_id))

    print("== 9. 修改口令 ==")
    expect_error("原口令错误时改密被拒绝",
                 lambda: svc.change_password(reader["id"], "wrongold", "newpass1"))
    svc.change_password(reader["id"], "test123", "newpass1")
    check("新口令可登录", svc.authenticate("tuser", "newpass1")["role"] == "reader")

    print("== 10. 统计报表 ==")
    d = svc.dashboard()
    check("统计含全部关键指标", all(k in d for k in
          ["book_kinds", "book_copies", "readers", "borrowed", "overdue"]))
    check("读者统计包含种子与新注册", d["readers"] >= 2)
    check("分类统计非空", len(svc.category_stats()) >= 6)
    check("热门图书排行非空", len(svc.hot_books()) >= 5)
    check("操作日志有记录", len(svc.recent_logs()) > 0)

    print("== 11. 账号停用 ==")
    svc.set_user_status(reader["id"], "disabled")
    expect_error("被停用账号无法登录", lambda: svc.authenticate("tuser", "newpass1"))
    svc.set_user_status(reader["id"], "active")
    check("重新启用后可登录", svc.authenticate("tuser", "newpass1") is not None)

    os.remove(tmp)
    print(f"\n结果：{PASS} 项通过，{FAIL} 项失败。")
    sys.exit(1 if FAIL else 0)


if __name__ == "__main__":
    main()
