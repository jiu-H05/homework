package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"lms/backend/internal/service"
)

// writeJSON 以 JSON 写出响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// errorBody 统一错误响应体。
type errorBody struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}

// decodeJSON 解析请求体 JSON。
func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(v)
}

// fail 根据错误类型写出响应：业务规则错误返回 400，其余返回 500。
func fail(w http.ResponseWriter, err error) {
	var se service.Error
	if errors.As(err, &se) {
		writeError(w, http.StatusBadRequest, se.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "服务器内部错误")
}
