package httpapi

import (
	"net/http"
)

// handleCreateSnapshot 创建快照草稿。
func (a *API) handleCreateSnapshot(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.CreateSnapshot(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// handleListSnapshots 列出全部快照。
func (a *API) handleListSnapshots(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Snapshots.List()
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleGetSnapshot 读取快照。
func (a *API) handleGetSnapshot(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Snapshots.Get(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handlePublishSnapshot 发布快照。
func (a *API) handlePublishSnapshot(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Snapshots.Publish(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleSupersedeSnapshot 用新快照替代旧快照。
func (a *API) handleSupersedeSnapshot(w http.ResponseWriter, r *http.Request) {
	var body struct {
		NewSnapshotID string `json:"new_snapshot_id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}
	out, err := a.svc.Snapshots.Supersede(r.PathValue("id"), body.NewSnapshotID)
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
