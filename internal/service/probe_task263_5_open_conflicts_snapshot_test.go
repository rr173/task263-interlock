package service

import (
	"errors"
	"testing"

	"task263-interlock/internal/model"
)

func TestSnapshotRejectsOpenConflicts(t *testing.T) {
	svc := newTestServices(t)
	for _, id := range []string{"seg-a", "seg-b", "seg-c"} {
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
	ver, err := svc.Versions.Create("snap-gate")
	if err != nil {
		t.Fatal(err)
	}
	for _, rid := range []string{r1.ID, r2.ID} {
		if _, err := svc.Versions.AttachRoute(ver.ID, rid); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.ValidateVersion(ver.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateSnapshot(ver.ID); !errors.Is(err, model.ErrVersionNotValidated) {
		if err == nil {
			t.Fatal("snapshot must not be created while open conflicts remain")
		}
		t.Fatalf("expected ErrVersionNotValidated, got %v", err)
	}
}
