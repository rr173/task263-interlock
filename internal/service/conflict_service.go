package service

import (
	"task263-interlock/internal/model"
	"task263-interlock/internal/store"
)

// ConflictService 冲突查询编排。
type ConflictService struct {
	store *store.ConflictStore
}

// NewConflictService 创建冲突服务。
func NewConflictService(s *store.ConflictStore) *ConflictService {
	return &ConflictService{store: s}
}

// ListByVersion 列出版本的全部冲突。
func (s *ConflictService) ListByVersion(versionID string) ([]*model.Conflict, error) {
	return s.store.ListByVersion(versionID)
}

// Get 读取冲突详情（含状态链）。
func (s *ConflictService) Get(id string) (*model.Conflict, error) {
	return s.store.Get(id)
}
