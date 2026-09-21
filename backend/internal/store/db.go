// Package store 是持久化适配器：基于 SQLite 实现 service 层定义的仓储端口。
package store

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // 纯 Go SQLite 驱动，无需 CGO
)

// Store 持有数据库连接并实现全部仓储方法。
type Store struct {
	db   *sql.DB
	path string // 数据库文件绝对路径，用于导出/导入
}

// New 打开（必要时创建）数据库并执行建表与种子数据初始化。
func New(dbPath string) (*Store, error) {
	abs, err := filepath.Abs(dbPath)
	if err != nil {
		abs = dbPath
	}
	s := &Store{path: abs}
	if err := s.open(); err != nil {
		return nil, err
	}
	return s, nil
}

// open 打开连接并建表、补列、种子。Restore 后复用此方法热加载新库。
func (s *Store) open() error {
	db, err := sql.Open("sqlite", s.path)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	db.SetMaxOpenConns(1) // SQLite 单写，避免锁竞争
	// WAL 模式：多人同时在线时读写不互相阻塞，适合前后端分离的多客户端场景。
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return fmt.Errorf("开启 WAL 失败: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return fmt.Errorf("建表失败: %w", err)
	}
	if err := ensureColumn(db, "books", "region", "TEXT"); err != nil {
		_ = db.Close()
		return fmt.Errorf("迁移失败: %w", err)
	}
	s.db = db
	return s.seedIfEmpty()
}

// Backup 把当前数据库文件完整复制到 dest（供管理员导出下载）。
// 单连接模式下读文件与业务写入互斥，文件级拷贝即为一致性快照。
func (s *Store) Backup(dest string) error {
	// 先 checkpoint，确保无未写盘内容；modernc 走 rollback journal，这里直接文件拷贝即可。
	if _, err := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return err
	}
	in, err := os.Open(s.path)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// Restore 用 src 这个合法 SQLite 文件替换当前数据库：
// 校验 -> 关闭旧连接 -> 覆盖文件 -> 重新打开。
func (s *Store) Restore(src string) error {
	// 1. 校验 src 是本系统数据库：能打开且含 books 表。
	if err := validateBackup(src); err != nil {
		return err
	}
	// 2. 关闭旧连接。
	if err := s.db.Close(); err != nil {
		return err
	}
	// 3. 覆盖到当前 db 路径。
	in, err := os.Open(src)
	if err != nil {
		_ = s.open() // 尽量恢复
		return err
	}
	defer in.Close()
	out, err := os.Create(s.path)
	if err != nil {
		_ = s.open()
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = s.open()
		return err
	}
	_ = out.Close()
	// 清掉可能存在的 -wal/-shm，避免旧快照残留。
	_ = os.Remove(s.path + "-wal")
	_ = os.Remove(s.path + "-shm")
	// 4. 重新打开新库。
	return s.open()
}

// validateBackup 打开候选文件，确认它是带 books 表的合法库。
func validateBackup(p string) error {
	c, err := sql.Open("sqlite", p)
	if err != nil {
		return fmt.Errorf("文件不是有效的数据库")
	}
	defer c.Close()
	var name string
	err = c.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='books'").Scan(&name)
	if err != nil {
		return fmt.Errorf("这不是图书管理系统的备份文件")
	}
	return nil
}

// DBPath 返回数据库文件路径。
func (s *Store) DBPath() string { return s.path }

// BackupToFile / RestoreFromFile 实现 service.Repo 接口。
func (s *Store) BackupToFile(dest string) error  { return s.Backup(dest) }
func (s *Store) RestoreFromFile(src string) error { return s.Restore(src) }

// ensureColumn 在表缺少指定列时执行 ALTER TABLE ADD COLUMN（SQLite 不支持 IF NOT EXISTS）。
func ensureColumn(db *sql.DB, table, column, typ string) error {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			return nil // 列已存在
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	stmt := "ALTER TABLE " + table + " ADD COLUMN " + column + " " + typ
	_, err = db.Exec(stmt)
	return err
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
    region TEXT DEFAULT '',
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
