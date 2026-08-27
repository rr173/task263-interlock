package service

import (
	"testing"

	"task263-interlock/internal/model"
)

func TestApprovedExceptionSuppressesConflict(t *testing.T) {
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
	ver, err := svc.Versions.Create("exc-suppress")
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
	var shared *model.Conflict
	for _, c := range conflicts {
		if c.Kind == model.ConflictSharedSegment {
			shared = c
			break
		}
	}
	if shared == nil {
		t.Fatal("expected shared segment conflict")
	}
	exc, err := svc.Exceptions.Create(ver.ID, shared.ID, "共享区段例外", "tester")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Exceptions.Approve(exc.ID); err != nil {
		t.Fatal(err)
	}
	c, err := svc.Conflicts.Get(shared.ID)
	if err != nil {
		t.Fatal(err)
	}
	if c.State != model.ConflictSuppressed {
		t.Fatalf("approved exception must suppress shared_segment conflict, got %s", c.State)
	}
}
