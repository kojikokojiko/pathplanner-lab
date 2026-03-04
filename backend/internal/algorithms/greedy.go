package algorithms

import (
	"container/heap"
	"time"
)

type greedyItem struct {
	point Point
	h     int
	index int
}

type greedyPQ []*greedyItem

func (pq greedyPQ) Len() int           { return len(pq) }
func (pq greedyPQ) Less(i, j int) bool { return pq[i].h < pq[j].h }
func (pq greedyPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *greedyPQ) Push(x interface{}) {
	item := x.(*greedyItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}
func (pq *greedyPQ) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[:n-1]
	return item
}

type Greedy struct{}

func (g *Greedy) Run(input AlgorithmInput) AlgorithmOutput {
	start := time.Now()
	grid := input.Grid
	src, dst := input.Start, input.Goal
	n := input.Params.Neighbors

	visited := make(map[string]bool)
	parent := make(map[string]Point)
	order := make(map[string]int)

	pq := &greedyPQ{{point: src, h: Manhattan(src, dst)}}
	heap.Init(pq)
	expanded := 0

	for pq.Len() > 0 {
		cur := heap.Pop(pq).(*greedyItem)
		key := cur.point.String()

		if visited[key] {
			continue
		}
		visited[key] = true
		order[key] = expanded
		expanded++

		if cur.point == dst {
			path := ReconstructPath(parent, src, dst)
			return AlgorithmOutput{
				Found:         true,
				Path:          path,
				ExpandedNodes: expanded,
				TimeMicros:    time.Since(start).Microseconds(),
				Turns:         CountTurns(path),
				ExploredOrder: order,
			}
		}

		for _, nb := range GetNeighbors(cur.point, grid, n) {
			if !visited[nb.String()] {
				parent[nb.String()] = cur.point
				heap.Push(pq, &greedyItem{point: nb, h: Manhattan(nb, dst)})
			}
		}
	}

	return AlgorithmOutput{
		Found:         false,
		ExpandedNodes: expanded,
		TimeMicros:    time.Since(start).Microseconds(),
		ExploredOrder: order,
	}
}
