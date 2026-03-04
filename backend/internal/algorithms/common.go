package algorithms

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

type Cell struct {
	Obstacle bool
	Cost     int
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (p Point) String() string {
	return fmt.Sprintf("%d,%d", p.X, p.Y)
}

type AlgorithmParams struct {
	Neighbors  int     `json:"neighbors"`
	Heuristic  string  `json:"heuristic"`
	W          float64 `json:"w"`
	TieBreak   string  `json:"tie_break"`
}

type AlgorithmInput struct {
	Grid   [][]Cell
	Start  Point
	Goal   Point
	Params AlgorithmParams
}

type AlgorithmOutput struct {
	Found         bool
	Path          []Point
	ExpandedNodes int
	TimeMicros    int64
	Turns         int
	ExploredOrder map[string]int
}

type Algorithm interface {
	Run(input AlgorithmInput) AlgorithmOutput
}

func ParseParams(raw json.RawMessage) AlgorithmParams {
	p := AlgorithmParams{
		Neighbors: 4,
		Heuristic: "manhattan",
		W:         1.0,
	}
	if raw != nil {
		_ = json.Unmarshal(raw, &p)
	}
	if p.W < 1.0 {
		p.W = 1.0
	}
	return p
}

func ParseGrid(gridText string, width, height int) [][]Cell {
	grid := make([][]Cell, height)
	for i := range grid {
		grid[i] = make([]Cell, width)
		for j := range grid[i] {
			grid[i][j] = Cell{Cost: 1}
		}
	}
	lines := strings.Split(gridText, "\n")
	for y, line := range lines {
		if y >= height {
			break
		}
		for x, ch := range line {
			if x >= width {
				break
			}
			if ch == '#' {
				grid[y][x].Obstacle = true
			}
		}
	}
	return grid
}

func Neighbors4(p Point, grid [][]Cell) []Point {
	dirs := []Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	height := len(grid)
	width := len(grid[0])
	var result []Point
	for _, d := range dirs {
		nx, ny := p.X+d.X, p.Y+d.Y
		if nx >= 0 && nx < width && ny >= 0 && ny < height && !grid[ny][nx].Obstacle {
			result = append(result, Point{nx, ny})
		}
	}
	return result
}

func Neighbors8(p Point, grid [][]Cell) []Point {
	dirs := []Point{
		{0, -1}, {0, 1}, {-1, 0}, {1, 0},
		{-1, -1}, {1, -1}, {-1, 1}, {1, 1},
	}
	height := len(grid)
	width := len(grid[0])
	var result []Point
	for _, d := range dirs {
		nx, ny := p.X+d.X, p.Y+d.Y
		if nx >= 0 && nx < width && ny >= 0 && ny < height && !grid[ny][nx].Obstacle {
			result = append(result, Point{nx, ny})
		}
	}
	return result
}

func GetNeighbors(p Point, grid [][]Cell, n int) []Point {
	if n == 8 {
		return Neighbors8(p, grid)
	}
	return Neighbors4(p, grid)
}

func Manhattan(a, b Point) int {
	return int(math.Abs(float64(a.X-b.X)) + math.Abs(float64(a.Y-b.Y)))
}

func ReconstructPath(parent map[string]Point, start, goal Point) []Point {
	var path []Point
	cur := goal
	for cur != start {
		path = append(path, cur)
		p, ok := parent[cur.String()]
		if !ok {
			break
		}
		cur = p
	}
	path = append(path, start)
	// reverse
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func CountTurns(path []Point) int {
	if len(path) < 3 {
		return 0
	}
	turns := 0
	for i := 1; i < len(path)-1; i++ {
		dx1 := path[i].X - path[i-1].X
		dy1 := path[i].Y - path[i-1].Y
		dx2 := path[i+1].X - path[i].X
		dy2 := path[i+1].Y - path[i].Y
		if dx1 != dx2 || dy1 != dy2 {
			turns++
		}
	}
	return turns
}

