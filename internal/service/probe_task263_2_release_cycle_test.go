package service

import (
	"testing"

	"task263-interlock/internal/model"
)

func TestReleaseCycleDetected(t *testing.T) {
	svc := newTestServices(t)
	for _, id := range []string{"seg-a", "seg-b", "seg-c", "seg-d"} {
		if _, err := svc.Segments.Create(&model.Segment{
			ID: id, Name: id, Kind: model.SegmentPlain, LengthM: 100,
		}); err != nil {
			t.Fatal(err)
		}
	}
	r1, err := svc.Routes.Create(&model.Route{
		ID: "r1", Name: "r1", OriginSeg: "seg-a", DestSeg: "seg-c",
		PathSegs: []string{"seg-a", "seg-b", "seg-c"},
		Release:  []model.ReleaseCondition{{SegmentIDs: []string{"seg-b"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	r2, err := svc.Routes.Create(&model.Route{
		ID: "r2", Name: "r2", OriginSeg: "seg-d", DestSeg: "seg-b",
		PathSegs: []string{"seg-d", "seg-b"},
		Release:  []model.ReleaseCondition{{SegmentIDs: []string{"seg-b"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ver, err := svc.Versions.Create("release-cycle")
	if err != nil {
		t.Fatal(err)
	}
	for _, rid := range []string{r1.ID, r2.ID} {
		if _, err := svc.Versions.AttachRoute(ver.ID, rid); err != nil {
			t.Fatal(err)
		}
	}
	conflicts, err := svc.ValidateVersion(ver.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range conflicts {
		if c.Kind == model.ConflictReleaseCycle {
			return
		}
	}
	t.Fatalf("release dependency cycle must be reported, got %d conflicts", len(conflicts))
}
