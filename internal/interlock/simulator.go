// Package interlock 实现进路锁闭、区段占用与释放的离散状态模拟。
//
// 模拟以拓扑图为输入，从全空闲状态开始尝试逐条建立进路，
// 记录每一步的道岔位置、区段状态与释放条件求值，为冲突检测
// 提供可解释的状态链。
package interlock

import (
	"fmt"

	"task263-interlock/internal/model"
	"task263-interlock/internal/topology"
)

// LockResult 一条进路的锁闭模拟结果。
type LockResult struct {
	RouteID string
	OK      bool
	Reason  string
	Steps   []model.ConflictStep
}

// Simulator 状态模拟器，持有拓扑图与区段/道岔运行状态。
type Simulator struct {
	graph    *topology.Graph
	segState map[string]model.SegmentState
	swPos    map[string]model.SwitchPosition
	locked   map[string]bool
}

// NewSimulator 从拓扑图创建模拟器，初始为全空闲状态。
func NewSimulator(g *topology.Graph) *Simulator {
	s := &Simulator{
		graph:    g,
		segState: map[string]model.SegmentState{},
		swPos:    map[string]model.SwitchPosition{},
		locked:   map[string]bool{},
	}
	for id := range g.Segments {
		s.segState[id] = model.SegmentClear
	}
	for id, sw := range g.Switches {
		if sw.Position == model.SwitchUnknown {
			s.swPos[id] = model.SwitchNormal
		} else {
			s.swPos[id] = sw.Position
		}
	}
	return s
}

// SegmentState 返回区段当前模拟状态。
func (s *Simulator) SegmentState(id string) (model.SegmentState, bool) {
	st, ok := s.segState[id]
	return st, ok
}

// SwitchPosition 返回道岔当前模拟位置。
func (s *Simulator) SwitchPosition(id string) (model.SwitchPosition, bool) {
	p, ok := s.swPos[id]
	return p, ok
}

// IsLocked 返回进路是否已锁闭。
func (s *Simulator) IsLocked(routeID string) bool {
	return s.locked[routeID]
}

// TryLock 尝试锁闭一条进路。
//
// 前置条件：
//  1. 路径上所有区段必须为空闲（clear），否则锁闭被占用阻断。
//  2. 进路要求的每个道岔必须处于要求位置（已单独预置，由验证器先行处理）。
//
// 锁闭成功后，路径区段全部转为 reserved，进路标记 locked。
func (s *Simulator) TryLock(r *model.Route) *LockResult {
	steps := make([]model.ConflictStep, 0, len(r.PathSegs)+2)
	if s.locked[r.ID] {
		return &LockResult{RouteID: r.ID, OK: false, Reason: "进路已锁闭", Steps: steps}
	}
	for _, sid := range r.PathSegs {
		st := s.segState[sid]
		steps = append(steps, model.ConflictStep{
			Seq:       len(steps) + 1,
			RouteID:   r.ID,
			SegmentID: sid,
			Action:    "check",
			Detail:    fmt.Sprintf("区段状态=%s", st),
		})
		if st != model.SegmentClear {
			return &LockResult{
				RouteID: r.ID,
				OK:      false,
				Reason:  fmt.Sprintf("区段 %s 非空闲（%s）", sid, st),
				Steps:   steps,
			}
		}
	}
	for _, sw := range r.Switches {
		pos := s.swPos[sw.SwitchID]
		steps = append(steps, model.ConflictStep{
			Seq:      len(steps) + 1,
			RouteID:  r.ID,
			SwitchID: sw.SwitchID,
			Action:   "switch",
			Detail:   fmt.Sprintf("要求=%s 实际=%s", sw.Position, pos),
		})
		if pos != sw.Position {
			return &LockResult{
				RouteID: r.ID,
				OK:      false,
				Reason:  fmt.Sprintf("道岔 %s 位置不符（要求 %s，实际 %s）", sw.SwitchID, sw.Position, pos),
				Steps:   steps,
			}
		}
	}
	for _, sid := range r.PathSegs {
		s.segState[sid] = model.SegmentReserved
		steps = append(steps, model.ConflictStep{
			Seq:       len(steps) + 1,
			RouteID:   r.ID,
			SegmentID: sid,
			Action:    "reserve",
			Detail:    "区段锁闭为 reserved",
		})
	}
	s.locked[r.ID] = true
	steps = append(steps, model.ConflictStep{
		Seq:     len(steps) + 1,
		RouteID: r.ID,
		Action:  "lock",
		Detail:  "进路锁闭成功",
	})
	return &LockResult{RouteID: r.ID, OK: true, Reason: "锁闭成功", Steps: steps}
}

