package service

import (
	"errors"
	"fmt"
	"strings"

	"lms/backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// ListReaders 管理员查询读者列表。
func (s *Service) ListReaders(keyword string) ([]models.ReaderWithUser, error) {
	return s.repo.ListReaders(strings.TrimSpace(keyword))
}

// AddReader 管理员新增读者，返回读者编号。
func (s *Service) AddReader(actor *models.User, username, password, realName, phone string) (string, error) {
	username = strings.TrimSpace(username)
	realName = strings.TrimSpace(realName)
	if len(username) < 3 {
		return "", Error("用户名至少 3 个字符")
	}
	if len(password) < 6 {
		return "", Error("初始口令至少 6 位")
	}
	if realName == "" {
		return "", Error("姓名不能为空")
	}
	if _, err := s.repo.GetUserByUsername(username); err == nil {
		return "", Error("用户名已存在")
	} else if !errors.Is(err, models.ErrNotFound) {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	no, err := s.repo.CreateReaderAccount(username, string(hash), realName, strings.TrimSpace(phone))
	if err != nil {
		return "", Error("新增失败：用户名可能已存在")
	}
	_ = s.log(actor.ID, actor.Username, "新增读者", realName+"（"+no+"）")
	return no, nil
}

// SetReaderStatus 管理员启用/停用读者账号。
func (s *Service) SetReaderStatus(actor *models.User, userID int64, status string) error {
	if status != "active" && status != "disabled" {
		return Error("状态值非法")
	}
	u, err := s.repo.GetUserByID(userID)
	if err != nil {
		return Error("用户不存在")
	}
	if u.Role != "reader" {
		return Error("只能对读者账号执行此操作")
	}
	if err := s.repo.SetUserStatus(userID, status); err != nil {
		return err
	}
	label := "启用"
	if status == "disabled" {
		label = "停用"
	}
	return s.log(actor.ID, actor.Username, label+"读者", u.RealName)
}

// ResetReaderPassword 管理员重置读者口令，返回新临时口令。
func (s *Service) ResetReaderPassword(actor *models.User, userID int64) (string, error) {
	u, err := s.repo.GetUserByID(userID)
	if err != nil {
		return "", Error("用户不存在")
	}
	if u.Role != "reader" {
		return "", Error("只能重置读者口令")
	}
	tmp := fmt.Sprintf("reset%06d", userID)
	hash, err := bcrypt.GenerateFromPassword([]byte(tmp), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	if err := s.repo.ResetPassword(userID, string(hash)); err != nil {
		return err
	}
	_ = s.log(actor.ID, actor.Username, "重置口令", u.RealName)
	return tmp, nil
}

// ListCategories 查询全部分类。
func (s *Service) ListCategories() ([]models.Category, error) {
	return s.repo.ListCategories()
}

// AddCategory 新增分类。
func (s *Service) AddCategory(actor *models.User, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return Error("分类名称不能为空")
	}
	if err := s.repo.CreateCategory(name); err != nil {
		return Error("分类已存在")
	}
	return s.log(actor.ID, actor.Username, "新增分类", name)
}
