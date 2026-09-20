// Package models 定义领域实体与数据传输对象，不依赖任何具体存储实现。
package models

import "errors"

// ErrNotFound 仓储层"记录不存在"的统一哨兵错误（中性包，供各层共享）。
var ErrNotFound = errors.New("记录不存在")

// User 系统用户（管理员或读者）。
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	RealName     string `json:"realName"`
	Status       string `json:"status"`
	CreatedAt    string `json:"createdAt"`
}

// Reader 读者档案。
type Reader struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	ReaderNo  string `json:"readerNo"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	CreatedAt string `json:"createdAt"`
}

// ReaderWithUser 读者列表行（含登录账号信息）。
type ReaderWithUser struct {
	UserID    int64  `json:"userId"`
	ReaderNo  string `json:"readerNo"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Username  string `json:"username"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// Category 图书分类。
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Book 图书。
type Book struct {
	ID           int64   `json:"id"`
	ISBN         string  `json:"isbn"`
	Title        string  `json:"title"`
	Author       string  `json:"author"`
	Publisher    string  `json:"publisher"`
	CategoryID   *int64  `json:"categoryId"`
	CategoryName string  `json:"category"`
	Location     string `json:"location"`
	TotalQty     int     `json:"totalQty"`
	AvailableQty int     `json:"availableQty"`
	CreatedAt    string  `json:"createdAt"`
}

// BorrowRecord 借阅记录（列表场景含关联字段）。
type BorrowRecord struct {
	ID         int64   `json:"id"`
	ReaderID   int64   `json:"readerId"`
	BookID     int64   `json:"bookId"`
	BorrowDate string  `json:"borrowDate"`
	DueDate    string  `json:"dueDate"`
	ReturnDate string  `json:"returnDate"`
	RenewCount int     `json:"renewCount"`
	Status     string `json:"status"`
	Fine       float64 `json:"fine"`
	// 关联展示字段
	Title      string `json:"title"`
	Author     string `json:"author"`
	ReaderName string `json:"readerName"`
	ReaderNo   string `json:"readerNo"`
	State      string `json:"state"`
}

// OperationLog 操作日志。
type OperationLog struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"userId"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	Detail    string `json:"detail"`
	CreatedAt string `json:"createdAt"`
}

// Dashboard 控制台汇总指标。
type Dashboard struct {
	BookKinds  int `json:"bookKinds"`
	BookCopies int `json:"bookCopies"`
	Readers    int `json:"readers"`
	Borrowed   int `json:"borrowed"`
	Overdue    int `json:"overdue"`
}

// CategoryStat 分类馆藏统计。
type CategoryStat struct {
	Name   string `json:"name"`
	Kinds  int    `json:"kinds"`
	Copies int    `json:"copies"`
}

// HotBook 图书借阅热度。
type HotBook struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Times  int    `json:"times"`
}
