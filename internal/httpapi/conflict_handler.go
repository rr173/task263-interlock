package httpapi

import (
	"net/http"
)

// handleListConflicts 列出版本的冲突。
func (a *API) handleListConflicts(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Conflicts.ListByVersion(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleGetConflict 读取冲突详情（含状态链）。
func (a *API) handleGetConflict(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Conflicts.Get(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
