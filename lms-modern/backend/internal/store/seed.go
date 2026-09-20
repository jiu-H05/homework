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

	// 分类
	cats := []string{"计算机", "文学", "历史", "自然科学", "艺术", "哲学"}
	catID := map[string]int64{}
	for _, c := range cats {
		r, err := tx.Exec("INSERT INTO categories(name) VALUES(?)", c)
		if err != nil {
			return err
		}
		id, _ := r.LastInsertId()
		catID[c] = id
	}

	// 图书
	books := []struct {
		isbn, title, author, publisher, cat, loc string
		qty                                       int
	}{
		{"9787111213826", "软件工程：实践者的研究方法", "Roger S. Pressman", "机械工业出版社", "计算机", "N502-A01", 5},
		{"9787115428028", "Python编程：从入门到实践", "Eric Matthes", "人民邮电出版社", "计算机", "N502-A02", 4},
		{"9787302259862", "算法导论", "Thomas H. Cormen", "清华大学出版社", "计算机", "N502-A03", 3},
		{"9787020008735", "红楼梦", "曹雪芹", "人民文学出版社", "文学", "N503-B01", 6},
		{"9787020008728", "西游记", "吴承恩", "人民文学出版社", "文学", "N503-B02", 6},
		{"9787101003048", "史记", "司马迁", "中华书局", "历史", "N504-C01", 3},
		{"9787535732224", "时间简史", "Stephen Hawking", "湖南科学技术出版社", "自然科学", "N505-D01", 4},
		{"9787544270878", "万物简史", "Bill Bryson", "接力出版社", "自然科学", "N505-D02", 3},
		{"9787102048161", "中国艺术史", "苏立文", "上海人民出版社", "艺术", "N506-E01", 2},
		{"9787544720885", "西方哲学史", "Bertrand Russell", "译林出版社", "哲学", "N505-D03", 3},
	}
	bookID := map[string]int64{}
	for _, b := range books {
		r, err := tx.Exec(
			"INSERT INTO books(isbn,title,author,publisher,category_id,location,total_qty,available_qty,created_at) "+
				"VALUES(?,?,?,?,?,?,?,?,?)",
			b.isbn, b.title, b.author, b.publisher, catID[b.cat], b.loc, b.qty, b.qty, nowStr())
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
