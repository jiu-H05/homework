package store

import (
	"database/sql"
	"errors"

	"lms/backend/internal/models"
)

const bookSelect = "SELECT b.id,b.isbn,b.title,b.author,b.publisher,b.category_id," +
	"c.name,b.location,b.total_qty,b.available_qty,b.created_at " +
	"FROM books b LEFT JOIN categories c ON b.category_id=c.id "

func scanBook(sc interface{ Scan(...any) error }) (models.Book, error) {
	var b models.Book
	var isbn, author, publisher, location, catName sql.NullString
	var catID sql.NullInt64
	err := sc.Scan(&b.ID, &isbn, &b.Title, &b.Author, &publisher, &catID,
		&catName, &location, &b.TotalQty, &b.AvailableQty, &b.CreatedAt)
	b.ISBN = isbn.String
	b.Author = author.String
	b.Publisher = publisher.String
	b.Location = location.String
	b.CategoryName = catName.String
	if catID.Valid {
		id := catID.Int64
		b.CategoryID = &id
	}
	return b, err
}

// ListBooks 按关键字与分类查询图书。
func (s *Store) ListBooks(keyword string, categoryID *int64) {
	q := bookSelect + "WHERE 1=1 "
	args := []any{}
	if categoryID != nil {
		q += "AND b.category_id=? "
		args = append(args, *categoryID)
	}
	if keyword != "" {
		q += "AND (b.title LIKE ? OR b.author LIKE ? OR b.isbn LIKE ?) "
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}
	q += "ORDER BY b.id"
	rows, err := s.db.Query(q, args...)
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

// GetBook 按主键查询图书。
func (s *Store) GetBook(id int64) (*models.Book, error) {
	row := s.db.QueryRow(bookSelect+"WHERE b.id=?", id)
	b, err := scanBook(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// CreateBook 新增图书。
func (s *Store) CreateBook(b *models.Book) error {
	res, err := s.db.Exec(
		"INSERT INTO books(isbn,title,author,publisher,category_id,location,total_qty,available_qty,created_at) "+
			"VALUES(?,?,?,?,?,?,?,?,?)",
		b.ISBN, b.Title, b.Author, b.Publisher, b.CategoryID, b.Location,
		b.TotalQty, b.TotalQty, nowStr())
	id, _ := res.LastInsertId()
	b.ID = id
	b.AvailableQty = b.TotalQty
	return nil
}

// UpdateBook 修改图书信息；馆藏数量调整时同步在馆数量（已借出数量保持不变）。
func (s *Store) UpdateBook(b *models.Book) error {
	var borrowed int
	err := s.db.QueryRow("SELECT total_qty-available_qty FROM books WHERE id=?", b.ID).Scan(&borrowed)
	if err != nil {
		return err
	}
	available := b.TotalQty - borrowed
	_, err = s.db.Exec(
		"UPDATE books SET isbn=?,title=?,author=?,publisher=?,category_id=?,location=?," +
			"total_qty=?,available_qty=? WHERE id=?",
		b.ISBN, b.Title, b.Author, b.Publisher, b.CategoryID, b.Location,
		b.TotalQty, available, b.ID)
	return err
}

// DeleteBook 删除无在借记录的图书。
func (s *Store) DeleteBook(id int64) error {
	_, err := s.db.Exec("DELETE FROM books WHERE id=?", id)
	return err
}

// CountActiveBorrowsForBook 统计某图书未归还借阅数量。
func (s *Store) CountActiveBorrowsForBook(id int64) (int, error) {
	var n int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM borrow_records WHERE book_id=? AND status='borrowed'", id).Scan(&n)
	return n, err
}

// DecAvailable 借出时库存减一。
func (s *Store) DecAvailable(bookID int64) error {
	_, err := s.db.Exec("UPDATE books SET available_qty=available_qty-1 WHERE id=?", bookID)
	return err
}

// IncAvailable 归还时库存加一。
func (s *Store) IncAvailable(bookID int64) error {
	_, err := s.db.Exec("UPDATE books SET available_qty=available_qty+1 WHERE id=?", bookID)
	return err
}
