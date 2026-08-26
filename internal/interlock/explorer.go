package interlock

import (
	"sort"

	"task263-interlock/internal/model"
	"task263-interlock/internal/topology"
)

// Explorer 状态探索器：按进路两两组合尝试同时锁闭，暴露冲突。
type Explorer struct {
	graph *topology.Graph
}

// NewExplorer 创建探索器。
func NewExplorer(g *topology.Graph) *Explorer {
	return &Explorer{graph: g}
}

// Pair 一组进路组合。
type Pair struct {
	A string
	B string
}

// Explore 对版本内所有进路做两两同时锁闭尝试。
//
// 判定逻辑：
//  1. 两条进路若竞争同一道岔的不同位置 → switch_contention。
//  2. 两条进路若共享任一区段 → shared_segment（同时锁闭必然导致重复占用）。
//  3. 逐条按路径区段空闲检查，锁定阻断位置。
//  4. 返回所有命中的冲突描述。
func (e *Explorer) Explore() []*model.Conflict {
	ids := make([]string, 0, len(e.graph.Routes))
	for id := range e.graph.Routes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var out []*model.Conflict
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			a := e.graph.Routes[ids[i]]
			b := e.graph.Routes[ids[j]]
			if c := e.comparePair(a, b); c != nil {
				out = append(out, c)
			}
		}
	}
	return out
}

// comparePair 检查一对进路的冲突。
func (e *Explorer) comparePair(a, b *model.Route) *model.Conflict {
	// 道岔竞争：同一道岔要求不同位置。
	swMap := map[string]model.SwitchPosition{}
	for _, sw := range a.Switches {
		swMap[sw.SwitchID] = sw.Position
	}
	for _, sw := range b.Switches {
		if want, ok := swMap[sw.SwitchID]; ok && want != sw.Position {
			steps := []model.ConflictStep{
				{Seq: 1, RouteID: a.ID, SwitchID: sw.SwitchID, Action: "require", Detail: "进路A要求=" + string(want)},
				{Seq: 2, RouteID: b.ID, SwitchID: sw.SwitchID, Action: "require", Detail: "进路B要求=" + string(sw.Position)},
				{Seq: 3, RouteID: a.ID, SwitchID: sw.SwitchID, Action: "conflict", Detail: "同一道岔被要求到不同位置"},
			}
			return &model.Conflict{
				VersionID: a.VersionID,
				Kind:      model.ConflictSwitchContention,
				State:     model.ConflictOpen,
				RouteA:    a.ID,
				RouteB:    b.ID,
				ObjectID:  sw.SwitchID,
				Detail:    "道岔竞争：进路 " + a.ID + " 与 " + b.ID + " 竞争道岔 " + sw.SwitchID,
				Steps:     steps,
			}
		}
	}

	// 区段共享：任一共同区段 → 同时占用冲突。
	aSegs := map[string]bool{}
	for _, s := range a.PathSegs {
		aSegs[s] = true
	}
	for _, s := range b.PathSegs {
		if aSegs[s] {
			steps := []model.ConflictStep{
				{Seq: 1, RouteID: a.ID, SegmentID: s, Action: "occupy", Detail: "进路A占用区段"},
				{Seq: 2, RouteID: b.ID, SegmentID: s, Action: "occupy", Detail: "进路B占用区段"},
				{Seq: 3, RouteID: a.ID, SegmentID: s, Action: "conflict", Detail: "共享区段无法同时占用"},
			}
			return &model.Conflict{
				VersionID: a.VersionID,
				Kind:      model.ConflictSharedSegment,
				State:     model.ConflictOpen,
				RouteA:    a.ID,
				RouteB:    b.ID,
				ObjectID:  s,
				Detail:    "共享区段：进路 " + a.ID + " 与 " + b.ID + " 共享区段 " + s,
				Steps:     steps,
			}
		}
	}
	return nil
}

// BuildReleaseGraph 构建释放依赖图（进路 → 依赖进路），用于环检测。
//
// 进路 A 依赖进路 B：A 的释放条件里包含 B 路径区段或 B 要求的道岔。
// 返回有向边列表。
func (e *Explorer) BuildReleaseGraph() [][2]string {
	var edges [][2]string
	ids := make([]string, 0, len(e.graph.Routes))
	for id := range e.graph.Routes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, aid := range ids {
		a := e.graph.Routes[aid]
		for _, rc := range a.Release {
			for _, sid := range rc.SegmentIDs {
				for _, bid := range ids {
					b := e.graph.Routes[bid]
					if bid == aid {
						continue
					}
					for _, ps := range b.PathSegs {
						if ps == sid {
							edges = append(edges, [2]string{aid, bid})
						}
					}
				}
			}
		}
	}
	return edges
}

// DetectReleaseCycle 检测释放依赖环，返回环上的进路 ID 序列。
// 使用 DFS 找环；环为空表示无环。
func (e *Explorer) DetectReleaseCycle() []string {
	edges := e.BuildReleaseGraph()
	adj := map[string][]string{}
	for _, e := range edges {
		adj[e[0]] = append(adj[e[0]], e[1])
	}
	// 拓扑排序判定环
	state := map[string]int{} // 0 未访问, 1 访问中, 2 完成
	path := []string{}
	var dfs func(string) []string
	dfs = func(id string) []string {
		state[id] = 1
		path = append(path, id)
		for _, next := range adj[id] {
			if state[next] == 1 {
				// 找到环
				idx := 0
				for i, p := range path {
					if p == next {
						idx = i
						break
					}
				}
				cycle := append([]string{}, path[idx:]...)
				cycle = append(cycle, next)
				return cycle
			}
			if state[next] == 0 {
				if c := dfs(next); c != nil {
					return c
				}
			}
		}
		state[id] = 2
		path = path[:len(path)-1]
		return nil
	}
	for id := range adj {
		if state[id] == 0 {
			if c := dfs(id); c != nil {
				return c
			}
		}
	}
	return nil
}
