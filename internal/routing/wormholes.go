// internal/routing/wormholes.go
package routing

type WHLink struct {
	From int
	To   int
	Cost int
}

func (g *Graph) UpdateWormholes(links []WHLink) {
	g.dynamicAdj = make(map[int][]Edge)
	for _, l := range links {
		g.dynamicAdj[l.From] = append(g.dynamicAdj[l.From], Edge{To: l.To, Weight: l.Cost})
		g.dynamicAdj[l.To] = append(g.dynamicAdj[l.To], Edge{To: l.From, Weight: l.Cost})
	}
}
