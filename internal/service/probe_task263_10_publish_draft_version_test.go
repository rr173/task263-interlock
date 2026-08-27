package service

import (
	"errors"
	"testing"

	"task263-interlock/internal/model"
)

func TestPublishRejectsDraftVersionSnapshot(t *testing.T) {
	svc := newTestServices(t)
	if _, err := svc.Segments.Create(&model.Segment{
		ID: "seg-a", Name: "a", Kind: model.SegmentPlain, LengthM: 100,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Segments.Create(&model.Segment{
		ID: "seg-b", Name: "b", Kind: model.SegmentPlain, LengthM: 100,
	}); err != nil {
		t.Fatal(err)
	}
	rt, err := svc.Routes.Create(&model.Route{
		ID: "r1", Name: "r1", OriginSeg: "seg-a", DestSeg: "seg-b",
		PathSegs: []string{"seg-a", "seg-b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ver, err := svc.Versions.Create("draft-pub")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Versions.AttachRoute(ver.ID, rt.ID); err != nil {
		t.Fatal(err)
	}
	snap, err := svc.CreateSnapshot(ver.ID)
	if err != nil {
		return
	}
	if _, err := svc.Snapshots.Publish(snap.ID); err == nil {
		t.Fatal("publish must fail for draft/unvalidated version snapshot")
	}
	if _, err := svc.CreateSnapshot(ver.ID); errors.Is(err, model.ErrVersionNotValidated) {
		return
	}
	if err != nil {
		t.Fatalf("unexpected create snapshot error: %v", err)
	}
}
