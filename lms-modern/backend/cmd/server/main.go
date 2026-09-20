// Package main 是图书管理系统后端服务入口。
package main

import (
	"log"
	"net/http"

	"lms/backend/internal/config"
	"lms/backend/internal/httpapi"
	"lms/backend/internal/service"
	"lms/backend/internal/store"
)

func main() {
	cfg := config.Load()

	st, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer st.Close()

	svc := service.New(st)
	srv := httpapi.New(svc, cfg.TokenSecret)

	log.Printf("图书管理系统后端已启动，监听 %s（数据库：%s）", cfg.Addr, cfg.DBPath)
	if err := http.ListenAndServe(cfg.Addr, srv.Handler()); err != nil {
		log.Fatalf("服务退出: %v", err)
	}
}
