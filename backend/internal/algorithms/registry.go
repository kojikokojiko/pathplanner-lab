package algorithms

import "fmt"

func Get(key string) (Algorithm, error) {
	switch key {
	case "BFS":
		return &BFS{}, nil
	case "DIJKSTRA":
		return &Dijkstra{}, nil
	case "ASTAR":
		return &AStar{}, nil
	case "WEIGHTED_ASTAR":
		return &WeightedAStar{}, nil
	case "GREEDY":
		return &Greedy{}, nil
	default:
		return nil, fmt.Errorf("unknown algorithm: %s", key)
	}
}
