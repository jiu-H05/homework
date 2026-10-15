package httpapi

import (
	"net/http"
	"strings"
	"time"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResp struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	Role         string `json:"role"`
	Username     string `json:"username"`
	RealName     string `json:"realName"`
	ReaderNo     string `json:"readerNo"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	// 登录限流：同 IP 失败次数累计，超出则 429。
	if !s.allowLogin(r) {
		writeError(w, http.StatusTooManyRequests, "尝试过于频繁，请 1 分钟后再试")
		return
	}
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	u, err := s.svc.Login(req.Username, req.Password)
	if err != nil {
		s.recordLoginFail(r)
		fail(w, err)
		return
	}
	access, refresh, err := issueTokens(claims{
		UID: u.ID, Username: u.Username, Role: u.Role, RealName: u.RealName,
	}, s.secret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "令牌签发失败")
		return
	}
	s.recordLoginOK(r)
	resp := loginResp{Token: access, RefreshToken: refresh, Role: u.Role, Username: u.Username, RealName: u.RealName}
	if u.Role == "reader" {
		if rd, err := s.svc.MyReader(u); err == nil {
			resp.ReaderNo = rd.ReaderNo
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// refreshReq 用 refresh token 换新 access token。
type refreshReq struct {
	RefreshToken string `json:"refreshToken"`
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	c, err := parseToken(req.RefreshToken, s.secret)
	if err != nil || c.Type != "refresh" {
		writeError(w, http.StatusUnauthorized, "刷新令牌无效或已过期")
		return
	}
	access, newRefresh, err := issueTokens(claims{
		UID: c.UID, Username: c.Username, Role: c.Role, RealName: c.RealName,
	}, s.secret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "令牌签发失败")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"token":        access,
		"refreshToken": newRefresh,
	})
}

type registerReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	no, err := s.svc.Register(req.Username, req.Password, req.RealName, req.Phone)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"readerNo": no})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	a := actor(r)
	out := map[string]any{
		"id": a.ID, "username": a.Username, "role": a.Role, "realName": a.RealName,
	}
	if a.Role == "reader" {
		if rd, err := s.svc.MyReader(a); err == nil {
			out["readerNo"] = rd.ReaderNo
			out["phone"] = rd.Phone
		}
	}
	writeJSON(w, http.StatusOK, out)
}

type passwordReq struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

func (s *Server) changePassword(w http.ResponseWriter, r *http.Request) {
	var req passwordReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.svc.ChangePassword(actor(r).ID, req.OldPassword, req.NewPassword); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "口令修改成功"})
}

// clientIP 取请求来源 IP（Cloudflare 隧道会在 X-Forwarded-For 里带真实 IP）。
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.Index(xff, ","); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i > 0 {
		host = host[:i]
	}
	return strings.TrimPrefix(host, "[")
}

var _ = time.Now // 保留 import
