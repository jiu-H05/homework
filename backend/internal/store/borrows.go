package store

import (
	"database/sql"
	"errors"

	"lms/backend/internal/models"
)

const borrowSelect = "SELECT br.id,br.reader_id,br.book_id,br.borrow_date,br.due_date," +
	"br.return_date,br.renew_count,br.status,br.fine," +
	"b.title,b.author,r.name,r.reader_no " +
	"FROM borrow_records br JOIN books b ON br.book_id=b.id " +
	"JOIN readers r ON br.reader_id=r.id "

func scanBorrow(sc interface{ Scan(...any) error }) (models.BorrowRecord, error) {
	var r models.BorrowRecord
	var returnDate sql.NullString
	err := sc.Scan(&r.ID, &r.ReaderID, &r.BookID, &r.BorrowDate, &r.DueDate,
		&returnDate, &r.RenewCount, &r.Status, &r.Fine,
		&r.Title, &r.Author, &r.ReaderName, &r.ReaderNo)
	r.ReturnDate = returnDate.String
	return r, err
}

// CreateBorrow 写入借阅记录并返回主键。
func (s *Store) CreateBorrow(r *models.BorrowRecord) error {
	res, err := s.db.Exec(
		"INSERT INTO borrow_records(reader_id,book_id,borrow_date,due_date,renew_count,status,fine) "+
			"VALUES(?,?,?,?,?,?,?)",
		r.ReaderID, r.BookID, r.BorrowDate, r.DueDate, r.RenewCount, r.Status, r.Fine)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	r.ID = id
	return nil
}

// GetBorrow 按主键查询借阅记录（含关联信息）。
func (s *Store) GetBorrow(id int64) (*models.BorrowRecord, error) {
	row := s.db.QueryRow(borrowSelect+"WHERE br.id=?", id)
	r, err := scanBorrow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Store) listBorrow(q string, args ...any) ([]models.BorrowRecord, error) {
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.BorrowRecord{}
	for rows.Next() {
		r, err := scanBorrow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListActive 查询未归还借阅；readerID 为 nil 时查询全部。
func (s *Store) ListActive(readerID *int64) ([]models.BorrowRecord, error) {
	if readerID != nil {
		return s.listBorrow(borrowSelect+"WHERE br.status='borrowed' AND br.reader_id=? ORDER BY br.due_date",
			*readerID)
	}
	return s.listBorrow(borrowSelect + "WHERE br.status='borrowed' ORDER BY br.due_date")
}

// ListHistory 查询已归还历史；readerID 为 nil 时查询全部。
func (s *Store) ListHistory(readerID *int64) ([]models.BorrowRecord, error) {
	if readerID != nil {
		return s.listBorrow(borrowSelect+"WHERE br.status='returned' AND br.reader_id=? ORDER BY br.return_date DESC",
			*readerID)
	}
	return s.listBorrow(borrowSelect + "WHERE br.status='returned' ORDER BY br.return_date DESC")
}

// MarkReturned 标记归还、归还日期与罚款。
func (s *Store) MarkReturned(id int64, returnDate string, fine float64) error {
	_, err := s.db.Exec(
		"UPDATE borrow_records SET status='returned',return_date=?,fine=? WHERE id=?",
		returnDate, fine, id)
	return err
}

// MarkRenewed 更新续借次数与新应还日期。
func (s *Store) MarkRenewed(id int64, dueDate string, renewCount int) error {
	_, err := s.db.Exec(
		"UPDATE borrow_records SET renew_count=?,due_date=? WHERE id=?", renewCount, dueDate, id)
	return err
}

// CountActive 统计读者在借数量。
func (s *Store) CountActive(readerID int64) (int, error) {
	var n int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM borrow_records WHERE reader_id=? AND status='borrowed'", readerID).Scan(&n)
	return n, err
}

// CountOverdue 统计读者逾期数量。
func (s *Store) CountOverdue(readerID int64, today string) (int, error) {
	var n int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM borrow_records WHERE reader_id=? AND status='borrowed' AND due_date<?",
		readerID, today).Scan(&n)
	return n, err
}

// AddLog 写入操作日志。
func (s *Store) AddLog(userID *int64, username, action, detail, createdAt string) error {
	_, err := s.db.Exec(
		"INSERT INTO operation_logs(user_id,username,action,detail,created_at) VALUES(?,?,?,?,?)",
		userID, username, action, detail, createdAt)
	return err
}

// Dashboard 返回控制台汇总指标。
func (s *Store) Dashboard(today string) (*models.Dashboard, error) {
	d := &models.Dashboard{}
	one := func(q string, p ...any) (int, error) {
		var n int
		return n, s.db.QueryRow(q, p...).Scan(&n)
	}
	var err error
	if d.BookKinds, err = one("SELECT COUNT(*) FROM books"); err != nil {
		return nil, err
	}
	if d.BookCopies, err = one("SELECT COALESCE(SUM(total_qty),0) FROM books"); err != nil {
		return nil, err
	}
	if d.Readers, err = one("SELECT COUNT(*) FROM readers"); err != nil {
		return nil, err
	}
	if d.Borrowed, err = one("SELECT COUNT(*) FROM borrow_records WHERE status='borrowed'"); err != nil {
		return nil, err
	}
	if d.Overdue, err = one(
		"SELECT COUNT(*) FROM borrow_records WHERE status='borrowed' AND due_date<?", today); err != nil {
		return nil, err
	}
	return d, nil
}

// CategoryStats 返回分类馆藏统计。
func (s *Store) CategoryStats() ([]models.CategoryStat, error) {
	rows, err := s.db.Query(
		"SELECT c.name,COUNT(b.id),COALESCE(SUM(b.total_qty),0) " +
			"FROM categories c LEFT JOIN books b ON b.category_id=c.id GROUP BY c.id ORDER BY 3 DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CategoryStat{}
	for rows.Next() {
		var c models.CategoryStat
		if err := rows.Scan(&c.Name, &c.Kinds, &c.Copies); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// HotBooks 返回借阅热度排行。
func (s *Store) HotBooks(limit int) ([]models.HotBook, error) {
	rows, err := s.db.Query(
		"SELECT b.title,b.author,COUNT(br.id) FROM books b "+
			"LEFT JOIN borrow_records br ON br.book_id=b.id GROUP BY b.id ORDER BY 3 DESC,b.id LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.HotBook{}
	for rows.Next() {
		var h models.HotBook
		if err := rows.Scan(&h.Title, &h.Author, &h.Times); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// RecentLogs 返回最近操作日志。
func (s *Store) RecentLogs(limit int) ([]models.OperationLog, error) {
	rows, err := s.db.Query(
		"SELECT id,user_id,username,action,detail,created_at FROM operation_logs ORDER BY id DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.OperationLog{}
	for rows.Next() {
		var l models.OperationLog
		var uid sql.NullInt64
		if err := rows.Scan(&l.ID, &uid, &l.Username, &l.Action, &l.Detail, &l.CreatedAt); err != nil {
			return nil, err
		}
		if uid.Valid {
			id := uid.Int64
			l.UserID = id
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
