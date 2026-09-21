package service

import (
	"testing"
	"time"

	"lms/backend/internal/models"
)

// 第一轮：纯业务规则单元测试（无 I/O 依赖）。

func TestCheckBorrow(t *testing.T) {
	cases := []struct {
		name              string
		available, active, overdue int
		wantErr           bool
	}{
		{"全部满足", 3, 2, 0, false},
		{"无库存", 0, 0, 0, true},
		{"达到上限", 3, 5, 0, true},
		{"存在逾期", 3, 1, 1, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := checkBorrow(c.available, c.active, c.overdue)
			if (err != nil) != c.wantErr {
				t.Fatalf("期望错误=%v，实际=%v", c.wantErr, err)
			}
		})
	}
}

func TestCheckRenew(t *testing.T) {
	today := time.Now().Format(dateLayout)
	future := time.Now().AddDate(0, 0, 10).Format(dateLayout)
	past := time.Now().AddDate(0, 0, -2).Format(dateLayout)

	if err := checkRenew(0, future, today); err != nil {
		t.Fatalf("正常续借应通过，实际=%v", err)
	}
	if err := checkRenew(1, future, today); err == nil {
		t.Fatal("续借次数超限应拒绝")
	}
	if err := checkRenew(0, past, today); err == nil {
		t.Fatal("逾期图书续借应拒绝")
	}
}

func TestCheckTotalQty(t *testing.T) {
	if err := checkTotalQty(5, 3); err != nil {
		t.Fatalf("馆藏>=已借应通过，实际=%v", err)
	}
	if err := checkTotalQty(2, 3); err == nil {
		t.Fatal("馆藏小于已借应拒绝")
	}
}

func TestCalcFine(t *testing.T) {
	today := time.Now()
	due := today.AddDate(0, 0, -10).Format(dateLayout)
	ret := today.Format(dateLayout)
	fine := calcFine(due, ret)
	if fine != 10*FinePerDay {
		t.Fatalf("逾期10天罚款期望 %.1f，实际 %.1f", 10*FinePerDay, fine)
	}
	// 未逾期
	due2 := today.AddDate(0, 0, 5).Format(dateLayout)
	if f := calcFine(due2, ret); f != 0 {
		t.Fatalf("未逾期罚款应为0，实际 %.1f", f)
	}
}

func TestStateText(t *testing.T) {
	now := time.Now()
	// 已归还
	r := modelBorrow("returned", now.AddDate(0, 0, 5).Format(dateLayout))
	if got := stateText(r, now); got != "已归还" {
		t.Fatalf("已归还状态错误：%s", got)
	}
	// 逾期
	r2 := modelBorrow("borrowed", now.AddDate(0, 0, -7).Format(dateLayout))
	if got := stateText(r2, now); got != "已逾期 7 天" {
		t.Fatalf("逾期状态错误：%s", got)
	}
	// 借阅中
	r3 := modelBorrow("borrowed", now.AddDate(0, 0, 12).Format(dateLayout))
	if got := stateText(r3, now); got != "借阅中（剩 12 天）" {
		t.Fatalf("借阅中状态错误：%s", got)
	}
}

func modelBorrow(status, due string) models.BorrowRecord {
	return models.BorrowRecord{Status: status, DueDate: due}
}
