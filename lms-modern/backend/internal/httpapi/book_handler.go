package httpapi

import (
	"net/http"
	"strconv"

	"lms/backend/internal/models"
)

func (s *Server) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := s.svc.ListCategories()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

func (s *Server) addCategory(w http.ResponseWriter, r *http.Request) {
	a, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.svc.AddCategory(a, body.Name); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"message": "分类新增成功"})
}

func (s *Server) listBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var catID *int64
	if c := q.Get("category"); c != "" {
		if id, err := strconv.ParseInt(c, 10, 64); err == nil {
			catID = &id
		}
	}
	books, err := s.svc.ListBooks(q.Get("keyword"), catID)
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (s *Server) createBook(w http.ResponseWriter, r *http.Request) {
	a, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var b models.Book
	if err := decodeJSON(r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if err := s.svc.AddBook(a, &b); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (s *Server) updateBook(w http.ResponseWriter, r *http.Request) {
	a, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID 非法")
		return
	}
	var b models.Book
	if err := decodeJSON(r, &b); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	b.ID = id
	if err := s.svc.UpdateBook(a, &b); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) deleteBook(w http.ResponseWriter, r *http.Request) {
	a, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID 非法")
		return
	}
	if err := s.svc.DeleteBook(a, id); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "图书已删除"})
}