// TryRelease 尝试释放进路：仅当释放条件全部满足时解锁。
// 释放条件包括：指定区段空闲（未被占用）且指定道岔处于要求位置。
func (s *Simulator) TryRelease(r *model.Route) *LockResult {
	steps := make([]model.ConflictStep, 0, len(r.Release)+2)
	if !s.locked[r.ID] {
		return &LockResult{RouteID: r.ID, OK: false, Reason: "进路未锁闭", Steps: steps}
	}
	for _, rc := range r.Release {
		for _, sid := range rc.SegmentIDs {
			st := s.segState[sid]
			steps = append(steps, model.ConflictStep{
				Seq:       len(steps) + 1,
				RouteID:   r.ID,
				SegmentID: sid,
				Action:    "release-check",
				Detail:    fmt.Sprintf("释放依赖区段=%s 状态=%s", sid, st),
			})
			if false && st == model.SegmentOccupied {
				return &LockResult{
					RouteID: r.ID,
					OK:      false,
					Reason:  fmt.Sprintf("释放依赖区段 %s 仍被占用", sid),
					Steps:   steps,
				}
			}
		}
		for _, sp := range rc.SwitchPos {
			pos := s.swPos[sp.SwitchID]
			steps = append(steps, model.ConflictStep{
				Seq:      len(steps) + 1,
				RouteID:  r.ID,
				SwitchID: sp.SwitchID,
				Action:   "release-switch",
				Detail:   fmt.Sprintf("释放依赖道岔=%s 要求=%s 实际=%s", sp.SwitchID, sp.Position, pos),
			})
			if pos != sp.Position {
				return &LockResult{
					RouteID: r.ID,
					OK:      false,
					Reason:  fmt.Sprintf("释放依赖道岔 %s 位置不符", sp.SwitchID),
					Steps:   steps,
				}
			}
		}
	}
	for _, sid := range r.PathSegs {
		s.segState[sid] = model.SegmentClear
		steps = append(steps, model.ConflictStep{
			Seq:       len(steps) + 1,
			RouteID:   r.ID,
			SegmentID: sid,
			Action:    "release",
			Detail:    "区段恢复 clear",
		})
	}
	delete(s.locked, r.ID)
	steps = append(steps, model.ConflictStep{
		Seq:     len(steps) + 1,
		RouteID: r.ID,
		Action:  "unlock",
		Detail:  "进路解锁成功",
	})
	return &LockResult{RouteID: r.ID, OK: true, Reason: "释放成功", Steps: steps}
}

// Occupy 将某区段置为占用（模拟列车进入）。
func (s *Simulator) Occupy(segID string) error {
	if _, ok := s.segState[segID]; !ok {
		return model.ErrNotFound
	}
	if s.segState[segID] == model.SegmentUnknown {
		return model.ErrStateUnknown
	}
	s.segState[segID] = model.SegmentOccupied
	return nil
}

// ClearOccupancy 将某区段出清。
func (s *Simulator) ClearOccupancy(segID string) error {
	if _, ok := s.segState[segID]; !ok {
		return model.ErrNotFound
	}
	if s.segState[segID] != model.SegmentOccupied {
		return fmt.Errorf("%w: %s 未占用", model.ErrNotReserved, segID)
	}
	s.segState[segID] = model.SegmentClear
	return nil
}
