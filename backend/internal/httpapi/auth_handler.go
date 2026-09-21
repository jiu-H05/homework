package httpapi

import (
	"net/http"
	"time"
)

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResp struct {
	Token    string `json:"token"`
	Role     string `json:"role"`
	Username string `json:"username"`
	RealName string `json:"realName"`
	ReaderNo string `json:"readerNo"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	u, err := s.svc.Login(req.Username, req.Password)
	if err != nil {
		fail(w, err)
		return
	}
	token, err := signToken(claims{
		UID: u.ID, Username: u.Username, Role: u.Role, RealName: u.RealName,
		Exp: time.Now().Add(24 * time.Hour).Unix(),
	}, s.secret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "令牌签发失败")
		return
	}
	resp := loginResp{Token: token, Role: u.Role, Username: u.Username, RealName: u.RealName}
	if u.Role == "reader" {
		if rd, err := s.svc.MyReader(u); err == nil {
			resp.ReaderNo = rd.ReaderNo
		}
	}
	writeJSON(w, http.StatusOK, resp)
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
