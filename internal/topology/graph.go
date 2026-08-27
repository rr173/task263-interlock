// Package topology 维护线路区段、道岔与进路的引用关系图。
//
// 拓扑层不直接接触数据库，只对内存中的实体集合做一致性校验，
// 包括：区段是否存在、道岔端点是否引用有效区段、进路路径区段是否
// 全部存在、进路道岔要求是否引用有效道岔。
package topology

import (
	"fmt"

	"task263-interlock/internal/model"
)

// Graph 一次验证所需的内存拓扑快照。
type Graph struct {
	Segments map[string]*model.Segment
	Switches map[string]*model.Switch
	Routes   map[string]*model.Route
}

// NewGraph 构造空拓扑图。
func NewGraph() *Graph {
	return &Graph{
		Segments: map[string]*model.Segment{},
		Switches: map[string]*model.Switch{},
		Routes:   map[string]*model.Route{},
	}
}

// AddSegment 加入区段，ID 冲突时报错。
func (g *Graph) AddSegment(s *model.Segment) error {
	if err := s.Valid(); err != nil {
		return err
	}
	if _, ok := g.Segments[s.ID]; ok {
		return model.ErrAlreadyExists
	}
	g.Segments[s.ID] = s
	return nil
}

// AddSwitch 加入道岔，并校验其端点区段必须已存在。
func (g *Graph) AddSwitch(s *model.Switch) error {
	if err := s.Valid(); err != nil {
		return err
	}
	if _, ok := g.Segments[s.NormalTo]; !ok {
		return fmt.Errorf("%w: normal_to=%s", model.ErrSegmentMissing, s.NormalTo)
	}
	if _, ok := g.Segments[s.ReverseTo]; !ok {
		return fmt.Errorf("%w: reverse_to=%s", model.ErrSegmentMissing, s.ReverseTo)
	}
	if _, ok := g.Switches[s.ID]; ok {
		return model.ErrAlreadyExists
	}
	g.Switches[s.ID] = s
	return nil
}

// AddRoute 加入进路，校验路径区段与道岔要求全部引用有效实体。
func (g *Graph) AddRoute(r *model.Route) error {
	if err := r.Valid(); err != nil {
		return err
	}
	for _, segID := range r.PathSegs {
		if _, ok := g.Segments[segID]; !ok {
			return fmt.Errorf("%w: %s", model.ErrSegmentMissing, segID)
		}
	}
	for _, sw := range r.Switches {
		if _, ok := g.Switches[sw.SwitchID]; !ok {
			return fmt.Errorf("%w: %s", model.ErrSwitchMissing, sw.SwitchID)
		}
	}
	for _, rc := range r.Release {
		for _, sid := range rc.SegmentIDs {
			if _, ok := g.Segments[sid]; !ok {
				return fmt.Errorf("%w: release=%s", model.ErrSegmentMissing, sid)
			}
		}
		for _, sp := range rc.SwitchPos {
			if _, ok := g.Switches[sp.SwitchID]; !ok {
				return fmt.Errorf("%w: release switch=%s", model.ErrSwitchMissing, sp.SwitchID)
			}
		}
	}
	if _, ok := g.Routes[r.ID]; ok {
		return model.ErrAlreadyExists
	}
	g.Routes[r.ID] = r
	return nil
}

// Segment 按 ID 取区段。
func (g *Graph) Segment(id string) (*model.Segment, error) {
	s, ok := g.Segments[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return s, nil
}

// Switch 按 ID 取道岔。
func (g *Graph) Switch(id string) (*model.Switch, error) {
	s, ok := g.Switches[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return s, nil
}

// Route 按 ID 进取路。
func (g *Graph) Route(id string) (*model.Route, error) {
	r, ok := g.Routes[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return r, nil
}
