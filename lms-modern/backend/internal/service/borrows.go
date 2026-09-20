package service

import (
	"errors"
	"strconv"
	"time"

	"lms/backend/internal/models"
)

// Borrow 管理员为读者办理借阅。
func (s *Service) Borrow(actor *models.User, readerID, bookID int64) error {
	reader, err := s.repo.GetReaderByID(readerID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return Error("读者不存在")
		}
		return err
	}
	book, err := s.repo.GetBook(bookID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return Error("图书不存在")
		}
		return err
	}
	today := time.Now()
	todayStr := today.Format(dateLayout)
	active, err := s.repo.CountActive(readerID)
	if err != nil {
		return err
	}
	overdue, err := s.repo.CountOverdue(readerID, todayStr)
	if err != nil {
		return err
	}
	if err := checkBorrow(book.AvailableQty, active, overdue); err != nil {
		return err
	}
	rec := &models.BorrowRecord{
		ReaderID:   readerID,
		BookID:     bookID,
		BorrowDate: todayStr,
		DueDate:    today.AddDate(0, 0, LoanDays).Format(dateLayout),
		Status:     "borrowed",
	}
	if err := s.repo.CreateBorrow(rec); err != nil {
		return Error("办理借阅失败")
	}
	if err := s.repo.DecAvailable(bookID); err != nil {
		return err
	}
	return s.log(actor.ID, actor.Username, "借阅图书",
		reader.Name+" 借阅《"+book.Title+"》")
}

// SelfBorrow 读者自助借阅。
func (s *Service) SelfBorrow(actor *models.User, bookID int64) error {
	reader, err := s.repo.GetReaderByUserID(actor.ID)
	if err != nil {
		return Error("读者档案不存在")
	}
	return s.Borrow(actor, reader.ID, bookID)
}

// Renew 办理续借；ownerReaderID 非 nil 时校验记录归属（读者自助场景）。
func (s *Service) Renew(actor *models.User, recordID int64, ownerReaderID *int64) error {
	rec, err := s.repo.GetBorrow(recordID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return Error("借阅记录不存在")
		}
		return err
	}
	if rec.Status != "borrowed" {
		return Error("该记录已归还，无需续借")
	}
	if ownerReaderID != nil && rec.ReaderID != *ownerReaderID {
		return Error("只能续借本人的图书")
	}
	todayStr := time.Now().Format(dateLayout)
	if err := checkRenew(rec.RenewCount, rec.DueDate, todayStr); err != nil {
		return err
	}
	due, _ := time.Parse(dateLayout, rec.DueDate)
	newDue := due.AddDate(0, 0, RenewDays).Format(dateLayout)
	if err := s.repo.MarkRenewed(recordID, newDue, rec.RenewCount+1); err != nil {
		return Error("续借失败")
	}
	return s.log(actor.ID, actor.Username, "续借图书", "《"+rec.Title+"》")
}

// SelfRenew 读者自助续借。
func (s *Service) SelfRenew(actor *models.User, recordID int64) error {
	reader, err := s.repo.GetReaderByUserID(actor.ID)
	if err != nil {
		return Error("读者档案不存在")
	}
	id := reader.ID
	return s.Renew(actor, recordID, &id)
}

// Return 办理归还，返回应缴罚款。
func (s *Service) Return(actor *models.User, recordID int64) (float64, error) {
	rec, err := s.repo.GetBorrow(recordID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return 0, Error("借阅记录不存在")
		}
		return 0, err
	}
	if rec.Status != "borrowed" {
		return 0, Error("该记录已归还，请勿重复操作")
	}
	todayStr := time.Now().Format(dateLayout)
	fine := calcFine(rec.DueDate, todayStr)
	if err := s.repo.MarkReturned(recordID, todayStr, fine); err != nil {
		return 0, Error("归还失败")
	}
	if err := s.repo.IncAvailable(rec.BookID); err != nil {
		return 0, err
	}
	detail := "归还《" + rec.Title + "》"
	if fine > 0 {
		detail += "，逾期罚款 " + formatFloat(fine) + " 元"
	}
	_ = s.log(actor.ID, actor.Username, "归还图书", detail)
	return fine, nil
}

// ActiveBorrows 查询未归还借阅；readerID 为 nil 时查询全部（管理员）。
func (s *Service) ActiveBorrows(readerID *int64) ([]models.BorrowRecord, error) {
	recs, err := s.repo.ListActive(readerID)
	if err != nil {
		return nil, err
	}
	return annotate(recs), nil
}

// HistoryBorrows 查询已归还历史；readerID 为 nil 时查询全部（管理员）。
func (s *Service) HistoryBorrows(readerID *int64) ([]models.BorrowRecord, error) {
	recs, err := s.repo.ListHistory(readerID)
	if err != nil {
		return nil, err
	}
	return annotate(recs), nil
}

// MyReader 返回当前登录用户对应的读者档案。
func (s *Service) MyReader(actor *models.User) (*models.Reader, error) {
	r, err := s.repo.GetReaderByUserID(actor.ID)
	if err != nil {
		return nil, Error("读者档案不存在")
	}
	return r, nil
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
