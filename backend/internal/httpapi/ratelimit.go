package httpapi

import (
	"net/http"
	"sync"
	"time"
)

// loginLimiter 登录失败限流：同 IP 1 分钟窗口内最多 8 次失败，超过则锁 1 分钟。
type loginLimiter struct {
	mu    sync.Mutex
	fails map[string]*failWindow
}

type failWindow struct {
	count   int
	firstAt time.Time
	lockedAt time.Time
}

const (
	loginMaxFails   = 8
	loginWindow     = time.Minute
	loginLockPeriod = time.Minute
)

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{fails: map[string]*failWindow{}}
}

func (l *loginLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.fails[ip]
	if !ok {
		return true
	}
	// 锁定期内拒绝
	if !w.lockedAt.IsZero() && time.Since(w.lockedAt) < loginLockPeriod {
		return false
	}
	// 窗口过期则重置
	if time.Since(w.firstAt) > loginWindow {
		delete(l.fails, ip)
	}
	return true
}

func (l *loginLimiter) recordFail(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.fails[ip]
	if !ok || time.Since(w.firstAt) > loginWindow {
		l.fails[ip] = &failWindow{count: 1, firstAt: time.Now()}
		return
	}
	w.count++
	if w.count >= loginMaxFails {
		w.lockedAt = time.Now()
	}
}

func (l *loginLimiter) recordOK(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, ip)
}

// allowLogin / recordLoginFail / recordLoginOK 挂在 Server 上，方便 handler 调用。
func (s *Server) allowLogin(r *http.Request) bool {
	return s.limiter.allow(clientIP(r))
}
func (s *Server) recordLoginFail(r *http.Request) { s.limiter.recordFail(clientIP(r)) }
func (s *Server) recordLoginOK(r *http.Request)   { s.limiter.recordOK(clientIP(r)) }
