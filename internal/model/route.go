package model

import "time"

// RouteState 进路生命周期状态。
type RouteState string

const (
	// RouteCandidate 进路已定义但从未锁闭。
	RouteCandidate RouteState = "candidate"
	// RouteLocked 进路已建立并锁闭。
	RouteLocked RouteState = "locked"
	// RouteReleased 进路已正常解锁。
	RouteReleased RouteState = "released"
	// RouteConflict 进路在验证中发现冲突。
	RouteConflict RouteState = "conflict"
	// RouteExcluded 进路被排除出当前联锁版本。
	RouteExcluded RouteState = "excluded"
)

// SwitchRequirement 进路对某道岔的位置要求。
type SwitchRequirement struct {
	SwitchID string         `json:"switch_id"`
	Position SwitchPosition `json:"position"`
}

// ReleaseCondition 进路解锁条件：依赖区段空闲 + 依赖道岔位置。
type ReleaseCondition struct {
	SegmentIDs []string         `json:"segment_ids"`
	SwitchPos  []SwitchRequirement `json:"switch_pos"`
	Reason     string           `json:"reason"`
}

// Route 进路：起点区段到终点区段之间的锁闭路径声明。
type Route struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	OriginSeg   string             `json:"origin_seg"`
	DestSeg     string             `json:"dest_seg"`
	PathSegs    []string           `json:"path_segs"` // 经过区段（含起终点）
	Switches    []SwitchRequirement `json:"switches"`
	Release     []ReleaseCondition `json:"release"`
	State       RouteState         `json:"state"`
	VersionID   string             `json:"version_id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// Valid 校验进路定义。
func (r *Route) Valid() error {
	if r.ID == "" {
		return ErrEmptyID
	}
	if r.Name == "" {
		return ErrEmptyName
	}
	if r.OriginSeg == "" || r.DestSeg == "" {
		return ErrBadRouteEnds
	}
	if len(r.PathSegs) < 2 {
		return ErrBadPath
	}
	if r.OriginSeg != r.PathSegs[0] || r.DestSeg != r.PathSegs[len(r.PathSegs)-1] {
		return ErrBadPathEnds
	}
	seen := map[string]bool{}
	for _, p := range r.PathSegs {
		if seen[p] {
			return ErrSegmentSelfLoop
		}
		seen[p] = true
	}
	for _, sw := range r.Switches {
		if sw.SwitchID == "" {
			return ErrBadSwitchReq
		}
		if sw.Position != SwitchNormal && sw.Position != SwitchReverse {
			return ErrBadSwitchReq
		}
	}
	for _, rc := range r.Release {
		if len(rc.SegmentIDs) == 0 && len(rc.SwitchPos) == 0 {
			return ErrEmptyRelease
		}
	}
	return nil
}

// RequiresSwitch 返回进路对某道岔的要求（若未要求返回 false）。
func (r *Route) RequiresSwitch(sid string) (SwitchRequirement, bool) {
	for _, sw := range r.Switches {
		if sw.SwitchID == sid {
			return sw, true
		}
	}
	return SwitchRequirement{}, false
}
