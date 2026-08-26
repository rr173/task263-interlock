package service

import (
	"encoding/json"
	"testing"

	"task263-interlock/internal/model"
)

func TestTopologyHashIncludesPathSegments(t *testing.T) {
	svc := newTestServices(t)
	for _, id := range []string{"seg-a", "seg-b", "seg-x", "seg-y"} {
		if _, err := svc.Segments.Create(&model.Segment{
			ID: id, Name: id, Kind: model.SegmentPlain, LengthM: 100,
		}); err != nil {
			t.Fatal(err)
		}
	}
	rt, err := svc.Routes.Create(&model.Route{
		ID: "r1", Name: "main", OriginSeg: "seg-a", DestSeg: "seg-y",
		PathSegs: []string{"seg-a", "seg-b", "seg-y"},
	})
	if err != nil {
		t.Fatal(err)
	}
	ver, err := svc.Versions.Create("hash-path")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Versions.AttachRoute(ver.ID, rt.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ValidateVersion(ver.ID); err != nil {
		t.Fatal(err)
	}
	snap1, err := svc.CreateSnapshot(ver.ID)
	if err != nil {
		t.Fatal(err)
	}
	hash1 := snap1.TopologyHash
	newPath, err := json.Marshal([]string{"seg-a", "seg-x", "seg-y"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`UPDATE routes SET path_segs=? WHERE id=?`, string(newPath), rt.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`DELETE FROM snapshots WHERE id=?`, snap1.ID); err != nil {
		t.Fatal(err)
	}
	snap2, err := svc.CreateSnapshot(ver.ID)
	if err != nil {
		t.Fatal(err)
	}
	if hash1 == snap2.TopologyHash {
		t.Fatalf("path change must change topology hash: %s", hash1)
	}
}
