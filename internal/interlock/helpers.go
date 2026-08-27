package interlock

import (
	"sort"

	"task263-interlock/internal/topology"
)

// sortedRouteIDs 返回按字典序排列的进路 ID。
func sortedRouteIDs(g *topology.Graph) []string {
	ids := make([]string, 0, len(g.Routes))
	for id := range g.Routes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// joinIDs 将 ID 序列拼接为可读字符串。
func joinIDs(ids []string) string {
	out := ""
	for i, id := range ids {
		if i > 0 {
			out += "→"
		}
		out += id
	}
	return out
}
