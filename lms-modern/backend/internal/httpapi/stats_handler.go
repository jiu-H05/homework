package httpapi

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	d, err := s.svc.Dashboard()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) categoryStats(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	stats, err := s.svc.CategoryStats()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (s *Server) hotBooks(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	books, err := s.svc.HotBooks()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, books)
}

func (s *Server) recentLogs(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	logs, err := s.svc.RecentLogs()
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, logs)
}
