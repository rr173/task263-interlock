package service

import (
	"testing"

	"task263-interlock/internal/model"
)

func TestOpenConflictsKeepHasConflictState(t *testing.T) {
	svc := newTestServices(t)
	for _, id := range []string{"seg-a", "seg-b", "seg-c", "seg-d"} {
		if _, err := svc.Segments.Create(&model.Segment{
			ID: id, Name: id, Kind: model.SegmentPlain, LengthM: 100,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.Switches.Create(&model.Switch{
		ID: "sw-1", Name: "sw1", Position: model.SwitchNormal, NormalTo: "seg-b", ReverseTo: "seg-c",
	}); err != nil {
		t.Fatal(err)
	}
	r1, err := svc.Routes.Create(&model.Route{
		ID: "r1", Name: "直向", OriginSeg: "seg-a", DestSeg: "seg-b",
		PathSegs: []string{"seg-a", "seg-b"},
		Switches: []model.SwitchRequirement{{SwitchID: "sw-1", Position: model.SwitchNormal}},
	})
	if err != nil {
		t.Fatal(err)
	}
	r2, err := svc.Routes.Create(&model.Route{
		ID: "r2", Name: "侧向", OriginSeg: "seg-a", DestSeg: "seg-c",
		PathSegs: []string{"seg-a", "seg-c"},
		Switches: []model.SwitchRequirement{{SwitchID: "sw-1", Position: model.SwitchReverse}},
	})
	if err != nil {
		t.Fatal(err)
	}
	r3, err := svc.Routes.Create(&model.Route{
		ID: "r3", Name: "共享", OriginSeg: "seg-a", DestSeg: "seg-d",
		PathSegs: []string{"seg-a", "seg-d"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ver, err := svc.Versions.Create("partial-exc")
	if err != nil {
		t.Fatal(err)
	}
	for _, rid := range []string{r1.ID, r2.ID, r3.ID} {
		if _, err := svc.Versions.AttachRoute(ver.ID, rid); err != nil {
			t.Fatal(err)
		}
	}
	conflicts, err := svc.ValidateVersion(ver.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(conflicts) < 2 {
		t.Fatalf("expected multiple conflicts, got %d", len(conflicts))
	}
	exc, err := svc.Exceptions.Create(ver.ID, conflicts[0].ID, "部分例外", "tester")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Exceptions.Approve(exc.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ValidateVersion(ver.ID); err != nil {
		t.Fatal(err)
	}
	v, err := svc.Versions.Get(ver.ID)
	if err != nil {
		t.Fatal(err)
	}
	if v.State == model.VersionReleasable {
		t.Fatalf("open conflicts remain after partial exception, state must not be releasable")
	}
	openLeft := 0
	stored, err := svc.Conflicts.ListByVersion(ver.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range stored {
		if c.State == model.ConflictOpen {
			openLeft++
		}
	}
	if openLeft == 0 {
		t.Fatal("expected remaining open conflicts")
	}
}
