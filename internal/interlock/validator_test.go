package interlock

import (
	"testing"

	"task263-interlock/internal/model"
	"task263-interlock/internal/topology"
)

func buildTestGraph(t *testing.T, routes ...*model.Route) *topology.Graph {
	t.Helper()
	g := topology.NewGraph()
	for _, seg := range []*model.Segment{
		{ID: "seg-a", Name: "a", Kind: model.SegmentPlain, LengthM: 100, State: model.SegmentClear},
		{ID: "seg-b", Name: "b", Kind: model.SegmentPlain, LengthM: 100, State: model.SegmentClear},
		{ID: "seg-c", Name: "c", Kind: model.SegmentPlain, LengthM: 100, State: model.SegmentClear},
		{ID: "seg-d", Name: "d", Kind: model.SegmentPlain, LengthM: 100, State: model.SegmentClear},
	} {
		if err := g.AddSegment(seg); err != nil {
			t.Fatalf("add segment: %v", err)
		}
	}
	for _, sw := range []*model.Switch{
		{ID: "sw-1", Name: "sw1", Position: model.SwitchNormal, NormalTo: "seg-b", ReverseTo: "seg-c"},
	} {
		if err := g.AddSwitch(sw); err != nil {
			t.Fatalf("add switch: %v", err)
		}
	}
	for _, r := range routes {
		if err := g.AddRoute(r); err != nil {
			t.Fatalf("add route %s: %v", r.ID, err)
		}
	}
	return g
}

func TestExploreSwitchContention(t *testing.T) {
	routes := []*model.Route{
		{ID: "r1", Name: "r1", VersionID: "v1", OriginSeg: "seg-a", DestSeg: "seg-b",
			PathSegs: []string{"seg-a", "seg-b"},
			Switches: []model.SwitchRequirement{{SwitchID: "sw-1", Position: model.SwitchNormal}}},
		{ID: "r2", Name: "r2", VersionID: "v1", OriginSeg: "seg-a", DestSeg: "seg-c",
			PathSegs: []string{"seg-a", "seg-c"},
			Switches: []model.SwitchRequirement{{SwitchID: "sw-1", Position: model.SwitchReverse}}},
	}
	g := buildTestGraph(t, routes...)
	res := NewValidator(g).Validate()
	found := false
	for _, c := range res.Conflicts {
		if c.Kind == model.ConflictSwitchContention {
			found = true
			if c.RouteA != "r1" || c.RouteB != "r2" {
				t.Fatalf("冲突进路应为 r1/r2，实际 %s/%s", c.RouteA, c.RouteB)
			}
		}
	}
	if !found {
		t.Fatalf("应检测到道岔竞争冲突")
	}
}

func TestExploreSharedSegment(t *testing.T) {
	routes := []*model.Route{
		{ID: "r1", Name: "r1", VersionID: "v1", OriginSeg: "seg-a", DestSeg: "seg-b",
			PathSegs: []string{"seg-a", "seg-b"}},
		{ID: "r2", Name: "r2", VersionID: "v1", OriginSeg: "seg-a", DestSeg: "seg-d",
			PathSegs: []string{"seg-a", "seg-d"}},
	}
	g := buildTestGraph(t, routes...)
	res := NewValidator(g).Validate()
	found := false
	for _, c := range res.Conflicts {
		if c.Kind == model.ConflictSharedSegment {
			found = true
			if c.ObjectID != "seg-a" {
				t.Fatalf("共享区段应为 seg-a，实际 %s", c.ObjectID)
			}
		}
	}
	if !found {
		t.Fatalf("应检测到共享区段冲突")
	}
}

func TestNoConflictIndependentRoutes(t *testing.T) {
	// 独立进路（不共享区段、不竞争道岔）应无冲突
	routes := []*model.Route{
		{ID: "r1", Name: "r1", VersionID: "v1", OriginSeg: "seg-a", DestSeg: "seg-b",
			PathSegs: []string{"seg-a", "seg-b"},
			Switches: []model.SwitchRequirement{{SwitchID: "sw-1", Position: model.SwitchNormal}},
			Release:  []model.ReleaseCondition{{SegmentIDs: []string{"seg-b"}}}},
		{ID: "r2", Name: "r2", VersionID: "v1", OriginSeg: "seg-c", DestSeg: "seg-d",
			PathSegs: []string{"seg-c", "seg-d"}},
	}
	g := buildTestGraph(t, routes...)
	res := NewValidator(g).Validate()
	for _, c := range res.Conflicts {
		t.Fatalf("独立进路不应有冲突: %s", c.Detail)
	}
}

func TestSimulatorLockRelease(t *testing.T) {
	r := &model.Route{ID: "r1", Name: "r1", VersionID: "v1", OriginSeg: "seg-a", DestSeg: "seg-b",
		PathSegs: []string{"seg-a", "seg-b"},
		Switches: []model.SwitchRequirement{{SwitchID: "sw-1", Position: model.SwitchNormal}},
		Release:  []model.ReleaseCondition{{SegmentIDs: []string{"seg-b"}}}}
	g := buildTestGraph(t, r)
	sim := NewSimulator(g)
	lr := sim.TryLock(r)
	if !lr.OK {
		t.Fatalf("锁闭应成功: %s", lr.Reason)
	}
	if !sim.IsLocked("r1") {
		t.Fatalf("进路应处于锁闭态")
	}
	if st, _ := sim.SegmentState("seg-a"); st != model.SegmentReserved {
		t.Fatalf("路径区段应保留，实际 %s", st)
	}
	// 占用 seg-b 后释放应失败
	if err := sim.Occupy("seg-b"); err != nil {
		t.Fatalf("占用失败: %v", err)
	}
	rr := sim.TryRelease(r)
	if rr.OK {
		t.Fatalf("释放依赖区段被占用，释放应失败")
	}
	if err := sim.ClearOccupancy("seg-b"); err != nil {
		t.Fatalf("出清失败: %v", err)
	}
	rr2 := sim.TryRelease(r)
	if !rr2.OK {
		t.Fatalf("依赖满足后释放应成功: %s", rr2.Reason)
	}
	if sim.IsLocked("r1") {
		t.Fatalf("进路应已解锁")
	}
}
