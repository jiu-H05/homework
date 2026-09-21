package service_test

import (
	"os"
	"path/filepath"
	"testing"

	"lms/backend/internal/models"
	"lms/backend/internal/service"
	"lms/backend/internal/store"
)

// newSvc 创建基于临时数据库的服务（测试结束自动清理）。
func newSvc(t *testing.T) (*service.Service, func()) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("初始化测试库失败: %v", err)
	}
	return service.New(st), func() { _ = st.Close(); _ = os.Remove(dbPath) }
}

func admin() *models.User { return &models.User{ID: 1, Username: "admin", Role: "admin", RealName: "系统管理员"} }

func TestLoginFlow(t *testing.T) {
	svc, cleanup := newSvc(t)
	defer cleanup()

	if _, err := svc.Login("admin", "admin123"); err != nil {
		t.Fatalf("管理员登录失败: %v", err)
	}
	if _, err := svc.Login("admin", "wrong"); err == nil {
		t.Fatal("错误口令应登录失败")
	}
	if _, err := svc.Login("nobody", "x"); err == nil {
		t.Fatal("不存在用户应登录失败")
	}
}

func TestRegisterFlow(t *testing.T) {
	svc, cleanup := newSvc(t)
	defer cleanup()

	no, err := svc.Register("tuser", "test123", "测试读者", "13900000000")
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}
	if no == "" {
		t.Fatal("应返回读者编号")
	}
	if _, err := svc.Register("tuser", "test123", "重复", ""); err == nil {
		t.Fatal("重复用户名应失败")
	}
	if _, err := svc.Register("u2", "123", "短口令", ""); err == nil {
		t.Fatal("短口令应失败")
	}
	if _, err := svc.Login("tuser", "test123"); err != nil {
		t.Fatalf("新读者登录失败: %v", err)
	}
}

func TestBorrowLimitFlow(t *testing.T) {
	svc, cleanup := newSvc(t)
	defer cleanup()

	no, _ := svc.Register("limit", "test123", "限借读者", "")
	_ = no
	// 找到该读者
	u, _ := svc.Login("limit", "test123")
	rd, err := svc.MyReader(u)
	if err != nil {
		t.Fatal(err)
	}
	books, _ := svc.ListBooks("", nil, "")
	avail := []models.Book{}
	for _, b := range books {
		if b.AvailableQty > 0 {
			avail = append(avail, b)
		}
	}
	for i := 0; i < service.MaxBorrow; i++ {
		if err := svc.Borrow(admin(), rd.ID, avail[i].ID); err != nil {
			t.Fatalf("第 %d 本借阅失败: %v", i+1, err)
		}
	}
	active, _ := svc.ActiveBorrows(&rd.ID)
	if len(active) != service.MaxBorrow {
		t.Fatalf("在借数量期望 %d，实际 %d", service.MaxBorrow, len(active))
	}
	if err := svc.Borrow(admin(), rd.ID, avail[service.MaxBorrow].ID); err == nil {
		t.Fatal("超过上限应拒绝借阅")
	}
}

func TestRenewFlow(t *testing.T) {
	svc, cleanup := newSvc(t)
	defer cleanup()

	svc.Register("renewu", "test123", "续借读者", "")
	u, _ := svc.Login("renewu", "test123")
	rd, _ := svc.MyReader(u)
	books, _ := svc.ListBooks("", nil, "")
	var bookID int64
	for _, b := range books {
		if b.AvailableQty > 0 {
			bookID = b.ID
			break
		}
	}
	if err := svc.Borrow(admin(), rd.ID, bookID); err != nil {
		t.Fatal(err)
	}
	active, _ := svc.ActiveBorrows(&rd.ID)
	recID := active[0].ID
	if err := svc.SelfRenew(u, recID); err != nil {
		t.Fatalf("首次续借失败: %v", err)
	}
	active, _ = svc.ActiveBorrows(&rd.ID)
	if active[0].RenewCount != 1 {
		t.Fatal("续借次数应为 1")
	}
	if err := svc.SelfRenew(u, recID); err == nil {
		t.Fatal("二次续借应拒绝")
	}
}

