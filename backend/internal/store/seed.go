package store

import (
	"database/sql"
	"time"

	"lms/backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func nowStr() string { return time.Now().Format("2006-01-02 15:04:05") }

func hashBcrypt(pwd string) string {
	b, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(b)
}

// seedIfEmpty 在数据库首次创建时写入默认账号、分类、图书与示例借阅。
func (s *Store) seedIfEmpty() error {
	var n int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 管理员
	_, err = tx.Exec(
		"INSERT INTO users(username,password_hash,role,real_name,status,created_at) VALUES(?,?,?,?,?,?)",
		"admin", hashBcrypt("admin123"), "admin", "系统管理员", "active", nowStr())
	if err != nil {
		return err
	}
	// 示例读者
	res, err := tx.Exec(
		"INSERT INTO users(username,password_hash,role,real_name,status,created_at) VALUES(?,?,?,?,?,?)",
		"reader", hashBcrypt("reader123"), "reader", "张同学", "active", nowStr())
	if err != nil {
		return err
	}
	uid, _ := res.LastInsertId()
	_, err = tx.Exec(
		"INSERT INTO readers(user_id,reader_no,name,phone,created_at) VALUES(?,?,?,?,?)",
		uid, "R2023001", "张同学", "13800000000", nowStr())
	if err != nil {
		return err
	}
	var readerID int64
	err = tx.QueryRow("SELECT id FROM readers WHERE reader_no='R2023001'").Scan(&readerID)
	if err != nil {
		return err
	}

	// 分类（来自 seed_catalog.go，共 26 个）
	catID := map[string]int64{}
	for _, c := range seedCategories {
		r, err := tx.Exec("INSERT INTO categories(name) VALUES(?)", c)
		if err != nil {
			return err
		}
		id, _ := r.LastInsertId()
		catID[c] = id
	}

	// 图书（来自 seed_catalog.go，共 127 本）
	bookID := map[string]int64{}
	for _, b := range seedBooks {
		r, err := tx.Exec(
			"INSERT INTO books(isbn,title,author,publisher,category_id,region,location,total_qty,available_qty,created_at) "+
				"VALUES(?,?,?,?,?,?,?,?,?,?)",
			b.isbn, b.title, b.author, b.publisher, catID[b.cat], b.region, b.loc, b.qty, b.qty, nowStr())
		if err != nil {
			return err
		}
		id, _ := r.LastInsertId()
		bookID[b.title] = id
	}

	// 示例借阅：一本正常（5天前），一本逾期（40天前，应还10天前）
	today := time.Now()
	b1 := today.AddDate(0, 0, -5)
	d1 := b1.AddDate(0, 0, 30)
	if _, err := tx.Exec(
		"INSERT INTO borrow_records(reader_id,book_id,borrow_date,due_date,renew_count,status,fine) VALUES(?,?,?,?,?,?,?)",
		readerID, bookID["Python编程：从入门到实践"], b1.Format("2006-01-02"), d1.Format("2006-01-02"), 0, "borrowed", 0); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE books SET available_qty=available_qty-1 WHERE id=?",
		bookID["Python编程：从入门到实践"]); err != nil {
		return err
	}
	b2 := today.AddDate(0, 0, -40)
	d2 := b2.AddDate(0, 0, 30)
	if _, err := tx.Exec(
		"INSERT INTO borrow_records(reader_id,book_id,borrow_date,due_date,renew_count,status,fine) VALUES(?,?,?,?,?,?,?)",
		readerID, bookID["时间简史"], b2.Format("2006-01-02"), d2.Format("2006-01-02"), 0, "borrowed", 0); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE books SET available_qty=available_qty-1 WHERE id=?", bookID["时间简史"]); err != nil {
		return err
	}
	return tx.Commit()
}

var _ = sql.ErrNoRows
var _ models.User
