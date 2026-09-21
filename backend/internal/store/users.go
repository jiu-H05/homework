package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"lms/backend/internal/models"
)

// ErrNotFound 兼容别名，实际统一使用 models.ErrNotFound。
var ErrNotFound = models.ErrNotFound

// GetUserByUsername 按用户名查询用户。
func (s *Store) GetUserByUsername(username string) (*models.User, error) {
	row := s.db.QueryRow(
		"SELECT id,username,password_hash,role,real_name,status,created_at FROM users WHERE username=?", username)
	u := &models.User{}
	var real sql.NullString
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &real, &u.Status, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	u.RealName = real.String
	return u, err
}

// GetUserByID 按主键查询用户。
func (s *Store) GetUserByID(id int64) (*models.User, error) {
	row := s.db.QueryRow(
		"SELECT id,username,password_hash,role,real_name,status,created_at FROM users WHERE id=?", id)
	u := &models.User{}
	var real sql.NullString
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &real, &u.Status, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	u.RealName = real.String
	return u, err
}

// CreateReaderAccount 创建读者账号与读者档案（同一事务），返回读者编号。
func (s *Store) CreateReaderAccount(username, passwordHash, realName, phone string) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(
		"INSERT INTO users(username,password_hash,role,real_name,status,created_at) VALUES(?,?,?,?,?,?)",
		username, passwordHash, "reader", realName, "active", nowStr())
	if err != nil {
		return "", err
	}
	uid, _ := res.LastInsertId()

	var n int64
	if err := tx.QueryRow("SELECT COUNT(*) FROM readers").Scan(&n); err != nil {
		return "", err
	}
	readerNo := "R" + time.Now().Format("2006") + fmt.Sprintf("%03d", n+1)
	if _, err := tx.Exec(
		"INSERT INTO readers(user_id,reader_no,name,phone,created_at) VALUES(?,?,?,?,?)",
		uid, readerNo, realName, phone, nowStr()); err != nil {
		return "", err
	}
	return readerNo, tx.Commit()
}

// UpdatePassword 更新用户口令摘要。
func (s *Store) UpdatePassword(userID int64, hash string) error {
	_, err := s.db.Exec("UPDATE users SET password_hash=? WHERE id=?", hash, userID)
	return err
}

// SetUserStatus 启用/禁用用户。
func (s *Store) SetUserStatus(userID int64, status string) error {
	_, err := s.db.Exec("UPDATE users SET status=? WHERE id=?", status, userID)
	return err
}

// ResetPassword 将用户口令重置为给定摘要。
func (s *Store) ResetPassword(userID int64, hash string) error {
	return s.UpdatePassword(userID, hash)
}

// ListReaders 关键字查询读者（含账号信息）。
func (s *Store) ListReaders(keyword string) ([]models.ReaderWithUser, error) {
	q := "SELECT u.id, r.reader_no, r.name, r.phone, u.username, u.status, r.created_at " +
		"FROM readers r JOIN users u ON r.user_id=u.id "
	var rows *sql.Rows
	var err error
	if keyword != "" {
		kw := "%" + keyword + "%"
		rows, err = s.db.Query(q+"WHERE r.name LIKE ? OR r.reader_no LIKE ? OR u.username LIKE ? ORDER BY r.id",
			kw, kw, kw)
	} else {
		rows, err = s.db.Query(q + "ORDER BY r.id")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.ReaderWithUser{}
	for rows.Next() {
		var r models.ReaderWithUser
		var phone sql.NullString
		if err := rows.Scan(&r.UserID, &r.ReaderNo, &r.Name, &phone, &r.Username, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.Phone = phone.String
		out = append(out, r)
	}
	return out, rows.Err()
}

// GetReaderByUserID 按登录用户 ID 查询读者档案。
func (s *Store) GetReaderByUserID(userID int64) (*models.Reader, error) {
	row := s.db.QueryRow(
		"SELECT id,user_id,reader_no,name,phone,created_at FROM readers WHERE user_id=?", userID)
	r := &models.Reader{}
	var phone sql.NullString
	err := row.Scan(&r.ID, &r.UserID, &r.ReaderNo, &r.Name, &phone, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	r.Phone = phone.String
	return r, err
}

// GetReaderByID 按读者主键查询。
func (s *Store) GetReaderByID(id int64) (*models.Reader, error) {
	row := s.db.QueryRow(
		"SELECT id,user_id,reader_no,name,phone,created_at FROM readers WHERE id=?", id)
	r := &models.Reader{}
	var phone sql.NullString
	err := row.Scan(&r.ID, &r.UserID, &r.ReaderNo, &r.Name, &phone, &r.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	r.Phone = phone.String
	return r, err
}

// ListCategories 查询全部分类。
func (s *Store) ListCategories() ([]models.Category, error) {
	rows, err := s.db.Query("SELECT id,name FROM categories ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// CreateCategory 新增分类。
func (s *Store) CreateCategory(name string) error {
	_, err := s.db.Exec("INSERT INTO categories(name) VALUES(?)", name)
	return err
}
