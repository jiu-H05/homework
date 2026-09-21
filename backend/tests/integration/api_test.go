// Package integration 为第二轮测试：对完整 REST API 做黑盒集成测试。
package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"lms/backend/internal/httpapi"
	"lms/backend/internal/service"
	"lms/backend/internal/store"
)

// harness 封装测试服务器与请求辅助方法。
type harness struct {
	t    *testing.T
	srv  *httptest.Server
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "it.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("初始化失败: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	svc := service.New(st)
	h := &harness{t: t}
	h.srv = httptest.NewServer(httpapi.New(svc, []byte("test-secret")).Handler())
	t.Cleanup(h.srv.Close)
	return h
}

// do 发起请求；token 为空时不带鉴权。返回状态码与解析后的 JSON。
func (h *harness) do(method, path, token string, body any) (int, map[string]any) {
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, h.srv.URL+path, rdr)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.t.Fatalf("请求失败 %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return resp.StatusCode, out
}

// doList 发起请求并返回数组结果。
func (h *harness) doList(method, path, token string, body any) (int, []any) {
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, h.srv.URL+path, rdr)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		h.t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out []any
	_ = json.Unmarshal(raw, &out)
	return resp.StatusCode, out
}

func (h *harness) login(username, password string) string {
	code, body := h.do("POST", "/api/auth/login", "", map[string]string{
		"username": username, "password": password})
	if code != 200 {
		h.t.Fatalf("登录失败 code=%d body=%v", code, body)
	}
	return body["token"].(string)
}

func TestHealthAndAuth(t *testing.T) {
	h := newHarness(t)

	code, body := h.do("GET", "/api/health", "", nil)
	if code != 200 || body["status"] != "ok" {
		t.Fatal("健康检查失败")
	}
	// 错误口令
	if code, _ := h.do("POST", "/api/auth/login", "", map[string]string{
		"username": "admin", "password": "bad"}); code != 400 {
		t.Fatalf("错误口令应返回 400，实际 %d", code)
	}
	// 未带令牌访问受保护接口
	if code, _ := h.do("GET", "/api/books", "", nil); code != 401 {
		t.Fatalf("未登录应返回 401，实际 %d", code)
	}
	// 正常登录
	token := h.login("admin", "admin123")
	if token == "" {
		t.Fatal("应返回令牌")
	}
	code, body = h.do("GET", "/api/auth/me", token, nil)
	if code != 200 || body["role"] != "admin" {
		t.Fatalf("获取当前用户失败: %v", body)
	}
}

func TestRegisterAndSelfBorrow(t *testing.T) {
	h := newHarness(t)

	// 注册
	code, body := h.do("POST", "/api/auth/register", "", map[string]string{
		"username": "newreader", "password": "test123", "realName": "新读者", "phone": "13700000000"})
	if code != 201 {
		t.Fatalf("注册失败 code=%d body=%v", code, body)
	}
	token := h.login("newreader", "test123")

	// 图书检索
	code, books := h.doList("GET", "/api/books?keyword=Python", token, nil)
	if code != 200 || len(books) == 0 {
		t.Fatal("关键字检索应有结果")
	}
	code, all := h.doList("GET", "/api/books", token, nil)
	if len(all) < 10 {
		t.Fatal("初始图书应 >=10")
	}
	// 找可借图书
	avail := []int64{}
	for _, b := range all {
		m := b.(map[string]any)
		if int(m["availableQty"].(float64)) > 0 {
			avail = append(avail, int64(m["id"].(float64)))
		}
	}
	// 自助借满 5 本
	for i := 0; i < 5; i++ {
		code, _ = h.do("POST", "/api/me/borrows", token, map[string]int64{"bookId": avail[i]})
		if code != 201 {
			t.Fatalf("第 %d 本借阅失败 code=%d", i+1, code)
		}
	}
	_, active := h.doList("GET", "/api/me/borrows/active", token, nil)
	if len(active) != 5 {
		t.Fatalf("在借应 5 本，实际 %d", len(active))
	}
	// 第 6 本拒绝
	code, _ = h.do("POST", "/api/me/borrows", token, map[string]int64{"bookId": avail[5]})
	if code != 400 {
		t.Fatalf("超上限应 400，实际 %d", code)
	}
	// 续借一次成功、二次拒绝
	recID := int64(active[0].(map[string]any)["id"].(float64))
	if code, _ = h.do("POST", "/api/me/borrows/"+itoa(recID)+"/renew", token, nil); code != 200 {
		t.Fatalf("首次续借应成功 code=%d", code)
	}
	if code, _ = h.do("POST", "/api/me/borrows/"+itoa(recID)+"/renew", token, nil); code != 400 {
		t.Fatalf("二次续借应 400，实际 %d", code)
	}
}

