package service

import (
	"fmt"

	"task263-interlock/internal/model"
	"task263-interlock/internal/snapshot"
	"task263-interlock/internal/store"
	"task263-interlock/internal/topology"
)

// SnapshotService 验证快照编排。
type SnapshotService struct {
	snapshots  *store.SnapshotStore
	versions   *store.VersionStore
	conflicts  *store.ConflictStore
	exceptions *store.ExceptionStore
}

// NewSnapshotService 创建快照服务。
func NewSnapshotService(s *store.SnapshotStore, v *store.VersionStore, c *store.ConflictStore, e *store.ExceptionStore) *SnapshotService {
	return &SnapshotService{snapshots: s, versions: v, conflicts: c, exceptions: e}
}

// Create 创建快照草稿：校验发布条件并计算拓扑哈希。
func (s *SnapshotService) Create(versionID string, segS *store.SegmentStore, swS *store.SwitchStore, rtS *store.RouteStore) (*model.ValidationSnapshot, error) {
	v, err := s.versions.Get(versionID)
	if err != nil {
		return nil, err
	}
	conflicts, err := s.conflicts.ListByVersion(versionID)
	if err != nil {
		return nil, err
	}
	exceptions, err := s.exceptions.ListByVersion(versionID)
	if err != nil {
		return nil, err
	}
	frozen, err := snapshot.Prepare(v, conflicts, exceptions)
	if err != nil {
		return nil, err
	}
	segs, err := segS.List()
	if err != nil {
		return nil, err
	}
	sws, err := swS.List()
	if err != nil {
		return nil, err
	}
	routes, err := rtS.ListByVersion(versionID)
	if err != nil {
		return nil, err
	}
	hash, err := topology.Hash(segs, sws, routes)
	if err != nil {
		return nil, err
	}
	now := Now()
	snap := &model.ValidationSnapshot{
		ID:            fmt.Sprintf("snap-%d", now.UnixMilli()),
		VersionID:     versionID,
		Name:          frozen.Name,
		State:         model.SnapshotDraft,
		TopologyHash:  hash,
		ConflictTotal: frozen.ConflictTotal,
		ExceptionCount: frozen.ExceptionCount,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.snapshots.Create(snap); err != nil {
		return nil, err
	}
	return snap, nil
}

// Publish 发布快照（draft → published），并封存版本。
func (s *SnapshotService) Publish(id string) (*model.ValidationSnapshot, error) {
	snap, err := s.snapshots.Get(id)
	if err != nil {
		return nil, err
	}
	if err := snap.Publish(); err != nil {
		return nil, err
	}
	if err := s.snapshots.UpdateState(snap); err != nil {
		return nil, err
	}
	// 发布快照即封存对应版本
	v, err := s.versions.Get(snap.VersionID)
	if err != nil {
		return nil, err
	}
	if v.State != model.VersionSealed {
		if err := v.Transition(model.VersionSealed); err != nil {
			return nil, err
		}
		if err := s.versions.UpdateState(v); err != nil {
			return nil, err
		}
	}
	return snap, nil
}

// Supersede 用新快照替代旧快照（published → superseded）。
func (s *SnapshotService) Supersede(oldID, newID string) (*model.ValidationSnapshot, error) {
	old, err := s.snapshots.Get(oldID)
	if err != nil {
		return nil, err
	}
	newSnap, err := s.snapshots.Get(newID)
	if err != nil {
		return nil, err
	}
	if newSnap.State != model.SnapshotPublished {
		return nil, model.ErrBadTransition
	}
	if err := old.Supersede(newID); err != nil {
		return nil, err
	}
	if err := s.snapshots.UpdateState(old); err != nil {
		return nil, err
	}
	return old, nil
}

// List 列出全部快照。
func (s *SnapshotService) List() ([]*model.ValidationSnapshot, error) {
	return s.snapshots.List()
}

// Get 读取快照。
func (s *SnapshotService) Get(id string) (*model.ValidationSnapshot, error) {
	return s.snapshots.Get(id)
}