func TestReturnAndFineFlow(t *testing.T) {
	svc, cleanup := newSvc(t)
	defer cleanup()

	// 种子读者 reader（含一条逾期借阅：时间简史，40天前借出）
	recs, _ := svc.ActiveBorrows(nil)
	var overdue, normal int64
	for _, r := range recs {
		if r.Title == "时间简史" {
			overdue = r.ID
		}
		if r.Title == "Python编程：从入门到实践" {
			normal = r.ID
		}
	}
	fine, err := svc.Return(admin(), overdue)
	if err != nil {
		t.Fatalf("归还逾期图书失败: %v", err)
	}
	if fine < 4.5 || fine > 5.5 {
		t.Fatalf("逾期约10天罚款应约 5 元，实际 %.2f", fine)
	}
	fine2, err := svc.Return(admin(), normal)
	if err != nil {
		t.Fatalf("归还正常图书失败: %v", err)
	}
	if fine2 != 0 {
		t.Fatalf("未逾期罚款应为 0，实际 %.2f", fine2)
	}
	if _, err := svc.Return(admin(), overdue); err == nil {
		t.Fatal("重复归还应拒绝")
	}
}

func TestDeleteProtectionFlow(t *testing.T) {
	svc, cleanup := newSvc(t)
	defer cleanup()

	recs, _ := svc.ActiveBorrows(nil)
	if len(recs) == 0 {
		t.Fatal("种子数据应含在借记录")
	}
	if err := svc.DeleteBook(admin(), recs[0].BookID); err == nil {
		t.Fatal("有在借记录的图书应禁止删除")
	}
}

func TestPasswordFlow(t *testing.T) {
	svc, cleanup := newSvc(t)
	defer cleanup()

	u, _ := svc.Login("reader", "reader123")
	if err := svc.ChangePassword(u.ID, "wrong", "newpass1"); err == nil {
		t.Fatal("原口令错误应失败")
	}
	if err := svc.ChangePassword(u.ID, "reader123", "newpass1"); err != nil {
		t.Fatalf("改密失败: %v", err)
	}
	if _, err := svc.Login("reader", "newpass1"); err != nil {
		t.Fatal("新口令应可登录")
	}
}

func TestStatsFlow(t *testing.T) {
	svc, cleanup := newSvc(t)
	defer cleanup()

	d, err := svc.Dashboard()
	if err != nil {
		t.Fatal(err)
	}
	if d.BookKinds < 10 {
		t.Fatalf("图书种类应 >=10，实际 %d", d.BookKinds)
	}
	if d.Readers < 1 {
		t.Fatal("应至少有1名种子读者")
	}
	if d.Overdue < 1 {
		t.Fatal("种子数据应含1本逾期")
	}
	cats, _ := svc.CategoryStats()
	if len(cats) < 6 {
		t.Fatal("分类统计应 >=6")
	}
	hot, _ := svc.HotBooks()
	if len(hot) < 5 {
		t.Fatal("热门图书应 >=5")
	}
	if _, err := svc.Login("admin", "admin123"); err != nil {
		t.Fatal(err)
	}
	logs, _ := svc.RecentLogs()
	if len(logs) == 0 {
		t.Fatal("操作日志应非空")
	}
}

func TestDisableFlow(t *testing.T) {
	svc, cleanup := newSvc(t)
	defer cleanup()

	u, _ := svc.Login("reader", "reader123")
	if err := svc.SetReaderStatus(admin(), u.ID, "disabled"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Login("reader", "reader123"); err == nil {
		t.Fatal("停用账号应无法登录")
	}
	if err := svc.SetReaderStatus(admin(), u.ID, "active"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Login("reader", "reader123"); err != nil {
		t.Fatal("重新启用后应可登录")
	}
}
