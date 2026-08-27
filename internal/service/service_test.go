package service

import (
	"path/filepath"
	"testing"

	"task263-interlock/internal/model"
	"task263-interlock/internal/store"
)

func newTestServices(t *testing.T) *Services {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return New(db)
}

// TestFullLoop 覆盖完整闭环：拓扑 → 版本 → 冲突 → 例外 → 重新验证 → 快照。
func TestFullLoop(t *testing.T) {
	svc := newTestServices(t)

	for _, id := range []string{"seg-a", "seg-b", "seg-c"} {
		if _, err := svc.Segments.Create(&model.Segment{
			ID: id, Name: "seg" + id, Kind: model.SegmentPlain, LengthM: 100,
		}); err != nil {
			t.Fatalf("create segment %s: %v", id, err)
		}
	}
	if _, err := svc.Switches.Create(&model.Switch{
		ID: "sw-1", Name: "sw1", Position: model.SwitchNormal, NormalTo: "seg-b", ReverseTo: "seg-c",
	}); err != nil {
		t.Fatalf("create switch: %v", err)
	}

	r1, err := svc.Routes.Create(&model.Route{
		ID: "r1", Name: "直向", OriginSeg: "seg-a", DestSeg: "seg-b",
		PathSegs: []string{"seg-a", "seg-b"},
		Switches: []model.SwitchRequirement{{SwitchID: "sw-1", Position: model.SwitchNormal}},
	})
	if err != nil {
		t.Fatalf("create r1: %v", err)
	}
	r2, err := svc.Routes.Create(&model.Route{
		ID: "r2", Name: "侧向", OriginSeg: "seg-a", DestSeg: "seg-c",
		PathSegs: []string{"seg-a", "seg-c"},
		Switches: []model.SwitchRequirement{{SwitchID: "sw-1", Position: model.SwitchReverse}},
	})
	if err != nil {
		t.Fatalf("create r2: %v", err)
	}
	_ = r1
	_ = r2

	ver, err := svc.Versions.Create("v-1")
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	for _, rid := range []string{"r1", "r2"} {
		if _, err := svc.Versions.AttachRoute(ver.ID, rid); err != nil {
			t.Fatalf("attach %s: %v", rid, err)
		}
	}

	// 验证 → has_conflict
	conflicts, err := svc.ValidateVersion(ver.ID)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(conflicts) == 0 {
		t.Fatalf("应产生冲突")
	}
	ver2, _ := svc.Versions.Get(ver.ID)
	if ver2.State != model.VersionHasConflict {
		t.Fatalf("状态应为 has_conflict，实际 %s", ver2.State)
	}

	// 例外裁决：批准后冲突被压制
	exc, err := svc.Exceptions.Create(ver.ID, conflicts[0].ID, "受控例外", "tester")
	if err != nil {
		t.Fatalf("create exception: %v", err)
	}
	if _, err := svc.Exceptions.Approve(exc.ID); err != nil {
		t.Fatalf("approve exception: %v", err)
	}
	c, _ := svc.Conflicts.Get(conflicts[0].ID)
	if c.State != model.ConflictSuppressed {
		t.Fatalf("冲突应被压制，实际 %s", c.State)
	}

	// 重新验证后仍为 has_conflict（冲突存在但被压制）→ 需修订路径
	// 改为排除 r2 消除冲突
	if _, err := svc.Versions.ExcludeRoute(ver.ID, "r2"); err != nil {
		t.Fatalf("exclude r2: %v", err)
	}
	conflicts3, err := svc.ValidateVersion(ver.ID)
	if err != nil {
		t.Fatalf("revalidate: %v", err)
	}
	for _, c3 := range conflicts3 {
		if c3.Kind == model.ConflictSwitchContention {
			t.Fatalf("排除后不应再有道岔竞争: %s", c3.Detail)
		}
	}
	ver3, _ := svc.Versions.Get(ver.ID)
	if ver3.State != model.VersionReleasable {
		t.Fatalf("状态应为 releasable，实际 %s", ver3.State)
	}

	// 快照：创建 → 发布
	snap, err := svc.CreateSnapshot(ver.ID)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	pub, err := svc.Snapshots.Publish(snap.ID)
	if err != nil {
		t.Fatalf("publish snapshot: %v", err)
	}
	if pub.State != model.SnapshotPublished {
		t.Fatalf("快照应为 published，实际 %s", pub.State)
	}
	if pub.TopologyHash == "" {
		t.Fatalf("快照应记录拓扑哈希")
	}
	// 发布快照后版本封存，禁止修改
	if _, err := svc.Versions.ExcludeRoute(ver.ID, "r1"); err == nil {
		t.Fatalf("封存后排除进路应失败")
	}
}

