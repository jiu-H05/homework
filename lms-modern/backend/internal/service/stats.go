package service

import (
	"time"

	"lms/backend/internal/models"
)

// Dashboard 返回控制台汇总指标。
func (s *Service) Dashboard() (*models.Dashboard, error) {
	return s.repo.Dashboard(time.Now().Format(dateLayout))
}

// CategoryStats 返回分类馆藏统计。
func (s *Service) CategoryStats() ([]models.CategoryStat, error) {
	return s.repo.CategoryStats()
}

// HotBooks 返回图书借阅热度排行。
func (s *Service) HotBooks() ([]models.HotBook, error) {
	return s.repo.HotBooks(10)
}

// RecentLogs 返回最近操作日志。
func (s *Service) RecentLogs() ([]models.OperationLog, error) {
	return s.repo.RecentLogs(100)
}
