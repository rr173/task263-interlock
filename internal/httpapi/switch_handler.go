package httpapi

import (
	"net/http"

	"task263-interlock/internal/model"
)

// handleCreateSwitch 创建道岔。
func (a *API) handleCreateSwitch(w http.ResponseWriter, r *http.Request) {
	var sw model.Switch
	if err := decodeJSON(r, &sw); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}
	out, err := a.svc.Switches.Create(&sw)
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

// handleListSwitches 列出全部道岔。
func (a *API) handleListSwitches(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Switches.List()
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleGetSwitch 读取道岔。
func (a *API) handleGetSwitch(w http.ResponseWriter, r *http.Request) {
	out, err := a.svc.Switches.Get(r.PathValue("id"))
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleSetSwitchPosition 设置道岔位置。
func (a *API) handleSetSwitchPosition(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Position model.SwitchPosition `json:"position"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}
	out, err := a.svc.Switches.SetPosition(r.PathValue("id"), body.Position)
	if err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleDeleteSwitch 删除道岔。
func (a *API) handleDeleteSwitch(w http.ResponseWriter, r *http.Request) {
	if err := a.svc.Switches.Delete(r.PathValue("id")); err != nil {
		respond(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
