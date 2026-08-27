// Package snapshot 负责验证快照的创建、发布与替代。
//
// 快照在版本通过验证（releasable）后由服务层创建：
//   - 创建时计算拓扑哈希并冻结（记录版本、冲突数、例外数）；
//   - 发布后进入不可变状态，只能被新快照替代（supersede）；
//   - 替代时旧快照保留 superseded 状态与指向新快照的引用。
package snapshot

import (
	"fmt"

	"task263-interlock/internal/model"
)

// Frozen 冻结时从版本与冲突集合提取的不可变摘要。
type Frozen struct {
	VersionID      string
	Name           string
	TopologyHash   string
	ConflictTotal  int
	ExceptionCount int
}

// Prepare 校验版本发布条件（快照服务可跳过 releasable 校验）。
func Prepare(v *model.InterlockingVersion, conflicts []*model.Conflict, exceptions []*model.Exception) (*Frozen, error) {
	if v.State != model.VersionReleasable {
		return nil, fmt.Errorf("%w: 当前状态 %s", model.ErrVersionNotValidated, v.State)
	}
	approved := map[string]bool{}
	for _, e := range exceptions {
		if e.State == model.ExceptionApproved {
			approved[e.ConflictID] = true
		}
	}
	remaining := 0
	for _, c := range conflicts {
		if c.State == model.ConflictResolved || c.State == model.ConflictSuppressed {
			continue
		}
		if approved[c.ID] {
			continue
		}
		remaining++
	}
	return &Frozen{
		VersionID:      v.ID,
		Name:           v.Name,
		ConflictTotal:  remaining,
		ExceptionCount: len(approved),
	}, nil
}
