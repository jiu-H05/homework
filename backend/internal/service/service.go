// Package service 承载全部业务规则，只依赖自行定义的仓储端口（Repo），
// 不直接依赖 SQLite 或 HTTP，实现与具体技术的解耦。
package service

import (
	"lms/backend/internal/models"
)

// 业务常量
const (
	LoanDays    = 30   // 借期（天）
	MaxBorrow   = 5    // 每人最大在借数量
	MaxRenew    = 1    // 每本最大续借次数
	RenewDays   = 30   // 续借延长天数
	FinePerDay  = 0.5  // 逾期罚款（元/天）
)

// Error 业务规则错误，可直接向用户展示。
type Error string

func (e Error) Error() string { return string(e) }

// Repo 仓储端口：由存储层（SQLite 适配器）实现。
type Repo interface {
	// 用户 / 读者 / 分类
	GetUserByUsername(string) (*models.User, error)
	GetUserByID(int64) (*models.User, error)
	CreateReaderAccount(username, passwordHash, realName, phone string) (string, error)
	UpdatePassword(int64, string) error
	SetUserStatus(int64, string) error
	ResetPassword(int64, string) error
	ListReaders(string) ([]models.ReaderWithUser, error)
	GetReaderByUserID(int64) (*models.Reader, error)
	GetReaderByID(int64) (*models.Reader, error)
	ListCategories() ([]models.Category, error)
	CreateCategory(string) error
	// 图书
	ListRegions() ([]models.Region, error)
	ListBooks(string, *int64, string) ([]models.Book, error)
	GetBook(int64) (*models.Book, error)
	CreateBook(*models.Book) error
	UpdateBook(*models.Book) error
	DeleteBook(int64) error
	CountActiveBorrowsForBook(int64) (int, error)
	DecAvailable(int64) error
	IncAvailable(int64) error
	// 借阅
	CreateBorrow(*models.BorrowRecord) error
	GetBorrow(int64) (*models.BorrowRecord, error)
	ListActive(*int64) ([]models.BorrowRecord, error)
	ListHistory(*int64) ([]models.BorrowRecord, error)
	MarkReturned(int64, string, float64) error
	MarkRenewed(int64, string, int) error
	CountActive(int64) (int, error)
	CountOverdue(int64, string) (int, error)
	AddLog(userID *int64, username, action, detail, createdAt string) error
	// 统计
	Dashboard(string) (*models.Dashboard, error)
	CategoryStats() ([]models.CategoryStat, error)
	HotBooks(int) ([]models.HotBook, error)
	RecentLogs(int) ([]models.OperationLog, error)
	// 备份与恢复
	BackupToFile(dest string) error
	RestoreFromFile(src string) error
}

// Service 业务服务入口。
type Service struct {
	repo Repo
}

// New 创建业务服务。
func New(repo Repo) *Service {
	return &Service{repo: repo}
}

// Backup 把当前数据库导出到 dest（供管理员下载备份文件）。
func (s *Service) Backup(dest string) error { return s.repo.BackupToFile(dest) }

// Restore 用 src 备份文件替换当前数据库，成功后连接已热加载到新库。
func (s *Service) Restore(src string) error { return s.repo.RestoreFromFile(src) }
