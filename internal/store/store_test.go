package store

import (
	"path/filepath"
	"testing"
	"time"

	"task263-interlock/internal/model"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "store.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestVersionStorePersistsRouteIDs(t *testing.T) {
	db := openTestDB(t)
	vs := NewVersionStore(db)
	rs := NewRouteStore(db)

	now := time.Now().UTC()
	if err := vs.Create(&model.InterlockingVersion{
		ID: "ver-1", Name: "batch", State: model.VersionDraft,
		CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("create version: %v", err)
	}
	ver, err := vs.Get("ver-1")
	if err != nil {
		t.Fatalf("get version: %v", err)
	}
	if err := rs.Create(&model.Route{
		ID: "r1", Name: "route", OriginSeg: "s1", DestSeg: "s2",
		PathSegs: []string{"s1", "s2"}, State: model.RouteCandidate,
	}); err != nil {
		t.Fatalf("create route: %v", err)
	}
	if err := rs.UpdateVersion("r1", ver.ID); err != nil {
		t.Fatalf("bind route: %v", err)
	}
	ver.RouteIDs = append(ver.RouteIDs, "r1")
	if err := vs.UpdateState(ver); err != nil {
		t.Fatalf("update version: %v", err)
	}

	reloaded, err := vs.Get("ver-1")
	if err != nil {
		t.Fatalf("reload version: %v", err)
	}
	if len(reloaded.RouteIDs) != 1 || reloaded.RouteIDs[0] != "r1" {
		t.Fatalf("route_ids not persisted: %#v", reloaded.RouteIDs)
	}
}

func TestConflictStoreReplaceByVersion(t *testing.T) {
	db := openTestDB(t)
	cs := NewConflictStore(db)
	now := time.Now().UTC()
	for _, id := range []string{"cf-1", "cf-2"} {
		if err := cs.Create(&model.Conflict{
			ID: id, VersionID: "ver-1", Kind: model.ConflictSharedSegment,
			State: model.ConflictOpen, RouteA: "r1", RouteB: "r2",
			CreatedAt: now, UpdatedAt: now,
		}); err != nil {
			t.Fatalf("create %s: %v", id, err)
		}
	}
	if err := cs.DeleteByVersion("ver-1"); err != nil {
		t.Fatalf("delete by version: %v", err)
	}
	left, err := cs.ListByVersion("ver-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(left) != 0 {
		t.Fatalf("expected empty conflict set, got %d", len(left))
	}
}
