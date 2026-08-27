package snapshot

import (
	"errors"
	"testing"

	"task263-interlock/internal/model"
)

func openConflict(id string) *model.Conflict {
	return &model.Conflict{ID: id, State: model.ConflictOpen}
}

// TestPrepareRejectsOpenConflict 锁定：版本仍处于 has_conflict 且冲突列表中存在
// open 记录时，必须阻止快照创建（返回 ErrVersionNotValidated）。
func TestPrepareRejectsOpenConflict(t *testing.T) {
	v := &model.InterlockingVersion{ID: "v1", Name: "n", State: model.VersionHasConflict}
	_, err := Prepare(v, []*model.Conflict{openConflict("c1")}, nil)
	if !errors.Is(err, model.ErrVersionNotValidated) {
		t.Fatalf("存在 open 冲突应阻止快照创建，得到 %v", err)
	}
}

// TestPrepareRejectsAcknowledgedConflict acknowledged 同样视为未解决，必须阻止快照创建。
func TestPrepareRejectsAcknowledgedConflict(t *testing.T) {
	v := &model.InterlockingVersion{ID: "v1", Name: "n", State: model.VersionReleasable}
	ack := &model.Conflict{ID: "c1", State: model.ConflictAcknowledged}
	_, err := Prepare(v, []*model.Conflict{ack}, nil)
	if !errors.Is(err, model.ErrVersionNotValidated) {
		t.Fatalf("acknowledged 冲突应阻止快照创建，得到 %v", err)
	}
}

// TestPrepareRequiresReleasable 版本未达 releasable（如 has_conflict）即使无 open 冲突也应被拒。
func TestPrepareRequiresReleasable(t *testing.T) {
	for _, st := range []model.VersionState{
		model.VersionDraft,
		model.VersionValidating,
		model.VersionHasConflict,
		model.VersionSealed,
	} {
		v := &model.InterlockingVersion{ID: "v1", Name: "n", State: st}
		if _, err := Prepare(v, nil, nil); !errors.Is(err, model.ErrVersionNotValidated) {
			t.Fatalf("状态 %s 应阻止快照创建，得到 %v", st, err)
		}
	}
}

// TestPrepareAllowsReleasableWithoutConflicts releasable 且无剩余冲突时放行并冻结正确摘要。
func TestPrepareAllowsReleasableWithoutConflicts(t *testing.T) {
	v := &model.InterlockingVersion{ID: "v1", Name: "n", State: model.VersionReleasable}
	// resolved/suppressed 与被已批准例外覆盖的冲突不计入剩余
	resolved := &model.Conflict{ID: "c1", State: model.ConflictResolved}
	suppressed := &model.Conflict{ID: "c2", State: model.ConflictSuppressed}
	approved := []*model.Exception{
		{ID: "e1", ConflictID: "c3", State: model.ExceptionApproved},
	}
	covered := openConflict("c3") // open 但被已批准例外覆盖 → 不计入
	frozen, err := Prepare(v, []*model.Conflict{resolved, suppressed, covered}, approved)
	if err != nil {
		t.Fatalf("releasable 且无剩余冲突应放行，得到 %v", err)
	}
	if frozen.ConflictTotal != 0 {
		t.Fatalf("剩余冲突数应为 0，实际 %d", frozen.ConflictTotal)
	}
	if frozen.ExceptionCount != 1 {
		t.Fatalf("已批准例外数应为 1，实际 %d", frozen.ExceptionCount)
	}
}
