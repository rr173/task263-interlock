// Package httpapi 提供 HTTP 接口层，统一路由前缀 /api。
package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"task263-interlock/internal/service"
)

// API 持有服务聚合与路由。
type API struct {
	svc *service.Services
	mux *http.ServeMux
}

// New 创建 API 实例并注册全部路由。
func New(svc *service.Services) *API {
	a := &API{svc: svc, mux: http.NewServeMux()}
	a.register()
	return a
}

// Handler 返回 HTTP 处理器。
func (a *API) Handler() http.Handler {
	return a.mux
}

// register 注册全部路由。
func (a *API) register() {
	// 区段
	a.mux.HandleFunc("POST /api/segments", a.handleCreateSegment)
	a.mux.HandleFunc("GET /api/segments", a.handleListSegments)
	a.mux.HandleFunc("GET /api/segments/{id}", a.handleGetSegment)
	a.mux.HandleFunc("POST /api/segments/{id}/occupy", a.handleOccupySegment)
	a.mux.HandleFunc("POST /api/segments/{id}/release", a.handleReleaseSegment)
	a.mux.HandleFunc("DELETE /api/segments/{id}", a.handleDeleteSegment)

	// 道岔
	a.mux.HandleFunc("POST /api/switches", a.handleCreateSwitch)
	a.mux.HandleFunc("GET /api/switches", a.handleListSwitches)
	a.mux.HandleFunc("GET /api/switches/{id}", a.handleGetSwitch)
	a.mux.HandleFunc("PUT /api/switches/{id}/position", a.handleSetSwitchPosition)
	a.mux.HandleFunc("DELETE /api/switches/{id}", a.handleDeleteSwitch)

	// 进路
	a.mux.HandleFunc("POST /api/routes", a.handleCreateRoute)
	a.mux.HandleFunc("GET /api/routes", a.handleListRoutes)
	a.mux.HandleFunc("GET /api/routes/{id}", a.handleGetRoute)
	a.mux.HandleFunc("GET /api/routes/version/{version_id}", a.handleListRoutesByVersion)
	a.mux.HandleFunc("POST /api/routes/{id}/exclude", a.handleExcludeRoute)

	// 联锁版本
	a.mux.HandleFunc("POST /api/versions", a.handleCreateVersion)
	a.mux.HandleFunc("GET /api/versions", a.handleListVersions)
	a.mux.HandleFunc("GET /api/versions/{id}", a.handleGetVersion)
	a.mux.HandleFunc("POST /api/versions/{id}/routes/{route_id}", a.handleAttachRoute)
	a.mux.HandleFunc("DELETE /api/versions/{id}/routes/{route_id}", a.handleExcludeRouteFromVersion)
	a.mux.HandleFunc("POST /api/versions/{id}/validate", a.handleValidateVersion)
	a.mux.HandleFunc("POST /api/versions/{id}/seal", a.handleSealVersion)

	// 冲突
	a.mux.HandleFunc("GET /api/versions/{id}/conflicts", a.handleListConflicts)
	a.mux.HandleFunc("GET /api/conflicts/{id}", a.handleGetConflict)

	// 例外
	a.mux.HandleFunc("POST /api/exceptions", a.handleCreateException)
	a.mux.HandleFunc("GET /api/versions/{id}/exceptions", a.handleListExceptions)
	a.mux.HandleFunc("POST /api/exceptions/{id}/approve", a.handleApproveException)
	a.mux.HandleFunc("POST /api/exceptions/{id}/reject", a.handleRejectException)

	// 快照
	a.mux.HandleFunc("POST /api/versions/{id}/snapshots", a.handleCreateSnapshot)
	a.mux.HandleFunc("GET /api/snapshots", a.handleListSnapshots)
	a.mux.HandleFunc("GET /api/snapshots/{id}", a.handleGetSnapshot)
	a.mux.HandleFunc("POST /api/snapshots/{id}/publish", a.handlePublishSnapshot)
	a.mux.HandleFunc("POST /api/snapshots/{id}/supersede", a.handleSupersedeSnapshot)

	// 统计与健康
	a.mux.HandleFunc("GET /api/stats", a.handleStats)
	a.mux.HandleFunc("GET /api/health", a.handleHealth)
}

// writeJSON 输出 JSON 响应。
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

// writeErr 输出错误响应。
func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// decodeJSON 解析请求体。
func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