func TestAdminBorrowReturnFine(t *testing.T) {
	h := newHarness(t)
	adminToken := h.login("admin", "admin123")

	// 全部在借（含种子逾期记录）
	_, active := h.doList("GET", "/api/borrows/active", adminToken, nil)
	var overdueID int64
	for _, r := range active {
		m := r.(map[string]any)
		if m["title"] == "时间简史" {
			overdueID = int64(m["id"].(float64))
		}
	}
	if overdueID == 0 {
		t.Fatal("应存在逾期种子记录")
	}
	code, body := h.do("POST", "/api/borrows/"+itoa(overdueID)+"/return", adminToken, nil)
	if code != 200 {
		t.Fatalf("归还失败 code=%d", code)
	}
	fine := body["fine"].(float64)
	if fine < 4.5 || fine > 5.5 {
		t.Fatalf("逾期罚款应约 5 元，实际 %.2f", fine)
	}
}

func TestAdminBookCRUD(t *testing.T) {
	h := newHarness(t)
	token := h.login("admin", "admin123")

	book := map[string]any{
		"isbn": "9780000000001", "title": "测试新书", "author": "测试作者",
		"publisher": "测试出版社", "categoryId": 1, "location": "T-01", "totalQty": 3,
	}
	code, created := h.do("POST", "/api/books", token, book)
	if code != 201 {
		t.Fatalf("新增图书失败 code=%d", code)
	}
	id := int64(created["id"].(float64))
	book["title"] = "测试新书-改"
	book["totalQty"] = 2
	if code, _ := h.do("PUT", "/api/books/"+itoa(id), token, book); code != 200 {
		t.Fatal("修改图书失败")
	}
	if code, _ := h.do("DELETE", "/api/books/"+itoa(id), token, nil); code != 200 {
		t.Fatal("删除无借阅图书应成功")
	}
	// 删除不存在的图书
	if code, _ := h.do("DELETE", "/api/books/99999", token, nil); code != 400 {
		t.Fatal("删除不存在图书应 400")
	}
}

func TestAdminReaderManagement(t *testing.T) {
	h := newHarness(t)
	token := h.login("admin", "admin123")

	code, body := h.do("POST", "/api/readers", token, map[string]string{
		"username": "adminmade", "password": "init123", "realName": "管理员建的读者", "phone": ""})
	if code != 201 {
		t.Fatalf("管理员新增读者失败 code=%d body=%v", code, body)
	}
	// 列表
	_, readers := h.doList("GET", "/api/readers", token, nil)
	var userID int64
	for _, r := range readers {
		m := r.(map[string]any)
		if m["username"] == "adminmade" {
			userID = int64(m["userId"].(float64))
		}
	}
	if userID == 0 {
		t.Fatal("新读者应出现在列表")
	}
	// 停用
	if code, _ := h.do("PATCH", "/api/readers/"+itoa(userID)+"/status", token,
		map[string]string{"status": "disabled"}); code != 200 {
		t.Fatal("停用失败")
	}
	// 重置口令
	code, body = h.do("POST", "/api/readers/"+itoa(userID)+"/reset-password", token, nil)
	if code != 200 || body["tempPassword"] == "" {
		t.Fatal("重置口令失败")
	}
}

func TestAdminPermission(t *testing.T) {
	h := newHarness(t)
	readerToken := h.login("reader", "reader123")

	// 读者访问管理员接口应 403
	if code, _ := h.do("GET", "/api/readers", readerToken, nil); code != 403 {
		t.Fatalf("读者访问读者管理应 403，实际 %d", code)
	}
	if code, _ := h.do("GET", "/api/stats/dashboard", readerToken, nil); code != 403 {
		t.Fatalf("读者访问统计应 403，实际 %d", code)
	}
	if code, _ := h.do("POST", "/api/books", readerToken, map[string]any{}); code != 403 {
		t.Fatalf("读者新增图书应 403，实际 %d", code)
	}
}

func TestStatsEndpoints(t *testing.T) {
	h := newHarness(t)
	token := h.login("admin", "admin123")

	code, d := h.do("GET", "/api/stats/dashboard", token, nil)
	if code != 200 || d["bookKinds"].(float64) < 10 {
		t.Fatal("控制台统计失败")
	}
	if code, cats := h.doList("GET", "/api/stats/categories", token, nil); code != 200 || len(cats) < 6 {
		t.Fatal("分类统计失败")
	}
	if code, hot := h.doList("GET", "/api/stats/hot-books", token, nil); code != 200 || len(hot) < 5 {
		t.Fatal("热门图书失败")
	}
	if code, logs := h.doList("GET", "/api/logs", token, nil); code != 200 || len(logs) == 0 {
		t.Fatal("操作日志失败")
	}
}

func itoa(n int64) string {
	return jsonNumber(n)
}

func jsonNumber(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}
