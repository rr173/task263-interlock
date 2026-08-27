package service

import (
	"fmt"

	"task263-interlock/internal/model"
	"task263-interlock/internal/store"
)

// VersionService 联锁版本业务编排。
type VersionService struct {
	versions *store.VersionStore
	routes   *store.RouteStore
}

// NewVersionService 创建版本服务。
func NewVersionService(v *store.VersionStore, r *store.RouteStore) *VersionService {
	return &VersionService{versions: v, routes: r}
}

// Create 创建草稿版本。
func (s *VersionService) Create(name string) (*model.InterlockingVersion, error) {
	now := Now()
	v := &model.InterlockingVersion{
		ID:           fmt.Sprintf("ver-%d", now.UnixMilli()),
		Name:         name,
		State:        model.VersionDraft,
		SegmentIDs:   []string{},
		SwitchIDs:    []string{},
		RouteIDs:     []string{},
		ExceptionIDs: []string{},
		ConflictCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.versions.Create(v); err != nil {
		return nil, err
	}
	return v, nil
}

// Get 读取版本。
func (s *VersionService) Get(id string) (*model.InterlockingVersion, error) {
	return s.versions.Get(id)
}

// List 列出全部版本。
func (s *VersionService) List() ([]*model.InterlockingVersion, error) {
	return s.versions.List()
}

// AttachRoute 将进路加入版本（仅草稿态）。
func (s *VersionService) AttachRoute(versionID, routeID string) (*model.InterlockingVersion, error) {
	v, err := s.versions.Get(versionID)
	if err != nil {
		return nil, err
	}
	if v.State != model.VersionDraft {
		return nil, model.ErrNotDraft
	}
	r, err := s.routes.Get(routeID)
	if err != nil {
		return nil, err
	}
	// 持久化进路的版本归属（ListByVersion 依赖此字段）
	if err := s.routes.UpdateVersion(routeID, versionID); err != nil {
		return nil, err
	}
	r.VersionID = versionID
	for _, rid := range v.RouteIDs {
		if rid == routeID {
			return v, nil // 幂等
		}
	}
	v.RouteIDs = append(v.RouteIDs, routeID)
	if err := s.versions.UpdateState(v); err != nil {
		return nil, err
	}
	return v, nil
}

// persistRouteVersion 回写进路的版本归属（保留供将来扩展）。
func (s *VersionService) persistRouteVersion(r *model.Route) error {
	return s.routes.UpdateVersion(r.ID, r.VersionID)
}

// ExcludeRoute 将进路从版本移除（标记排除）。
func (s *VersionService) ExcludeRoute(versionID, routeID string) (*model.InterlockingVersion, error) {
	v, err := s.versions.Get(versionID)
	if err != nil {
		return nil, err
	}
	if v.State == model.VersionSealed {
		return nil, model.ErrSealed
	}
	out := v.RouteIDs[:0]
	for _, rid := range v.RouteIDs {
		if rid != routeID {
			out = append(out, rid)
		}
	}
	v.RouteIDs = out
	// 从版本摘除：标记排除并清空版本归属，ListByVersion 不再返回该进路
	if err := s.routes.UpdateState(routeID, model.RouteExcluded); err != nil {
		return nil, err
	}
	if err := s.routes.UpdateVersion(routeID, ""); err != nil {
		return nil, err
	}
	if err := s.versions.UpdateState(v); err != nil {
		return nil, err
	}
	return v, nil
}

// Validate 提交验证：draft → validating → (has_conflict | releasable)。
func (s *VersionService) Validate(versionID string, segS *store.SegmentStore, swS *store.SwitchStore, cfS *store.ConflictStore) ([]*model.Conflict, error) {
	v, err := s.versions.Get(versionID)
	if err != nil {
		return nil, err
	}
	if v.State != model.VersionDraft && v.State != model.VersionHasConflict {
		return nil, model.ErrBadTransition
	}
	if err := v.Transition(model.VersionValidating); err != nil {
		return nil, err
	}
	g, err := buildGraph(segS, swS, s.routes, versionID)
	if err != nil {
		return nil, err
	}
	conflicts, err := runValidation(v, g, cfS, s.versions)
	if err != nil {
		return nil, err
	}
	return conflicts, nil
}

// Seal 封存版本（releasable → sealed），封存后不可修改。
func (s *VersionService) Seal(versionID string) (*model.InterlockingVersion, error) {
	v, err := s.versions.Get(versionID)
	if err != nil {
		return nil, err
	}
	if err := v.Transition(model.VersionSealed); err != nil {
		return nil, err
	}
	if err := s.versions.UpdateState(v); err != nil {
		return nil, err
	}
	return v, nil
}