// TestApproveSuppressesAllKinds 验证例外批准对所有冲突类型统一压制，
// 而非仅道岔竞争（回归：共享区段冲突批准后曾停留在 open）。
func TestApproveSuppressesAllKinds(t *testing.T) {
	svc := newTestServices(t)

	for _, id := range []string{"seg-a", "seg-b", "seg-c"} {
		if _, err := svc.Segments.Create(&model.Segment{
			ID: id, Name: "seg" + id, Kind: model.SegmentPlain, LengthM: 100,
		}); err != nil {
			t.Fatalf("create segment %s: %v", id, err)
		}
	}

	// 两条进路共享 seg-a，但使用各自独立的终点，不竞争任何道岔 → shared_segment 冲突。
	if _, err := svc.Routes.Create(&model.Route{
		ID: "r1", Name: "直向", OriginSeg: "seg-a", DestSeg: "seg-b",
		PathSegs: []string{"seg-a", "seg-b"},
	}); err != nil {
		t.Fatalf("create r1: %v", err)
	}
	if _, err := svc.Routes.Create(&model.Route{
		ID: "r2", Name: "侧向", OriginSeg: "seg-a", DestSeg: "seg-c",
		PathSegs: []string{"seg-a", "seg-c"},
	}); err != nil {
		t.Fatalf("create r2: %v", err)
	}

	ver, err := svc.Versions.Create("v-shared")
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	for _, rid := range []string{"r1", "r2"} {
		if _, err := svc.Versions.AttachRoute(ver.ID, rid); err != nil {
			t.Fatalf("attach %s: %v", rid, err)
		}
	}

	conflicts, err := svc.ValidateVersion(ver.ID)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	var shared *model.Conflict
	for _, c := range conflicts {
		if c.Kind == model.ConflictSharedSegment {
			shared = c
			break
		}
	}
	if shared == nil {
		t.Fatalf("应产生共享区段冲突，实际冲突数=%d", len(conflicts))
	}
	if shared.State != model.ConflictOpen {
		t.Fatalf("新冲突应为 open，实际 %s", shared.State)
	}

	exc, err := svc.Exceptions.Create(ver.ID, shared.ID, "受控例外", "tester")
	if err != nil {
		t.Fatalf("create exception: %v", err)
	}
	if _, err := svc.Exceptions.Approve(exc.ID); err != nil {
		t.Fatalf("approve exception: %v", err)
	}

	got, _ := svc.Conflicts.Get(shared.ID)
	if got.State != model.ConflictSuppressed {
		t.Fatalf("共享区段冲突经批准应被压制，实际 %s", got.State)
	}
}

// TestPersistAndRecover 验证重启恢复：数据库重开后实体与状态仍在。
func TestPersistAndRecover(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recover.db")

	db1, err := store.Open(path)
	if err != nil {
		t.Fatalf("open db1: %v", err)
	}
	svc1 := New(db1)
	if _, err := svc1.Segments.Create(&model.Segment{
		ID: "s1", Name: "seg", Kind: model.SegmentPlain, LengthM: 50,
	}); err != nil {
		t.Fatalf("create: %v", err)
	}
	ver1, err := svc1.Versions.Create("recover-ver")
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	_ = db1.Close()

	db2, err := store.Open(path)
	if err != nil {
		t.Fatalf("open db2: %v", err)
	}
	defer db2.Close()
	svc2 := New(db2)
	seg, err := svc2.Segments.Get("s1")
	if err != nil {
		t.Fatalf("重启后读取区段: %v", err)
	}
	if seg.Name != "seg" {
		t.Fatalf("区段数据未恢复: %s", seg.Name)
	}
	ver2, err := svc2.Versions.Get(ver1.ID)
	if err != nil {
		t.Fatalf("重启后读取版本: %v", err)
	}
	if ver2.State != model.VersionDraft {
		t.Fatalf("版本状态未恢复: %s", ver2.State)
	}
}

// TestIdempotentAttach 重复挂载同一进路应幂等。
func TestIdempotentAttach(t *testing.T) {
	svc := newTestServices(t)
	if _, err := svc.Segments.Create(&model.Segment{
		ID: "s1", Name: "seg", Kind: model.SegmentPlain, LengthM: 50,
	}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := svc.Segments.Create(&model.Segment{
		ID: "s2", Name: "seg2", Kind: model.SegmentPlain, LengthM: 50,
	}); err != nil {
		t.Fatalf("create: %v", err)
	}
	rt, err := svc.Routes.Create(&model.Route{
		ID: "r1", Name: "route", OriginSeg: "s1", DestSeg: "s2",
		PathSegs: []string{"s1", "s2"},
	})
	if err != nil {
		t.Fatalf("create route: %v", err)
	}
	ver, err := svc.Versions.Create("v")
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	if _, err := svc.Versions.AttachRoute(ver.ID, rt.ID); err != nil {
		t.Fatalf("attach 1: %v", err)
	}
	v2, err := svc.Versions.AttachRoute(ver.ID, rt.ID)
	if err != nil {
		t.Fatalf("attach 2 应幂等: %v", err)
	}
	count := 0
	for _, rid := range v2.RouteIDs {
		if rid == rt.ID {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("重复挂载导致 RouteIDs 含 %d 个相同进路", count)
	}
}
