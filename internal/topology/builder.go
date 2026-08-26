package topology

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"task263-interlock/internal/model"
)

// Builder 从存储层实体集合构建内存拓扑图。
type Builder struct {
	segments []*model.Segment
	switches []*model.Switch
	routes   []*model.Route
}

// NewBuilder 创建构建器。
func NewBuilder() *Builder {
	return &Builder{}
}

// WithSegments 注入区段集合。
func (b *Builder) WithSegments(segs []*model.Segment) *Builder {
	b.segments = segs
	return b
}

// WithSwitches 注入道岔集合。
func (b *Builder) WithSwitches(sws []*model.Switch) *Builder {
	b.switches = sws
	return b
}

// WithRoutes 注入进路集合。
func (b *Builder) WithRoutes(routes []*model.Route) *Builder {
	b.routes = routes
	return b
}

// Build 组装拓扑图，任一引用不完整都会返回错误。
func (b *Builder) Build() (*Graph, error) {
	g := NewGraph()
	for _, s := range b.segments {
		if err := g.AddSegment(s); err != nil {
			return nil, fmt.Errorf("区段 %s: %w", s.ID, err)
		}
	}
	for _, s := range b.switches {
		if err := g.AddSwitch(s); err != nil {
			return nil, fmt.Errorf("道岔 %s: %w", s.ID, err)
		}
	}
	for _, r := range b.routes {
		if err := g.AddRoute(r); err != nil {
			return nil, fmt.Errorf("进路 %s: %w", r.ID, err)
		}
	}
	return g, nil
}

// Hash 计算拓扑的稳定哈希，用于快照冻结。
// 排序保证确定性：同一拓扑集合任意装载顺序得到同一哈希。
func Hash(segs []*model.Segment, sws []*model.Switch, routes []*model.Route) (string, error) {
	var lines []string
	for _, s := range segs {
		lines = append(lines, fmt.Sprintf("seg|%s|%s|%s|%.2f", s.ID, s.Name, s.Kind, s.LengthM))
	}
	for _, s := range sws {
		lines = append(lines, fmt.Sprintf("sw|%s|%s|%s|%s|%s", s.ID, s.Name, s.Position, s.NormalTo, s.ReverseTo))
	}
	for _, r := range routes {
		parts := []string{"route", r.ID, r.Name, r.OriginSeg, r.DestSeg, strings.Join(r.PathSegs, ">")}
		for _, sw := range r.Switches {
			parts = append(parts, sw.SwitchID+"="+string(sw.Position))
		}
		for _, rc := range r.Release {
			parts = append(parts, "rel:"+strings.Join(rc.SegmentIDs, ",")+"|"+strings.Join(swReqStrs(rc.SwitchPos), ","))
		}
		lines = append(lines, strings.Join(parts, "|"))
	}
	sort.Strings(lines)
	h := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(h[:]), nil
}

func swReqStrs(reqs []model.SwitchRequirement) []string {
	out := make([]string, 0, len(reqs))
	for _, r := range reqs {
		out = append(out, r.SwitchID+"="+string(r.Position))
	}
	sort.Strings(out)
	return out
}
