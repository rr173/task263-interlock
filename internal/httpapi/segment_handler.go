package httpapi

import (
	"net/http"

	"task263-interlock/internal/model"
)

// handleCreateSegment 创建区段。
func (a *API) handleCreateSegment(w http.ResponseWriter, r *http.Request) {
	var seg model.Segment
	if err := decodeJSON(r, &seg); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}
	out, err := a.svc.Segments.Create(&seg)
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// handleListSegments 列出全部区段。
func (a *API) handleListSegments(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Segments.List()
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleGetSegment 读取区段。
func (a *API) handleGetSegment(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Segments.Get(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleOccupySegment 占用区段。
func (a *API) handleOccupySegment(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Segments.Occupy(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleReleaseSegment 出清区段。
func (a *API) handleReleaseSegment(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Segments.Release(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleDeleteSegment 删除区段。
func (a *API) handleDeleteSegment(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.Segments.Delete(r.PathValue("id")); err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
