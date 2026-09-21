package service

import (
	"errors"
	"strings"
	"time"

	"lms/backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// Login 校验用户名与口令，返回登录用户。
func (s *Service) Login(username, password string) (*models.User, error) {
	u, err := s.repo.GetUserByUsername(strings.TrimSpace(username))
	if err != nil {
		return nil, Error("用户名或口令错误")
	}
	if u.Status == "disabled" {
		return nil, Error("账号已被停用，请联系管理员")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, Error("用户名或口令错误")
	}
	_ = s.log(u.ID, u.Username, "登录", "用户登录系统")
	return u, nil
}

// Register 读者自助注册，返回读者编号。
func (s *Service) Register(username, password, realName, phone string) (string, error) {
	username = strings.TrimSpace(username)
	realName = strings.TrimSpace(realName)
	if len(username) < 3 {
		return "", Error("用户名至少 3 个字符")
	}
	if len(password) < 6 {
		return "", Error("口令至少 6 位")
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
		return "", Error("注册失败：用户名可能已存在")
	}
	_ = s.log(0, username, "读者注册", "新读者 "+realName+" 注册，编号 "+no)
	return no, nil
}

// ChangePassword 修改当前用户口令。
func (s *Service) ChangePassword(userID int64, oldPassword, newPassword string) error {
	if len(newPassword) < 6 {
		return Error("新口令至少 6 位")
	}
	u, err := s.repo.GetUserByID(userID)
	if err != nil {
		return Error("用户不存在")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)) != nil {
		return Error("原口令错误")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(userID, string(hash)); err != nil {
		return err
	}
	return s.log(u.ID, u.Username, "修改口令", "用户修改登录口令")
}

// log 写入操作日志。
func (s *Service) log(userID int64, username, action, detail string) error {
	var uid *int64
	if userID != 0 {
		id := userID
		uid = &id
	}
	return s.repo.AddLog(uid, username, action, detail, time.Now().Format("2006-01-02 15:04:05"))
}
