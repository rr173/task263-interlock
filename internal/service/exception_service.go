package service

import (
	"fmt"

	"task263-interlock/internal/model"
	"task263-interlock/internal/store"
)

// ExceptionService 例外裁决编排。
type ExceptionService struct {
	exceptions *store.ExceptionStore
	versions   *store.VersionStore
	conflicts  *store.ConflictStore
}

// NewExceptionService 创建例外服务。
func NewExceptionService(e *store.ExceptionStore, v *store.VersionStore, c *store.ConflictStore) *ExceptionService {
	return &ExceptionService{exceptions: e, versions: v, conflicts: c}
}

// Create 为某冲突创建待复核例外。
func (s *ExceptionService) Create(versionID, conflictID, reason, owner string) (*model.Exception, error) {
	v, err := s.versions.Get(versionID)
	if err != nil {
		return nil, err
	}
	if v.State == model.VersionSealed {
		return nil, model.ErrSealed
	}
	c, err := s.conflicts.Get(conflictID)
	if err != nil {
		return nil, err
	}
	if c.State == model.ConflictResolved || c.State == model.ConflictSuppressed {
		return nil, model.ErrConflictResolved
	}
	now := Now()
	e := &model.Exception{
		ID:         fmt.Sprintf("exc-%d", now.UnixMilli()),
		VersionID:  versionID,
		ConflictID: conflictID,
		State:      model.ExceptionPending,
		Reason:     reason,
		Owner:      owner,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.exceptions.Create(e); err != nil {
		return nil, err
	}
	if err := s.versions.AppendExceptionID(versionID, e.ID); err != nil {
		return nil, err
	}
	return e, nil
}

// Approve 批准例外并压制对应冲突。
func (s *ExceptionService) Approve(id string) (*model.Exception, error) {
	e, err := s.exceptions.Get(id)
	if err != nil {
		return nil, err
	}
	if err := e.Approve(); err != nil {
		return nil, err
	}
	if err := s.exceptions.UpdateState(id, e.State); err != nil {
		return nil, err
	}
	// 批准例外即压制对应冲突，与冲突类型无关（道岔竞争、共享区段、释放环、
	// 锁定阻断等均应同步为 suppressed）。UpdateState 在冲突不存在时返回 ErrNotFound。
	if err := s.conflicts.UpdateState(e.ConflictID, model.ConflictSuppressed); err != nil {
		return nil, err
	}
	return e, nil
}

// Reject 驳回例外。
func (s *ExceptionService) Reject(id string) (*model.Exception, error) {
	e, err := s.exceptions.Get(id)
	if err != nil {
		return nil, err
	}
	if err := e.Reject(); err != nil {
		return nil, err
	}
	if err := s.exceptions.UpdateState(id, e.State); err != nil {
		return nil, err
	}
	return e, nil
}

// ListByVersion 列出版本的例外。
func (s *ExceptionService) ListByVersion(versionID string) ([]*model.Exception, error) {
	return s.exceptions.ListByVersion(versionID)
}
