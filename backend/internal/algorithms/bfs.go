package algorithms

import "time"

type BFS struct{}

func (b *BFS) Run(input AlgorithmInput) AlgorithmOutput {
	start := time.Now()
	grid := input.Grid
	src, dst := input.Start, input.Goal
	n := input.Params.Neighbors

	if grid[src.Y][src.X].Obstacle || grid[dst.Y][dst.X].Obstacle {
		return AlgorithmOutput{Found: false}
	}

	visited := make(map[string]bool)
	parent := make(map[string]Point)
	order := make(map[string]int)
	queue := []Point{src}
	visited[src.String()] = true
	expanded := 0

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		order[cur.String()] = expanded
		expanded++

		if cur == dst {
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

		for _, nb := range GetNeighbors(cur, grid, n) {
			if !visited[nb.String()] {
				visited[nb.String()] = true
				parent[nb.String()] = cur
				queue = append(queue, nb)
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
