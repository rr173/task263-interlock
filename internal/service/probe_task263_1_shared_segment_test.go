package service

import (
	"testing"

	"task263-interlock/internal/model"
)

func TestSharedSegmentConflictReported(t *testing.T) {
	svc := newTestServices(t)
	for _, id := range []string{"seg-a", "seg-b", "seg-d"} {
		if _, err := svc.Segments.Create(&model.Segment{
			ID: id, Name: id, Kind: model.SegmentPlain, LengthM: 100,
		}); err != nil {
			t.Fatal(err)
		}
	}
	r1, err := svc.Routes.Create(&model.Route{
		ID: "r1", Name: "经b", OriginSeg: "seg-a", DestSeg: "seg-b",
		PathSegs: []string{"seg-a", "seg-b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	r2, err := svc.Routes.Create(&model.Route{
		ID: "r2", Name: "经d", OriginSeg: "seg-a", DestSeg: "seg-d",
		PathSegs: []string{"seg-a", "seg-d"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ver, err := svc.Versions.Create("shared-seg")
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
		if c.Kind == model.ConflictSharedSegment && c.ObjectID == "seg-a" {
			return
		}
	}
	t.Fatalf("shared segment conflict on seg-a must be reported, got %d conflicts", len(conflicts))
}
