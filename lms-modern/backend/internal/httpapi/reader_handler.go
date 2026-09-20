package httpapi

import (
	"net/http"
	"strconv"
)

func (s *Server) listReaders(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	readers, err := s.svc.ListReaders(r.URL.Query().Get("keyword"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, readers)
}

type createReaderReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	RealName string `json:"realName"`
	Phone    string `json:"phone"`
}

func (s *Server) createReader(w http.ResponseWriter, r *http.Request) {
	a, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var req createReaderReq
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	no, err := s.svc.AddReader(a, req.Username, req.Password, req.RealName, req.Phone)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"readerNo": no})
}

func (s *Server) setReaderStatus(w http.ResponseWriter, r *http.Request) {
	a, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID 非法")
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.svc.SetReaderStatus(a, id, body.Status); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "状态已更新"})
}

func (s *Server) resetReaderPassword(w http.ResponseWriter, r *http.Request) {
	a, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID 非法")
		return
	}
	tmp, err := s.svc.ResetReaderPassword(a, id)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"tempPassword": tmp})
}
