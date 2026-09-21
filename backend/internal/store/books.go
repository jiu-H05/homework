package store

import (
	"database/sql"
	"errors"
	"strings"

	"lms/backend/internal/models"
)

const bookSelect = `
SELECT b.id, b.isbn, b.title, b.author, b.publisher,
       b.category_id, c.name, b.region, b.location,
       b.total_qty, b.available_qty, b.created_at
FROM books b
LEFT JOIN categories c ON c.id = b.category_id
`

func scanBook(sc interface {
	Scan(dest ...any) error
}) (models.Book, error) {
	var b models.Book
	var catID sql.NullInt64
	var catName, isbn, author, publisher, region, location, createdAt sql.NullString
	if err := sc.Scan(
		&b.ID, &isbn, &b.Title, &author, &publisher,
		&catID, &catName, &region, &location,
		&b.TotalQty, &b.AvailableQty, &createdAt,
	); err != nil {
		return b, err
	}
	b.ISBN = isbn.String
	b.Author = author.String
	b.Publisher = publisher.String
	b.Region = region.String
	b.Location = location.String
	b.CreatedAt = createdAt.String
	b.CategoryName = catName.String
	if catID.Valid {
		v := catID.Int64
		b.CategoryID = &v
	}
	return b, nil
}

// ListBooks 按关键字 / 题材分类 / 地区过滤并返回全部图书。
func (s *Store) ListBooks(keyword string, categoryID *int64, region string) ([]models.Book, error) {
	var sb strings.Builder
	sb.WriteString(bookSelect)
	var args []any
	var where []string
	if keyword != "" {
		where = append(where, "(b.title LIKE ? OR b.author LIKE ? OR b.isbn LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}
	if categoryID != nil {
		where = append(where, "b.category_id = ?")
		args = append(args, *categoryID)
	}
	if region != "" {
		where = append(where, "b.region = ?")
		args = append(args, region)
	}
	if len(where) > 0 {
		sb.WriteString(" WHERE " + strings.Join(where, " AND "))
	}
	sb.WriteString(" ORDER BY b.id")
	rows, err := s.db.Query(sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Book{}
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// GetBook 按 ID 返回单本图书；不存在时返回 models.ErrNotFound。
func (s *Store) GetBook(id int64) (*models.Book, error) {
	row := s.db.QueryRow(bookSelect+" WHERE b.id = ?", id)
	b, err := scanBook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// CreateBook 新增图书并回填 ID。
func (s *Store) CreateBook(b *models.Book) error {
	qty := b.TotalQty
	if qty < 0 {
		qty = 0
	}
	res, err := s.db.Exec(`INSERT INTO books
(isbn,title,author,publisher,category_id,region,location,total_qty,available_qty,created_at)
VALUES (?,?,?,?,?,?,?,?,?,?)`,
		b.ISBN, b.Title, b.Author, b.Publisher,
		b.CategoryID, b.Region, b.Location, qty, qty, nowStr())
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	b.ID = id
	return nil
}

// UpdateBook 更新图书；可借数 = 新总数 - 在借册数，保证不影响已借出图书。
func (s *Store) UpdateBook(b *models.Book) error {
	var oldTotal, oldAvail int
	if err := s.db.QueryRow("SELECT total_qty, available_qty FROM books WHERE id = ?", b.ID).
		Scan(&oldTotal, &oldAvail); err != nil {
		return err
	}
	borrowed := oldTotal - oldAvail
	newTotal := b.TotalQty
	if newTotal < 0 {
		newTotal = 0
	}
	newAvail := newTotal - borrowed
	if newAvail < 0 {
		newAvail = 0
	}
	_, err := s.db.Exec(`UPDATE books
SET isbn=?, title=?, author=?, publisher=?, category_id=?, region=?, location=?,
    total_qty=?, available_qty=?
WHERE id=?`,
		b.ISBN, b.Title, b.Author, b.Publisher,
		b.CategoryID, b.Region, b.Location, newTotal, newAvail, b.ID)
	return err
}

// DeleteBook 删除图书；存在未归还记录时由上层拦截。
func (s *Store) DeleteBook(id int64) error {
	_, err := s.db.Exec("DELETE FROM books WHERE id = ?", id)
	return err
}

// CountActiveBorrowsForBook 统计某书未归还的借阅记录数。
func (s *Store) CountActiveBorrowsForBook(id int64) (int, error) {
	var n int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM borrow_records WHERE book_id=? AND status='borrowed'", id).Scan(&n)
	return n, err
}

// DecAvailable 借出时将可借数减 1。
func (s *Store) DecAvailable(id int64) error {
	_, err := s.db.Exec(
		"UPDATE books SET available_qty = available_qty - 1 WHERE id=? AND available_qty > 0", id)
	return err
}

// IncAvailable 归还时将可借数加 1。
func (s *Store) IncAvailable(id int64) error {
	_, err := s.db.Exec(
		"UPDATE books SET available_qty = available_qty + 1 WHERE id=? AND available_qty < total_qty", id)
	return err
}

// ListRegions 按种子地区顺序返回各地区的馆藏种类数。
func (s *Store) ListRegions() ([]models.Region, error) {
	rows, err := s.db.Query("SELECT region, COUNT(*) FROM books GROUP BY region")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := map[string]int{}
	for rows.Next() {
		var name string
		var n int
		if err := rows.Scan(&name, &n); err != nil {
			return nil, err
		}
		counts[name] = n
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]models.Region, 0, len(seedRegions))
	known := map[string]bool{}
	for _, name := range seedRegions {
		known[name] = true
		out = append(out, models.Region{Name: name, Kinds: counts[name]})
	}
	for name, n := range counts {
		if name != "" && !known[name] {
			out = append(out, models.Region{Name: name, Kinds: n})
		}
	}
	return out, nil
}
