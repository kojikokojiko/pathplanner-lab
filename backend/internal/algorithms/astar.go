package algorithms

import (
	"container/heap"
	"time"
)

type astarItem struct {
	point Point
	g     float64
	f     float64
	index int
}

type astarPQ []*astarItem

func (pq astarPQ) Len() int           { return len(pq) }
func (pq astarPQ) Less(i, j int) bool { return pq[i].f < pq[j].f }
func (pq astarPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *astarPQ) Push(x interface{}) {
	item := x.(*astarItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}
func (pq *astarPQ) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[:n-1]
	return item
}

type AStar struct{}

func (a *AStar) Run(input AlgorithmInput) AlgorithmOutput {
	return runAStarFamily(input, 1.0)
}

func runAStarFamily(input AlgorithmInput, w float64) AlgorithmOutput {
	start := time.Now()
	grid := input.Grid
	src, dst := input.Start, input.Goal
	n := input.Params.Neighbors

	gCost := make(map[string]float64)
	parent := make(map[string]Point)
	order := make(map[string]int)
	gCost[src.String()] = 0

	h := float64(Manhattan(src, dst))
	pq := &astarPQ{{point: src, g: 0, f: w * h}}
	heap.Init(pq)
	expanded := 0

	for pq.Len() > 0 {
		cur := heap.Pop(pq).(*astarItem)
		key := cur.point.String()

		if _, visited := order[key]; visited {
			continue
		}
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
			newG := cur.g + float64(grid[nb.Y][nb.X].Cost)
			nbKey := nb.String()
			if old, ok := gCost[nbKey]; !ok || newG < old {
				gCost[nbKey] = newG
				parent[nbKey] = cur.point
				hNb := float64(Manhattan(nb, dst))
				heap.Push(pq, &astarItem{point: nb, g: newG, f: newG + w*hNb})
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
