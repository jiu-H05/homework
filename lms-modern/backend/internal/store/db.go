// Package store 是持久化适配器：基于 SQLite 实现 service 层定义的仓储端口。
package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // 纯 Go SQLite 驱动，无需 CGO
)

// Store 持有数据库连接并实现全部仓储方法。
type Store struct {
	db *sql.DB
}

// New 打开（必要时创建）数据库并执行建表与种子数据初始化。
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite 单写，避免锁竞争
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("建表失败: %w", err)
	}
	s := &Store{db: db}
	if err := s.seedIfEmpty(); err != nil {
		return nil, err
	}
	return s, nil
}

// Close 关闭数据库连接。
func (s *Store) Close() error { return s.db.Close() }

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('admin','reader')),
    real_name TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active','disabled')),
    created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS readers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reader_no TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    phone TEXT,
    created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    isbn TEXT,
    title TEXT NOT NULL,
    author TEXT,
    publisher TEXT,
    category_id INTEGER REFERENCES categories(id),
    location TEXT,
    total_qty INTEGER NOT NULL DEFAULT 1,
    available_qty INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS borrow_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    reader_id INTEGER NOT NULL REFERENCES readers(id),
    book_id INTEGER NOT NULL REFERENCES books(id),
    borrow_date TEXT NOT NULL,
    due_date TEXT NOT NULL,
    return_date TEXT,
    renew_count INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'borrowed' CHECK(status IN ('borrowed','returned')),
    fine REAL NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS operation_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    username TEXT,
    action TEXT NOT NULL,
    detail TEXT,
    created_at TEXT NOT NULL
);
`
