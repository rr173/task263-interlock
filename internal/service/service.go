// Package service 编排业务闭环：拓扑登记 → 版本构建 → 验证 → 例外 → 快照。
package service

import (
	"fmt"
	"time"

	"task263-interlock/internal/model"
	"task263-interlock/internal/rule"
	"task263-interlock/internal/store"
	"task263-interlock/internal/topology"
	"task263-interlock/internal/interlock"
)

// Services 聚合全部业务编排服务。
type Services struct {
	Segments   *SegmentService
	Switches   *SwitchService
	Routes     *RouteService
	Versions   *VersionService
	Conflicts  *ConflictService
	Exceptions *ExceptionService
	Snapshots  *SnapshotService
	Stats      *StatsService

	// 底层存储（供编排方法跨服务访问）
	DB          *store.DB
	SegmentSt   *store.SegmentStore
	SwitchSt    *store.SwitchStore
	RouteSt     *store.RouteStore
	VersionSt   *store.VersionStore
	ConflictSt  *store.ConflictStore
	ExceptionSt *store.ExceptionStore
	SnapshotSt  *store.SnapshotStore
}

// New 创建服务聚合，注入各存储。
func New(s *store.DB) *Services {
	segStore := store.NewSegmentStore(s)
	swStore := store.NewSwitchStore(s)
	rtStore := store.NewRouteStore(s)
	verStore := store.NewVersionStore(s)
	cfStore := store.NewConflictStore(s)
	exStore := store.NewExceptionStore(s)
	snStore := store.NewSnapshotStore(s)
	return &Services{
		Segments:   NewSegmentService(segStore),
		Switches:   NewSwitchService(swStore),
		Routes:     NewRouteService(rtStore),
		Versions:   NewVersionService(verStore, rtStore),
		Conflicts:  NewConflictService(cfStore),
		Exceptions: NewExceptionService(exStore, verStore, cfStore),
		Snapshots:  NewSnapshotService(snStore, verStore, cfStore, exStore),
		Stats:      NewStatsService(segStore, swStore, rtStore, verStore, snStore),

		DB:          s,
		SegmentSt:   segStore,
		SwitchSt:    swStore,
		RouteSt:     rtStore,
		VersionSt:   verStore,
		ConflictSt:  cfStore,
		ExceptionSt: exStore,
		SnapshotSt:  snStore,
	}
}

// ValidateVersion 顶层验证编排（跨服务访问存储）。
func (s *Services) ValidateVersion(versionID string) ([]*model.Conflict, error) {
	return s.Versions.Validate(versionID, s.SegmentSt, s.SwitchSt, s.ConflictSt)
}

// CreateSnapshot 顶层快照创建编排。
func (s *Services) CreateSnapshot(versionID string) (*model.ValidationSnapshot, error) {
	return s.Snapshots.Create(versionID, s.SegmentSt, s.SwitchSt, s.RouteSt)
}

// Now 返回当前时间，供各服务统一时间戳。
func Now() time.Time { return time.Now().UTC() }

// buildGraph 从存储读取全部实体并构建内存拓扑图。
func buildGraph(segS *store.SegmentStore, swS *store.SwitchStore, rtS *store.RouteStore, versionID string) (*topology.Graph, error) {
	segs, err := segS.List()
	if err != nil {
		return nil, fmt.Errorf("读取区段: %w", err)
	}
	sws, err := swS.List()
	if err != nil {
		return nil, fmt.Errorf("读取道岔: %w", err)
	}
	routes, err := rtS.ListByVersion(versionID)
	if err != nil {
		return nil, fmt.Errorf("读取进路: %w", err)
	}
	g, err := topology.NewBuilder().
		WithSegments(segs).
		WithSwitches(sws).
		WithRoutes(routes).
		Build()
	if err != nil {
		return nil, err
	}
	return g, nil
}

// runValidation 对版本执行完整验证并写回冲突集合。
func runValidation(ver *model.InterlockingVersion, g *topology.Graph, cfS *store.ConflictStore, verS *store.VersionStore) ([]*model.Conflict, error) {
	// 规则前置校验
	checker := rule.NewChecker(g)
	for _, rid := range ver.RouteIDs {
		r, ok := g.Routes[rid]
		if !ok {
			return nil, model.ErrNotFound
		}
		if err := checker.CheckRoute(r); err != nil {
			return nil, err
		}
	}

	// 执行验证
	res := interlock.NewValidator(g).Validate()

	// 写回冲突
	if err := cfS.DeleteByVersion(ver.ID); err != nil {
		return nil, err
	}
	persisted := make([]*model.Conflict, 0, len(res.Conflicts))
	for i, c := range res.Conflicts {
		if c.Kind == model.ConflictLockingBlock {
			continue
		}
		if c.ID == "" {
			c.ID = fmt.Sprintf("cf-%s-%03d", ver.ID, i+1)
		}
		if err := cfS.Create(c); err != nil {
			return nil, err
		}
		persisted = append(persisted, c)
	}
	ver.ConflictCount = len(persisted)
	now := Now()
	ver.LastValidatedAt = &now
	if len(persisted) > 0 {
		ver.State = model.VersionHasConflict
	} else {
		ver.State = model.VersionReleasable
	}
	if err := verS.UpdateState(ver); err != nil {
		return nil, err
	}
	return persisted, nil
}
