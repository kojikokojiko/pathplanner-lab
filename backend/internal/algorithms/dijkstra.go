package algorithms

import (
	"container/heap"
	"time"
)

type dijkstraItem struct {
	point Point
	cost  float64
	index int
}

type dijkstraPQ []*dijkstraItem

func (pq dijkstraPQ) Len() int            { return len(pq) }
func (pq dijkstraPQ) Less(i, j int) bool  { return pq[i].cost < pq[j].cost }
func (pq dijkstraPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *dijkstraPQ) Push(x interface{}) {
	item := x.(*dijkstraItem)
	item.index = len(*pq)
	*pq = append(*pq, item)
}
func (pq *dijkstraPQ) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[:n-1]
	return item
}

type Dijkstra struct{}

func (d *Dijkstra) Run(input AlgorithmInput) AlgorithmOutput {
	start := time.Now()
	grid := input.Grid
	src, dst := input.Start, input.Goal
	n := input.Params.Neighbors

	dist := make(map[string]float64)
	parent := make(map[string]Point)
	order := make(map[string]int)
	dist[src.String()] = 0

	pq := &dijkstraPQ{{point: src, cost: 0}}
	heap.Init(pq)
	expanded := 0

	for pq.Len() > 0 {
		cur := heap.Pop(pq).(*dijkstraItem)
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
			newCost := cur.cost + float64(grid[nb.Y][nb.X].Cost)
			nbKey := nb.String()
			if old, ok := dist[nbKey]; !ok || newCost < old {
				dist[nbKey] = newCost
				parent[nbKey] = cur.point
				heap.Push(pq, &dijkstraItem{point: nb, cost: newCost})
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
