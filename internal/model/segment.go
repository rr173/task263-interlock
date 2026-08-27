package model

import "time"

// SegmentKind 描述线路区段的物理类型。
type SegmentKind string

const (
	// SegmentPlain 普通区间区段，无道岔分支。
	SegmentPlain SegmentKind = "plain"
	// SegmentSwitchArea 道岔区段，可能被多条进路共享。
	SegmentSwitchArea SegmentKind = "switch_area"
	// SegmentPlatform 站台区段，通常位于进路两端。
	SegmentPlatform SegmentKind = "platform"
)

// SegmentState 区段占用状态。
type SegmentState string

const (
	// SegmentClear 区段空闲，可被进路锁闭。
	SegmentClear SegmentState = "clear"
	// SegmentOccupied 区段被列车占用。
	SegmentOccupied SegmentState = "occupied"
	// SegmentReserved 区段被进路锁闭保留。
	SegmentReserved SegmentState = "reserved"
	// SegmentUnknown 区段状态未知（轨旁设备失联）。
	SegmentUnknown SegmentState = "unknown"
)

// Segment 铁路线路区段，进路锁闭与占用校验的最小单元。
type Segment struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Kind      SegmentKind  `json:"kind"`
	LengthM   float64      `json:"length_m"`
	State     SegmentState `json:"state"`
	LineName  string       `json:"line_name"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// Valid 校验区段字段合法性。
func (s *Segment) Valid() error {
	if s.ID == "" {
		return ErrEmptyID
	}
	if s.Name == "" {
		return ErrEmptyName
	}
	switch s.Kind {
	case SegmentPlain, SegmentSwitchArea, SegmentPlatform:
	default:
		return ErrBadSegmentKind
	}
	if s.LengthM <= 0 {
		return ErrBadLength
	}
	switch s.State {
	case SegmentClear, SegmentOccupied, SegmentReserved, SegmentUnknown:
	default:
		return ErrBadSegmentState
	}
	return nil
}

// Occupy 将区段置为占用。
func (s *Segment) Occupy() error {
	if s.State == SegmentUnknown {
		return ErrStateUnknown
	}
	s.State = SegmentOccupied
	s.UpdatedAt = time.Now()
	return nil
}

// Release 将区段置为空闲（列车出清）。
func (s *Segment) Release() error {
	if s.State == SegmentReserved {
		return ErrReservedRelease
	}
	s.State = SegmentClear
	s.UpdatedAt = time.Now()
	return nil
}

// Reserve 将区段锁闭保留，仅允许从空闲转入。
func (s *Segment) Reserve() error {
	if s.State != SegmentClear {
		return ErrNotClear
	}
	s.State = SegmentReserved
	s.UpdatedAt = time.Now()
	return nil
}

// Free 将区段从保留转回空闲（进路取消）。
func (s *Segment) Free() error {
	if s.State != SegmentReserved {
		return ErrNotReserved
	}
	s.State = SegmentClear
	s.UpdatedAt = time.Now()
	return nil
}
