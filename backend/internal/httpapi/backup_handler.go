package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// exportBackup 管理员导出：把当前数据库作为 .db 文件下载。
func (s *Server) exportBackup(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	tmp, err := os.CreateTemp("", "lms-backup-*.db")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "无法创建临时备份")
		return
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	if err := s.svc.Backup(tmp.Name()); err != nil {
		writeError(w, http.StatusInternalServerError, "导出失败: "+err.Error())
		return
	}
	name := fmt.Sprintf("lms-backup-%s.db", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	http.ServeFile(w, r, tmp.Name())
}

// importBackup 管理员导入：上传一个备份 .db，校验后热替换当前数据库。
func (s *Server) importBackup(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	tmp, err := os.CreateTemp("", "lms-restore-*.db")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "无法接收上传文件")
		return
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, r.Body); err != nil {
		tmp.Close()
		writeError(w, http.StatusBadRequest, "读取上传失败")
		return
	}
	tmp.Close()

	if err := s.svc.Restore(tmp.Name()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "数据已恢复，建议刷新页面"})
}
