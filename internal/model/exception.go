package model

import "time"

// ExceptionState 例外裁决状态。
type ExceptionState string

const (
	// ExceptionPending 例外待复核。
	ExceptionPending ExceptionState = "pending"
	// ExceptionApproved 例外已批准，冲突被压制。
	ExceptionApproved ExceptionState = "approved"
	// ExceptionRejected 例外被驳回。
	ExceptionRejected ExceptionState = "rejected"
)

// Exception 工程师对某一冲突标记的受控例外。
type Exception struct {
	ID        string         `json:"id"`
	VersionID string         `json:"version_id"`
	ConflictID string        `json:"conflict_id"`
	State     ExceptionState `json:"state"`
	Reason    string         `json:"reason"`
	Owner     string         `json:"owner"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Valid 校验例外字段。
func (e *Exception) Valid() error {
	if e.ID == "" {
		return ErrEmptyID
	}
	if e.VersionID == "" {
		return ErrEmptyVersion
	}
	if e.ConflictID == "" {
		return ErrEmptyConflict
	}
	switch e.State {
	case ExceptionPending, ExceptionApproved, ExceptionRejected:
	default:
		return ErrBadExceptionState
	}
	if e.Reason == "" {
		return ErrEmptyReason
	}
	return nil
}

// Approve 批准例外。
func (e *Exception) Approve() error {
	if e.State != ExceptionPending {
		return ErrBadTransition
	}
	e.State = ExceptionApproved
	e.UpdatedAt = time.Now()
	return nil
}

// Reject 驳回例外。
func (e *Exception) Reject() error {
	if e.State != ExceptionPending {
		return ErrBadTransition
	}
	e.State = ExceptionRejected
	e.UpdatedAt = time.Now()
	return nil
}
