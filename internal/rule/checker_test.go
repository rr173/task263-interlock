package rule

import (
	"errors"
	"testing"

	"task263-interlock/internal/model"
	"task263-interlock/internal/topology"
)

func testGraph(t *testing.T) *topology.Graph {
	t.Helper()
	g := topology.NewGraph()
	for _, s := range []*model.Segment{
		{ID: "s1", Name: "a", Kind: model.SegmentPlain, LengthM: 100, State: model.SegmentClear},
		{ID: "s2", Name: "b", Kind: model.SegmentPlain, LengthM: 100, State: model.SegmentClear},
		{ID: "s3", Name: "c", Kind: model.SegmentPlain, LengthM: 100, State: model.SegmentClear},
	} {
		if err := g.AddSegment(s); err != nil {
			t.Fatalf("segment: %v", err)
		}
	}
	if err := g.AddSwitch(&model.Switch{
		ID: "sw1", Name: "sw", Position: model.SwitchNormal, NormalTo: "s2", ReverseTo: "s3",
	}); err != nil {
		t.Fatalf("switch: %v", err)
	}
	return g
}

func TestCheckRouteRejectsPreconditionGap(t *testing.T) {
	g := testGraph(t)
	r := &model.Route{
		ID: "r1", Name: "r1", OriginSeg: "s1", DestSeg: "s2",
		PathSegs: []string{"s1", "s2"},
		Release:  []model.ReleaseCondition{{SegmentIDs: []string{"s3"}}},
	}
	if err := NewChecker(g).CheckRoute(r); !errors.Is(err, model.ErrPreconditionGap) {
		t.Fatalf("expected precondition gap, got %v", err)
	}
}

func TestPresetSwitchesDetectsContention(t *testing.T) {
	g := testGraph(t)
	routes := []*model.Route{
		{ID: "r1", PathSegs: []string{"s1", "s2"}, Switches: []model.SwitchRequirement{{SwitchID: "sw1", Position: model.SwitchNormal}}},
		{ID: "r2", PathSegs: []string{"s1", "s3"}, Switches: []model.SwitchRequirement{{SwitchID: "sw1", Position: model.SwitchReverse}}},
	}
	if _, err := NewChecker(g).PresetSwitches(routes); !errors.Is(err, model.ErrSwitchContention) {
		t.Fatalf("expected switch contention, got %v", err)
	}
}
