package httpapi

import (
	"net/http"
)

// handleCreateException 为冲突创建例外。
func (a *API) handleCreateException(w http.ResponseWriter, r *http.Request) {
	var body struct {
		VersionID  string `json:"version_id"`
		ConflictID string `json:"conflict_id"`
		Reason     string `json:"reason"`
		Owner      string `json:"owner"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}
	out, err := a.svc.Exceptions.Create(body.VersionID, body.ConflictID, body.Reason, body.Owner)
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// handleListExceptions 列出版本的例外。
func (a *API) handleListExceptions(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Exceptions.ListByVersion(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleApproveException 批准例外。
func (a *API) handleApproveException(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Exceptions.Approve(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleRejectException 驳回例外。
func (a *API) handleRejectException(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Exceptions.Reject(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
