package httpapi

import (
	"net/http"
)

// handleStats 统计摘要。
func (a *API) handleStats(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Stats.Summarize()
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleHealth 健康检查。
func (a *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
