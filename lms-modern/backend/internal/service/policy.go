package service

import (
	"strconv"
	"time"

	"lms/backend/internal/models"
)

// 本文件集中放置纯业务规则判定函数，无 I/O 依赖，便于单元测试与复用（高内聚）。

const dateLayout = "2006-01-02"

// checkBorrow 判定是否允许借阅：库存、在借上限、逾期三个条件。
func checkBorrow(available, activeCount, overdueCount int) error {
	switch {
	case available <= 0:
		return Error("该图书暂无在馆库存")
	case activeCount >= MaxBorrow:
		return Error("在借图书已达上限（5 本），请先归还")
	case overdueCount > 0:
		return Error("您有逾期未还的图书，请先归还后再借阅")
	}
	return nil
}

// checkRenew 判定是否允许续借：续借次数与是否逾期。
func checkRenew(renewCount int, dueDate, today string) error {
	switch {
	case renewCount >= MaxRenew:
		return Error("该图书已续借 1 次，无法再次续借")
	case dueDate < today:
		return Error("图书已逾期，无法续借，请先归还")
	}
	return nil
}

// checkTotalQty 判定修改馆藏数量时不能小于已借出数量。
func checkTotalQty(totalQty, borrowed int) error {
	if totalQty < borrowed {
		return Error("馆藏数量不能小于当前已借出数量")
	}
	return nil
}

// calcFine 按应还日期与实际归还日期计算逾期罚款；未逾期为 0。
func calcFine(dueDate, returnDate string) float64 {
	due, err1 := time.Parse(dateLayout)
	ret, err2 := time.Parse(dateLayout, returnDate)
	if err1 != nil || err2 != nil {
		return 0
	}
	if !ret.After(due) {
		return 0
	}
	return float64(ret.Sub(due).Hours()/24) * FinePerDay
}

// stateText 根据状态与应还日期动态生成借阅状态文本。
func stateText(rec models.BorrowRecord, today time.Time) string {
	if rec.Status == "returned" {
		return "已归还"
	}
	due, err := time.Parse(dateLayout, rec.DueDate)
	if err != nil {
		return rec.Status
	}
	nowDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)
	dueDay := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, time.Local)
	days := int(dueDay.Sub(nowDay).Hours() / 24)
	if days < 0 {
		return "已逾期 " + strconv.Itoa(-days) + " 天"
	}
	return "借阅中（剩 " + strconv.Itoa(days) + " 天）"
}

// annotate 为借阅记录批量填充动态状态字段。
func annotate(recs []models.BorrowRecord) {
	today := time.Now()
	for i := range recs {
		recs[i].State = stateText(recs[i], today)
	}
	return recs
}
