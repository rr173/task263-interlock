package topology

import (
	"testing"

	"task263-interlock/internal/model"
)

func TestHashDeterministic(t *testing.T) {
	segs := []*model.Segment{
		{ID: "s1", Name: "seg1", Kind: model.SegmentPlain, LengthM: 100},
		{ID: "s2", Name: "seg2", Kind: model.SegmentSwitchArea, LengthM: 150},
	}
	sws := []*model.Switch{
		{ID: "w1", Name: "sw1", Position: model.SwitchNormal, NormalTo: "s1", ReverseTo: "s2"},
	}
	routes := []*model.Route{
		{ID: "r1", Name: "r1", OriginSeg: "s1", DestSeg: "s2",
			PathSegs: []string{"s1", "s2"},
			Switches: []model.SwitchRequirement{{SwitchID: "w1", Position: model.SwitchNormal}}},
	}
	h1, err := Hash(segs, sws, routes)
	if err != nil {
		t.Fatalf("hash1: %v", err)
	}
	h2, err := Hash(segs, sws, routes)
	if err != nil {
		t.Fatalf("hash2: %v", err)
	}
	if h1 != h2 {
		t.Fatalf("相同拓扑哈希应一致: %s != %s", h1, h2)
	}
	if len(h1) != 64 {
		t.Fatalf("SHA256 哈希应为 64 字符，实际 %d", len(h1))
	}
	// 修改进路后哈希应变化
	routes[0].DestSeg = "s1"
	h3, _ := Hash(segs, sws, routes)
	if h3 == h1 {
		t.Fatalf("拓扑变化后哈希应不同")
	}
}

func TestBuilderReferenceCheck(t *testing.T) {
	b := NewBuilder()
	// 道岔引用了不存在的区段 → 构建失败
	b.WithSegments([]*model.Segment{{ID: "s1", Name: "seg", Kind: model.SegmentPlain, LengthM: 1}})
	b.WithSwitches([]*model.Switch{{ID: "w1", Name: "sw", Position: model.SwitchNormal, NormalTo: "s1", ReverseTo: "missing"}})
	if _, err := b.Build(); err == nil {
		t.Fatalf("道岔引用缺失区段应构建失败")
	}
}
