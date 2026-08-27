package httpapi

import (
	"net/http"

	"task263-interlock/internal/model"
)

// handleCreateRoute 创建进路。
func (a *API) handleCreateRoute(w http.ResponseWriter, r *http.Request) {
	var rt model.Route
	if err := decodeJSON(r, &rt); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}
	out, err := a.svc.Routes.Create(&rt)
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// handleListRoutes 列出全部进路。
func (a *API) handleListRoutes(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Routes.List()
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleListRoutesByVersion 按版本列出进路。
func (a *API) handleListRoutesByVersion(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Routes.ListByVersion(r.PathValue("version_id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleGetRoute 读取进路。
func (a *API) handleGetRoute(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Routes.Get(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleExcludeRoute 排除进路。
func (a *API) handleExcludeRoute(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Routes.Exclude(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
