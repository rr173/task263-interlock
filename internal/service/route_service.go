package service

import (
	"task263-interlock/internal/model"
	"task263-interlock/internal/store"
)

// RouteService 进路业务编排。
type RouteService struct {
	store *store.RouteStore
}

// NewRouteService 创建进路服务。
func NewRouteService(s *store.RouteStore) *RouteService {
	return &RouteService{store: s}
}

// Create 创建进路（候选态）。
func (s *RouteService) Create(r *model.Route) (*model.Route, error) {
	now := Now()
	r.CreatedAt = now
	r.UpdatedAt = now
	if r.State == "" {
		r.State = model.RouteCandidate
	}
	if err := s.store.Create(r); err != nil {
		return nil, err
	}
	return r, nil
}

// Get 读取进路。
func (s *RouteService) Get(id string) (*model.Route, error) {
	return s.store.Get(id)
}

// List 列出全部进路。
func (s *RouteService) List() ([]*model.Route, error) {
	return s.store.List()
}

// ListByVersion 按版本列出进路。
func (s *RouteService) ListByVersion(versionID string) ([]*model.Route, error) {
	return s.store.ListByVersion(versionID)
}

// Exclude 排除进路（不再参与验证）。
func (s *RouteService) Exclude(id string) (*model.Route, error) {
	r, err := s.store.Get(id)
	if err != nil {
		return nil, err
	}
	if r.State == model.RouteExcluded {
		return r, nil
	}
	if err := s.store.UpdateState(id, model.RouteExcluded); err != nil {
		return nil, err
	}
	r.State = model.RouteExcluded
	return r, nil
}
