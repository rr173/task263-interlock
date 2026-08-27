package model

import "time"

// SwitchPosition 道岔位置。
type SwitchPosition string

const (
	// SwitchNormal 道岔位于定位（直向）。
	SwitchNormal SwitchPosition = "normal"
	// SwitchReverse 道岔位于反位（侧向）。
	SwitchReverse SwitchPosition = "reverse"
	// SwitchUnknown 道岔位置未知（未表示）。
	SwitchUnknown SwitchPosition = "unknown"
)

// Switch 道岔，进路竞争校验的核心对象。
type Switch struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Position  SwitchPosition `json:"position"`
	NormalTo  string         `json:"normal_to"`
	ReverseTo string         `json:"reverse_to"`
	LineName  string         `json:"line_name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// Valid 校验道岔字段。
func (s *Switch) Valid() error {
	if s.ID == "" {
		return ErrEmptyID
	}
	if s.Name == "" {
		return ErrEmptyName
	}
	switch s.Position {
	case SwitchNormal, SwitchReverse, SwitchUnknown:
	default:
		return ErrBadSwitchPosition
	}
	if s.NormalTo == "" || s.ReverseTo == "" {
		return ErrBadSwitchEnds
	}
	if s.NormalTo == s.ReverseTo {
		return ErrSwitchSelfLoop
	}
	return nil
}

// SetPosition 设置道岔表示位置。
func (s *Switch) SetPosition(p SwitchPosition) error {
	if p == SwitchUnknown {
		return ErrUnknownSwitchPos
	}
	s.Position = p
	s.UpdatedAt = time.Now()
	return nil
}

// Requires 判断某进路对该道岔是否要求指定位置。
func (s *Switch) Requires(pos SwitchPosition) bool {
	return s.Position == pos
}
