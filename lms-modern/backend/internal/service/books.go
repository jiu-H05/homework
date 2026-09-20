package service

import (
	"errors"
	"strings"

	"lms/backend/internal/models"
)

// ListBooks 查询图书。
func (s *Service) ListBooks(keyword string, categoryID *int64) ([]models.Book, error) {
	return s.repo.ListBooks(strings.TrimSpace(keyword), categoryID)
}

// AddBook 新增图书。
func (s *Service) AddBook(actor *models.User, b *models.Book) error {
	b.Title = strings.TrimSpace(b.Title)
	if b.Title == "" {
		return Error("书名不能为空")
	}
	if b.TotalQty < 1 {
		return Error("馆藏数量至少为 1")
	}
	if err := s.repo.CreateBook(b); err != nil {
		return Error("新增图书失败")
	}
	return s.log(actor.ID, actor.Username, "新增图书", b.Title)
}

// UpdateBook 修改图书。
func (s *Service) UpdateBook(actor *models.User, b *models.Book) error {
	b.Title = strings.TrimSpace(b.Title)
	if b.Title == "" {
		return Error("书名不能为空")
	}
	old, err := s.repo.GetBook(b.ID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return Error("图书不存在")
		}
		return err
	}
	borrowed := old.TotalQty - old.AvailableQty
	if err := checkTotalQty(b.TotalQty, borrowed); err != nil {
		return err
	}
	if err := s.repo.UpdateBook(b); err != nil {
		return Error("修改图书失败")
	}
	return s.log(actor.ID, actor.Username, "修改图书", b.Title)
}

// DeleteBook 删除图书（存在在借记录时拒绝）。
func (s *Service) DeleteBook(actor *models.User, id int64) error {
	book, err := s.repo.GetBook(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return Error("图书不存在")
		}
		return err
	}
	n, err := s.repo.CountActiveBorrowsForBook(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return Error("该图书存在未归还的借阅记录，无法删除")
	}
	if err := s.repo.DeleteBook(id); err != nil {
		return Error("删除图书失败")
	}
	return s.log(actor.ID, actor.Username, "删除图书", book.Title)
}
