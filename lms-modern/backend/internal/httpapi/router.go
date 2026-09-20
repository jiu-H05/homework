package httpapi

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lms/backend/internal/models"
	"lms/backend/internal/service"
)

// ctxKey 上下文键类型，避免键冲突。
type ctxKey string

const actorKey ctxKey = "actor"

// Server HTTP API 服务器。
type Server struct {
	svc    *service.Service
	secret []byte
}

// New 创建 API 服务器。
func New(svc *service.Service, secret []byte) *Server {
	return &Server{svc: svc, secret: secret}
}

// publicPaths 无需登录即可访问的接口。
var publicPaths = map[string]bool{
	"POST /api/auth/login":    true,
	"POST /api/auth/register": true,
}

// Handler 组装路由并套用中间件。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// 认证
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/auth/register", s.register)
	mux.HandleFunc("GET /api/auth/me", s.me)
	mux.HandleFunc("POST /api/auth/password", s.changePassword)

	// 分类与图书
	mux.HandleFunc("GET /api/categories", s.listCategories)
	mux.HandleFunc("POST /api/categories", s.addCategory)
	mux.HandleFunc("GET /api/books", s.listBooks)
	mux.HandleFunc("POST /api/books", s.createBook)
	mux.HandleFunc("PUT /api/books/{id}", s.updateBook)
	mux.HandleFunc("DELETE /api/books/{id}", s.deleteBook)

	// 读者管理（管理员）
	mux.HandleFunc("GET /api/readers", s.listReaders)
	mux.HandleFunc("POST /api/readers", s.createReader)
	mux.HandleFunc("PATCH /api/readers/{id}/status", s.setReaderStatus)
	mux.HandleFunc("POST /api/readers/{id}/reset-password", s.resetReaderPassword)

	// 借阅（管理员办理 / 全量查询）
	mux.HandleFunc("GET /api/borrows/active", s.listActiveAll)
	mux.HandleFunc("GET /api/borrows/history", s.listHistoryAll)
	mux.HandleFunc("POST /api/borrows", s.createBorrow)
	mux.HandleFunc("POST /api/borrows/{id}/return", s.returnBook)

	// 读者自助
	mux.HandleFunc("GET /api/me/reader", s.myReader)
	mux.HandleFunc("GET /api/me/borrows/active", s.myActive)
	mux.HandleFunc("GET /api/me/borrows/history", s.myHistory)
	mux.HandleFunc("POST /api/me/borrows", s.selfBorrow)
	mux.HandleFunc("POST /api/me/borrows/{id}/renew", s.selfRenew)

	// 统计与日志
	mux.HandleFunc("GET /api/stats/dashboard", s.dashboard)
	mux.HandleFunc("GET /api/stats/categories", s.categoryStats)
	mux.HandleFunc("GET /api/stats/hot-books", s.hotBooks)
	mux.HandleFunc("GET /api/logs", s.recentLogs)

	// 健康检查
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return s.recoverMW(s.logMW(s.corsMW(s.authMW(mux))))
}

// logMW 把每个到达的请求追加到临时日志文件，用于诊断 WebView 的连通性。
func (s *Server) logMW(next http.Handler) http.Handler {
	logPath := filepath.Join(os.TempDir(), "lms_http.log")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			_, _ = f.WriteString(time.Now().Format("15:04:05") + " " + r.Method + " " +
				r.URL.Path + " origin=" + r.Header.Get("Origin") + "\n")
			_ = f.Close()
		}
		next.ServeHTTP(w, r)
	})
}

// recoverMW 捕获处理器 panic，避免单次异常导致服务退出。
func (s *Server) recoverMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v %s", rec, r.URL.Path)
				writeError(w, http.StatusInternalServerError, "服务器内部错误")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// corsMW 允许 Tauri WebView（tauri.localhost / 本地源）跨域访问本地 API。
func (s *Server) corsMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" || strings.Contains(origin, "tauri") || strings.HasPrefix(origin, "http://localhost") ||
			strings.HasPrefix(origin, "http://127.0.0.1") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if origin == "" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authMW 校验 Bearer 令牌并将登录用户写入上下文；公开路径放行。
func (s *Server) authMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path
		if publicPaths[key] || strings.HasPrefix(r.URL.Path, "/api/health") {
			next.ServeHTTP(w, r)
			return
		}
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
		writeError(w, http.StatusUnauthorized, "未登录或登录已失效")
			return
		}
		c, err := parseToken(strings.TrimPrefix(h, "Bearer "), s.secret)
		if err != nil {
		writeError(w, http.StatusUnauthorized, "登录已失效，请重新登录")
			return
		}
		actor := &models.User{ID: c.UID, Username: c.Username, Role: c.Role, RealName: c.RealName}
		ctx := context.WithValue(r.Context(), actorKey, actor)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// actor 从请求上下文取出登录用户。
func actor(r *http.Request) *models.User {
	u, _ := r.Context().Value(actorKey).(*models.User)
	return u
}

// requireAdmin 校验管理员身份，不满足时写入 403 并返回 false。
func requireAdmin(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	a := actor(r)
	if a == nil || a.Role != "admin" {
		writeError(w, http.StatusForbidden, "需要管理员权限")
		return nil, false
	}
	return a, true
}
