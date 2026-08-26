package service

import (
	"task263-interlock/internal/model"
	"task263-interlock/internal/store"
)

// SwitchService 道岔业务编排。
type SwitchService struct {
	store *store.SwitchStore
}

// NewSwitchService 创建道岔服务。
func NewSwitchService(s *store.SwitchStore) *SwitchService {
	return &SwitchService{store: s}
}

// Create 创建道岔。
func (s *SwitchService) Create(sw *model.Switch) (*model.Switch, error) {
	now := Now()
	sw.CreatedAt = now
	sw.UpdatedAt = now
	if sw.Position == "" {
		sw.Position = model.SwitchNormal
	}
	if err := s.store.Create(sw); err != nil {
		return nil, err
	}
	return sw, nil
}

// Get 读取道岔。
func (s *SwitchService) Get(id string) (*model.Switch, error) {
	return s.store.Get(id)
}

// List 列出全部道岔。
func (s *SwitchService) List() ([]*model.Switch, error) {
	return s.store.List()
}

// SetPosition 设置道岔位置。
func (s *SwitchService) SetPosition(id string, pos model.SwitchPosition) (*model.Switch, error) {
	if err := s.store.UpdatePosition(id, pos); err != nil {
		return nil, err
	}
	return s.store.Get(id)
}

// Delete 删除道岔。
func (s *SwitchService) Delete(id string) error {
	return s.store.Delete(id)
}
