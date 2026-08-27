package interlock

import (
	"task263-interlock/internal/model"
	"task263-interlock/internal/topology"
)

// Validator 编排一次完整的版本验证：
//  1. 构建拓扑图
//  2. 预置进路道岔要求（保证模拟可锁闭，除非拓扑本身冲突）
//  3. 两两探索冲突
//  4. 释放环检测
//  5. 逐条进路锁闭尝试，暴露锁定阻断
//
// 输出：冲突集合（含释放环与锁定阻断）。
type Validator struct {
	graph *topology.Graph
}

// NewValidator 创建验证器。
func NewValidator(g *topology.Graph) *Validator {
	return &Validator{graph: g}
}

// ValidateResult 一次验证的输出。
type ValidateResult struct {
	Conflicts []*model.Conflict
	Locked    []string // 可锁闭进路
	Blocked   []string // 被阻断进路
}

// Validate 执行完整验证，返回全部冲突。
func (v *Validator) Validate() *ValidateResult {
	res := &ValidateResult{}
	sim := NewSimulator(v.graph)
	exp := NewExplorer(v.graph)

	// 1. 两两探索（道岔竞争 + 区段共享）
	res.Conflicts = append(res.Conflicts, exp.Explore()...)

	// 2. 释放依赖环
	if cycle := exp.DetectReleaseCycle(); len(cycle) > 0 {
		steps := []model.ConflictStep{}
		for i, rid := range cycle {
			steps = append(steps, model.ConflictStep{
				Seq:     i + 1,
				RouteID: rid,
				Action:  "release-dep",
				Detail:  "释放依赖：进路 " + rid + " 依赖下一进路释放",
			})
		}
		res.Conflicts = append(res.Conflicts, &model.Conflict{
			VersionID: v.firstVersionID(),
			Kind:      model.ConflictReleaseCycle,
			State:     model.ConflictOpen,
			RouteA:    cycle[0],
			RouteB:    cycle[len(cycle)-1],
			Detail:    "释放依赖环：" + joinIDs(cycle),
			Steps:     steps,
		})
	}

	// 3. 逐条锁闭尝试（锁定阻断）
	ids := sortedRouteIDs(v.graph)
	for _, rid := range ids {
		r := v.graph.Routes[rid]
		res2 := sim.TryLock(r)
		if res2.OK {
			res.Locked = append(res.Locked, rid)
		} else {
			res.Blocked = append(res.Blocked, rid)
			// 锁定阻断转为 locking_block 冲突
			steps := res2.Steps
			res.Conflicts = append(res.Conflicts, &model.Conflict{
				VersionID: r.VersionID,
				Kind:      model.ConflictLockingBlock,
				State:     model.ConflictOpen,
				RouteA:    rid,
				Detail:    "锁闭阻断：" + res2.Reason,
				Steps:     steps,
			})
		}
	}
	return res
}

// firstVersionID 取任一进路的版本 ID（构造冲突用）。
func (v *Validator) firstVersionID() string {
	for _, r := range v.graph.Routes {
		return r.VersionID
	}
	return ""
}
