package model

import "time"

// VersionState 联锁版本生命周期状态。
type VersionState string

const (
	// VersionDraft 草稿，可编辑拓扑与进路。
	VersionDraft VersionState = "draft"
	// VersionValidating 正在执行冲突验证。
	VersionValidating VersionState = "validating"
	// VersionHasConflict 验证发现冲突，等待修订或例外裁决。
	VersionHasConflict VersionState = "has_conflict"
	// VersionReleasable 验证通过，可发布快照。
	VersionReleasable VersionState = "releasable"
	// VersionSealed 版本已封存，禁止修改。
	VersionSealed VersionState = "sealed"
)

// InterlockingVersion 联锁版本：一组进路定义与道岔状态的快照式组合。
type InterlockingVersion struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	State          VersionState `json:"state"`
	SegmentIDs     []string     `json:"segment_ids"`
	SwitchIDs      []string     `json:"switch_ids"`
	RouteIDs       []string     `json:"route_ids"`
	ExceptionIDs   []string     `json:"exception_ids"`
	ConflictCount  int          `json:"conflict_count"`
	LastValidatedAt *time.Time  `json:"last_validated_at"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// Valid 校验版本基本字段。
func (v *InterlockingVersion) Valid() error {
	if v.ID == "" {
		return ErrEmptyID
	}
	if v.Name == "" {
		return ErrEmptyName
	}
	switch v.State {
	case VersionDraft, VersionValidating, VersionHasConflict, VersionReleasable, VersionSealed:
	default:
		return ErrBadVersionState
	}
	return nil
}

// CanTransition 判定状态机迁移是否合法。
func (v *InterlockingVersion) CanTransition(next VersionState) bool {
	switch v.State {
	case VersionDraft:
		return next == VersionValidating
	case VersionValidating:
		return next == VersionHasConflict || next == VersionReleasable
	case VersionHasConflict:
		return next == VersionValidating || next == VersionReleasable || next == VersionSealed
	case VersionReleasable:
		return next == VersionSealed
	case VersionSealed:
		return false
	}
	return false
}

// Transition 执行状态迁移，非法迁移返回 ErrBadTransition。
func (v *InterlockingVersion) Transition(next VersionState) error {
	if !v.CanTransition(next) {
		return ErrBadTransition
	}
	v.State = next
	v.UpdatedAt = time.Now()
	return nil
}
