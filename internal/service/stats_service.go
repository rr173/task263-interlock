package service

import (
	"task263-interlock/internal/store"
)

// StatsService 汇总统计编排。
type StatsService struct {
	segments  *store.SegmentStore
	switches  *store.SwitchStore
	routes    *store.RouteStore
	versions  *store.VersionStore
	snapshots *store.SnapshotStore
}

// NewStatsService 创建统计服务。
func NewStatsService(seg *store.SegmentStore, sw *store.SwitchStore, rt *store.RouteStore, ver *store.VersionStore, snap *store.SnapshotStore) *StatsService {
	return &StatsService{
		segments:  seg,
		switches:  sw,
		routes:    rt,
		versions:  ver,
		snapshots: snap,
	}
}

// Stats 汇总统计视图。
type Stats struct {
	SegmentCount     int `json:"segment_count"`
	SwitchCount      int `json:"switch_count"`
	RouteCount       int `json:"route_count"`
	VersionCount     int `json:"version_count"`
	SnapshotCount    int `json:"snapshot_count"`
	PublishedSnapshot int `json:"published_snapshot"`
}

// Summarize 统计全库实体数量。
func (s *StatsService) Summarize() (*Stats, error) {
	segs, err := s.segments.List()
	if err != nil {
		return nil, err
	}
	sws, err := s.switches.List()
	if err != nil {
		return nil, err
	}
	routes, err := s.routes.List()
	if err != nil {
		return nil, err
	}
	versions, err := s.versions.List()
	if err != nil {
		return nil, err
	}
	snaps, err := s.snapshots.List()
	if err != nil {
		return nil, err
	}
	published := 0
	for _, sn := range snaps {
		if sn.State == "published" {
			published++
		}
	}
	return &Stats{
		SegmentCount:      len(segs),
		SwitchCount:       len(sws),
		RouteCount:        len(routes),
		VersionCount:      len(versions),
		SnapshotCount:     len(snaps),
		PublishedSnapshot: published,
	}, nil
}
