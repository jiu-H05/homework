// Package config 负责运行配置：监听地址、数据文件路径、令牌密钥等。
package config

import (
	"os"
	"path"
)

// Config 保存后端运行参数。
type Config struct {
	Addr       string // HTTP 监听地址
	DBPath     string // SQLite 数据库路径
	TokenSecret []byte // 登录令牌 HMAC 密钥
}

// appDir 返回可执行文件（或开发时项目）所在目录，保证数据库与程序同目录、便于便携。
func appDir() string {
	if exe, err := os.Executable(); err == nil {
		return path.Dir(exe)
	}
	return "."
}

// Load 从环境变量读取配置，未设置时使用默认值。
func Load() *Config {
	addr := os.Getenv("LMS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8765"
	}
	dbPath := os.Getenv("LMS_DB")
	if dbPath == "" {
		dbPath = path.Join(appDir(), "library_data.db")
	}
	secret := os.Getenv("LMS_SECRET")
	if secret == "" {
		secret = "lms-local-default-secret-change-me"
	}
	return &Config{
		Addr:        addr,
		DBPath:      dbPath,
		TokenSecret: []byte(seline)