// internal/routing/dijkstra.go
package routing

import "container/heap"

type item struct{ node, dist int }
type pqueue []item

func (pq pqueue) Len() int            { return len(pq) }
func (pq pqueue) Less(i, j int) bool  { return pq[i].dist < pq[j].dist }
func (pq pqueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *pqueue) Push(x interface{}) { *pq = append(*pq, x.(item)) }
func (pq *pqueue) Pop() interface{} {
	old := *pq
	n := len(old)
	it := old[n-1]
	*pq = old[:n-1]
	return it
}

func (g *Graph) neighbors(u int) []Edge {
	out := g.staticAdj[u]
	if ds, ok := g.dynamicAdj[u]; ok {
		out = append(out, ds...)
	}
	return out
}

func (g *Graph) ShortestPath(src, dst int) (dist map[int]int, prev map[int]int) {
	const INF = int(1e9)
	dist = make(map[int]int)
	prev = make(map[int]int)
	pq := &pqueue{}
	heap.Init(pq)

	// Seed
	dist[src] = 0
	heap.Push(pq, item{node: src, dist: 0})

	visited := make(map[int]bool)

	for pq.Len() > 0 {
		it := heap.Pop(pq).(item)
		u := it.node
		if visited[u] {
			continue
		}
		visited[u] = true
		if u == dst {
			break
		}
		for _, e := range g.neighbors(u) {
			nd := dist[u] + e.Weight
			d, ok := dist[e.To]
			if !ok {
				d = INF
			}
			if nd < d {
				dist[e.To] = nd
				prev[e.To] = u
				heap.Push(pq, item{node: e.To, dist: nd})
			}
		}
	}
	return
}
