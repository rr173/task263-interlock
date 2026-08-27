package model

import "time"

// SnapshotState 验证快照生命周期状态。
type SnapshotState string

const (
	// SnapshotDraft 快照草稿。
	SnapshotDraft SnapshotState = "draft"
	// SnapshotPublished 快照已发布（不可变）。
	SnapshotPublished SnapshotState = "published"
	// SnapshotSuperseded 快照已被新快照替代。
	SnapshotSuperseded SnapshotState = "superseded"
)

// ValidationSnapshot 发布后的不可变验证快照。
type ValidationSnapshot struct {
	ID          string        `json:"id"`
	VersionID   string        `json:"version_id"`
	Name        string        `json:"name"`
	State       SnapshotState `json:"state"`
	TopologyHash string       `json:"topology_hash"`  // 冻结拓扑版本哈希
	ConflictTotal int         `json:"conflict_total"` // 发布时剩余冲突数（批准例外除外）
	ExceptionCount int        `json:"exception_count"`
	PublishedAt *time.Time    `json:"published_at"`
	SupersededBy string       `json:"superseded_by,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// Valid 校验快照字段。
func (s *ValidationSnapshot) Valid() error {
	if s.ID == "" {
		return ErrEmptyID
	}
	if s.VersionID == "" {
		return ErrEmptyVersion
	}
	if s.Name == "" {
		return ErrEmptyName
	}
	switch s.State {
	case SnapshotDraft, SnapshotPublished, SnapshotSuperseded:
	default:
		return ErrBadSnapshotState
	}
	return nil
}

// Publish 发布快照。
func (s *ValidationSnapshot) Publish() error {
	if s.State != SnapshotDraft {
		return ErrBadTransition
	}
	s.State = SnapshotPublished
	now := time.Now()
	s.PublishedAt = &now
	s.UpdatedAt = now
	return nil
}

// Supersede 用新快照替代当前已发布快照。
func (s *ValidationSnapshot) Supersede(newID string) error {
	if s.State != SnapshotPublished {
		return ErrBadTransition
	}
	s.State = SnapshotSuperseded
	s.SupersededBy = newID
	s.UpdatedAt = time.Now()
	return nil
}
