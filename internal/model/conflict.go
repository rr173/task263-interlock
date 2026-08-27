package model

import "time"

// ConflictKind 冲突类型。
type ConflictKind string

const (
	// ConflictSwitchContention 两条进路竞争同一道岔的不同位置。
	ConflictSwitchContention ConflictKind = "switch_contention"
	// ConflictSharedSegment 两条进路共享同一区段，无法同时占用。
	ConflictSharedSegment ConflictKind = "shared_segment"
	// ConflictReleaseCycle 释放条件依赖构成环。
	ConflictReleaseCycle ConflictKind = "release_cycle"
	// ConflictLockingBlock 锁闭时依赖区段已被占用。
	ConflictLockingBlock ConflictKind = "locking_block"
	// ConflictPreconditionGap 释放前提不满足（悬空依赖）。
	ConflictPreconditionGap ConflictKind = "precondition_gap"
)

// ConflictState 冲突记录状态。
type ConflictState string

const (
	// ConflictOpen 冲突待处理。
	ConflictOpen ConflictState = "open"
	// ConflictAcknowledged 已确认，等待裁决。
	ConflictAcknowledged ConflictState = "acknowledged"
	// ConflictResolved 已解决（修订或例外）。
	ConflictResolved ConflictState = "resolved"
	// ConflictSuppressed 被例外裁决压制。
	ConflictSuppressed ConflictState = "suppressed"
)

// ConflictStep 冲突路径中的一个状态链节点。
type ConflictStep struct {
	Seq      int    `json:"seq"`
	RouteID  string `json:"route_id,omitempty"`
	SegmentID string `json:"segment_id,omitempty"`
	SwitchID string `json:"switch_id,omitempty"`
	Action   string `json:"action"`
	Detail   string `json:"detail"`
}

// Conflict 一次验证中发现的冲突及其解释路径。
type Conflict struct {
	ID        string        `json:"id"`
	VersionID string        `json:"version_id"`
	Kind      ConflictKind  `json:"kind"`
	State     ConflictState `json:"state"`
	RouteA    string        `json:"route_a"`
	RouteB    string        `json:"route_b,omitempty"`
	ObjectID  string        `json:"object_id,omitempty"` // 竞争的道岔或区段
	Detail    string        `json:"detail"`
	Steps     []ConflictStep `json:"steps"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// Valid 校验冲突字段。
func (c *Conflict) Valid() error {
	if c.ID == "" {
		return ErrEmptyID
	}
	if c.VersionID == "" {
		return ErrEmptyVersion
	}
	switch c.Kind {
	case ConflictSwitchContention, ConflictSharedSegment, ConflictReleaseCycle, ConflictLockingBlock, ConflictPreconditionGap:
	default:
		return ErrBadConflictKind
	}
	switch c.State {
	case ConflictOpen, ConflictAcknowledged, ConflictResolved, ConflictSuppressed:
	default:
		return ErrBadConflictState
	}
	if c.RouteA == "" {
		return ErrEmptyRoute
	}
	return nil
}
