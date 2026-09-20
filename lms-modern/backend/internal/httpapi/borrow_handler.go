package httpapi

import (
	"net/http"
	"strconv"
)

func parseReaderID(r *http.Request) *int64 {
	if v := r.URL.Query().Get("readerId"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			return &id
		}
	}
	return nil
}

func (s *Server) listActiveAll(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	recs, err := s.svc.ActiveBorrows(parseReaderID(r))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recs)
}

func (s *Server) listHistoryAll(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	recs, err := s.svc.HistoryBorrows(parseReaderID(r))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recs)
}

func (s *Server) createBorrow(w http.ResponseWriter, r *http.Request) {
	a, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var body struct {
		ReaderID int64 `json:"readerId"`
		BookID   int64 `json:"bookId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.svc.Borrow(a, body.ReaderID, body.BookID); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "借阅成功"})
}

func (s *Server) returnBook(w http.ResponseWriter, r *http.Request) {
	a, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID 非法")
		return
	}
	fine, err := s.svc.Return(a, id)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"fine": fine, "message": "归还成功"})
}

// ---- 读者自助 ----

func (s *Server) myReader(w http.ResponseWriter, r *http.Request) {
	rd, err := s.svc.MyReader(actor(r))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rd)
}

func (s *Server) myActive(w http.ResponseWriter, r *http.Request) {
	rd, err := s.svc.MyReader(actor(r))
	if err != nil {
		fail(w, err)
		return
	}
	recs, err := s.svc.ActiveBorrows(&rd.ID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recs)
}

func (s *Server) myHistory(w http.ResponseWriter, r *http.Request) {
	rd, err := s.svc.MyReader(actor(r))
	if err != nil {
		fail(w, err)
		return
	}
	recs, err := s.svc.HistoryBorrows(&rd.ID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recs)
}

func (s *Server) selfBorrow(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BookID int64 `json:"bookId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.svc.SelfBorrow(actor(r), body.BookID); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "借阅成功"})
}

func (s *Server) selfRenew(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID 非法")
		return
	}
	if err := s.svc.SelfRenew(actor(r), id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "续借成功"})
}
