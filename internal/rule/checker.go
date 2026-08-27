// Package rule 提供进路定义的前提规则校验与悬空依赖分析。
//
// 与 topology 的引用完整性校验不同，rule 关注业务规则：
//   - 道岔位置预置的可行性（进路要求的道岔必须能摆到要求位置）；
//   - 释放条件中引用的区段/道岔是否属于进路路径或路径依赖；
//   - 释放条件是否自相矛盾。
package rule

import (
	"fmt"

	"task263-interlock/internal/model"
	"task263-interlock/internal/topology"
)

// Checker 规则校验器。
type Checker struct {
	graph *topology.Graph
}

// NewChecker 创建校验器。
func NewChecker(g *topology.Graph) *Checker {
	return &Checker{graph: g}
}

// CheckRoute 校验单条进路的业务规则。
func (c *Checker) CheckRoute(r *model.Route) error {
	// 释放条件引用的区段必须存在（环检测由 explorer 汇总）。
	for _, rc := range r.Release {
		for _, sid := range rc.SegmentIDs {
			if _, ok := c.graph.Segments[sid]; !ok {
				return fmt.Errorf("%w: %s", model.ErrSegmentMissing, sid)
			}
			// 悬空依赖：释放条件引用了不在路径上的区段。
			if !contains(r.PathSegs, sid) {
				return fmt.Errorf("%w: 释放条件引用非路径区段 %s", model.ErrPreconditionGap, sid)
			}
		}
	}
	return nil
}

// PresetSwitches 返回进路要求的道岔位置预置方案。
// 若两条进路要求同一道岔的不同位置，返回错误。
func (c *Checker) PresetSwitches(routes []*model.Route) (map[string]model.SwitchPosition, error) {
	preset := map[string]model.SwitchPosition{}
	for _, r := range routes {
		for _, sw := range r.Switches {
			if want, ok := preset[sw.SwitchID]; ok && want != sw.Position {
				return nil, fmt.Errorf("%w: 道岔 %s 同时要求 %s 与 %s", model.ErrSwitchContention, sw.SwitchID, want, sw.Position)
			}
			preset[sw.SwitchID] = sw.Position
		}
	}
	return preset, nil
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
