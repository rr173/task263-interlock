package httpapi

import (
	"net/http"
)

// handleCreateVersion 创建联锁版本。
func (a *API) handleCreateVersion(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}
	out, err := a.svc.Versions.Create(body.Name)
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// handleListVersions 列出全部版本。
func (a *API) handleListVersions(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Versions.List()
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleGetVersion 读取版本。
func (a *API) handleGetVersion(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Versions.Get(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleAttachRoute 将进路加入版本。
func (a *API) handleAttachRoute(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Versions.AttachRoute(r.PathValue("id"), r.PathValue("route_id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleExcludeRouteFromVersion 从版本移除进路。
func (a *API) handleExcludeRouteFromVersion(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Versions.ExcludeRoute(r.PathValue("id"), r.PathValue("route_id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleValidateVersion 提交版本验证。
func (a *API) handleValidateVersion(w http.ResponseWriter, r *http.Request) {
	conflicts, err := a.svc.ValidateVersion(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"conflict_count": len(conflicts),
		"conflicts":      conflicts,
	})
}

// handleSealVersion 封存版本。
func (a *API) handleSealVersion(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Versions.Seal(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
