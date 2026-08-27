package service

import (
	"testing"

	"task263-interlock/internal/interlock"
	"task263-interlock/internal/model"
)

func TestReleaseBlockedWhenDependencyOccupied(t *testing.T) {
	svc := newTestServices(t)
	for _, id := range []string{"seg-a", "seg-b"} {
		if _, err := svc.Segments.Create(&model.Segment{
			ID: id, Name: id, Kind: model.SegmentPlain, LengthM: 100,
		}); err != nil {
			t.Fatal(err)
		}
	}
	rt, err := svc.Routes.Create(&model.Route{
		ID: "r1", Name: "r1", OriginSeg: "seg-a", DestSeg: "seg-b",
		PathSegs: []string{"seg-a", "seg-b"},
		Release:  []model.ReleaseCondition{{SegmentIDs: []string{"seg-b"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ver, err := svc.Versions.Create("release-dep")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Versions.AttachRoute(ver.ID, rt.ID); err != nil {
		t.Fatal(err)
	}
	g, err := buildGraph(svc.SegmentSt, svc.SwitchSt, svc.RouteSt, ver.ID)
	if err != nil {
		t.Fatal(err)
	}
	route := g.Routes[rt.ID]
	sim := interlock.NewSimulator(g)
	if lr := sim.TryLock(route); !lr.OK {
		t.Fatalf("lock should succeed: %s", lr.Reason)
	}
	if err := sim.Occupy("seg-b"); err != nil {
		t.Fatal(err)
	}
	if rr := sim.TryRelease(route); rr.OK {
		t.Fatal("release must fail while dependency segment is occupied")
	}
}
