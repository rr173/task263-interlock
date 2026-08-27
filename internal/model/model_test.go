package model

import "testing"

func TestVersionTransition(t *testing.T) {
	v := &InterlockingVersion{ID: "v1", Name: "x", State: VersionDraft}
	if err := v.Transition(VersionValidating); err != nil {
		t.Fatalf("draft→validating 应合法: %v", err)
	}
	if err := v.Transition(VersionHasConflict); err != nil {
		t.Fatalf("validating→has_conflict 应合法: %v", err)
	}
	if err := v.Transition(VersionValidating); err != nil {
		t.Fatalf("has_conflict→validating 应合法: %v", err)
	}
	if err := v.Transition(VersionReleasable); err != nil {
		t.Fatalf("validating→releasable 应合法: %v", err)
	}
	if err := v.Transition(VersionSealed); err != nil {
		t.Fatalf("releasable→sealed 应合法: %v", err)
	}
	// 封存后禁止任何迁移
	if err := v.Transition(VersionValidating); err == nil {
		t.Fatalf("sealed→validating 应非法")
	}
	// 草稿直接跳 releasable 非法
	d := &InterlockingVersion{ID: "v2", Name: "y", State: VersionDraft}
	if err := d.Transition(VersionReleasable); err == nil {
		t.Fatalf("draft→releasable 应非法")
	}
}

func TestSegmentLifecycle(t *testing.T) {
	s := &Segment{ID: "s1", Name: "seg", Kind: SegmentPlain, LengthM: 100, State: SegmentClear}
	if err := s.Reserve(); err != nil {
		t.Fatalf("clear→reserved 应合法: %v", err)
	}
	// 进路锁闭后列车进入：reserved→occupied 合法
	if err := s.Occupy(); err != nil {
		t.Fatalf("reserved→occupied 应合法: %v", err)
	}
	if err := s.Release(); err != nil {
		t.Fatalf("occupied→clear 应合法: %v", err)
	}
	if err := s.Free(); err == nil {
		t.Fatalf("clear 状态直接 free 应非法")
	}
	if err := s.Reserve(); err != nil {
		t.Fatalf("clear→reserved 应合法: %v", err)
	}
	if err := s.Free(); err != nil {
		t.Fatalf("reserved→clear(free) 应合法: %v", err)
	}
	if err := s.Occupy(); err != nil {
		t.Fatalf("clear→occupied 应合法: %v", err)
	}
	if err := s.Release(); err != nil {
		t.Fatalf("occupied→clear 应合法: %v", err)
	}
	// reserved 状态不能直接 release（必须先 free）
	s2 := &Segment{ID: "s2", Name: "seg2", Kind: SegmentPlain, LengthM: 100, State: SegmentClear}
	_ = s2.Reserve()
	if err := s2.Release(); err == nil {
		t.Fatalf("reserved 状态直接 release 应非法")
	}
}

func TestExceptionLifecycle(t *testing.T) {
	e := &Exception{ID: "e1", VersionID: "v1", ConflictID: "c1", State: ExceptionPending, Reason: "r"}
	if err := e.Approve(); err != nil {
		t.Fatalf("pending→approved 应合法: %v", err)
	}
	if err := e.Reject(); err == nil {
		t.Fatalf("approved→rejected 应非法")
	}
}

func TestSnapshotLifecycle(t *testing.T) {
	s := &ValidationSnapshot{ID: "s1", VersionID: "v1", Name: "n", State: SnapshotDraft}
	if err := s.Publish(); err != nil {
		t.Fatalf("draft→published 应合法: %v", err)
	}
	if err := s.Supersede("s2"); err != nil {
		t.Fatalf("published→superseded 应合法: %v", err)
	}
	if s.SupersededBy != "s2" {
		t.Fatalf("superseded_by 应为 s2")
	}
}
