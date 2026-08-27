package service

import (
	"task263-interlock/internal/model"
	"task263-interlock/internal/store"
)

// SegmentService 区段业务编排。
type SegmentService struct {
	store *store.SegmentStore
}

// NewSegmentService 创建区段服务。
func NewSegmentService(s *store.SegmentStore) *SegmentService {
	return &SegmentService{store: s}
}

// Create 创建区段。
func (s *SegmentService) Create(seg *model.Segment) (*model.Segment, error) {
	now := Now()
	seg.CreatedAt = now
	seg.UpdatedAt = now
	if seg.State == "" {
		seg.State = model.SegmentClear
	}
	if err := s.store.Create(seg); err != nil {
		return nil, err
	}
	return seg, nil
}

// Get 读取区段。
func (s *SegmentService) Get(id string) (*model.Segment, error) {
	return s.store.Get(id)
}

// List 列出全部区段。
func (s *SegmentService) List() ([]*model.Segment, error) {
	return s.store.List()
}

// Occupy 占用区段。
func (s *SegmentService) Occupy(id string) (*model.Segment, error) {
	seg, err := s.store.Get(id)
	if err != nil {
		return nil, err
	}
	if err := seg.Occupy(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateState(id, seg.State); err != nil {
		return nil, err
	}
	return seg, nil
}

// Release 出清区段。
func (s *SegmentService) Release(id string) (*model.Segment, error) {
	seg, err := s.store.Get(id)
	if err != nil {
		return nil, err
	}
	if err := seg.Release(); err != nil {
		return nil, err
	}
	if err := s.store.UpdateState(id, seg.State); err != nil {
		return nil, err
	}
	return seg, nil
}

// Delete 删除区段。
func (s *SegmentService) Delete(id string) error {
	return s.store.Delete(id)
}
