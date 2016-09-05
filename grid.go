package force

import (
	"math"
	"math/rand"

	"github.com/fogleman/astar"
)

type Grid struct {
	W, H     int
	WallGrid []bool
	WallList []IntPoint
	CostGrid []float64
	cache    map[IntPoint][]IntPoint
}

func NewGrid(w, h int) *Grid {
	wall := make([]bool, w*h)
	cost := make([]float64, w*h)
	return &Grid{w, h, wall, nil, cost, nil}
}

func (g *Grid) i(x, y int) int {
	return y*g.W + x
}

func (g *Grid) xy(i int) (int, int) {
	return i % g.W, i / g.W
}

func (g *Grid) AddWall(x, y int) {
	g.WallList = append(g.WallList, IntPoint{x, y})
	g.WallGrid[g.i(x, y)] = true
	g.cache = nil
}

// func (g *Grid) RemoveWall(x, y int) {
// 	g.WallGrid[g.i(x, y)] = false
// 	g.cache = nil
// }

func (g *Grid) HasWall(x, y int) bool {
	return g.WallGrid[g.i(x, y)]
}

func (g *Grid) RandomEmptyCell() (int, int) {
	for {
		x, y := rand.Intn(g.W), rand.Intn(g.H)
		if !g.HasWall(x, y) {
			return x, y
		}
	}
}

func (g *Grid) Edges(node int) []astar.Edge {
	x, y := g.xy(node)
	if x < 0 || y < 0 || x >= g.W || y >= g.H {
		return nil
	}
	if g.WallGrid[node] {
		return nil
	}
	edges := make([]astar.Edge, 0, 8)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx := x + dx
			ny := y + dy
			if nx < 0 || ny < 0 || nx >= g.W || ny >= g.H {
				continue
			}
			i := g.i(nx, ny)
			if g.WallGrid[i] {
				continue
			}
			distance := 1.0
			if dx != 0 && dy != 0 {
				if g.HasWall(x+dx, y) || g.HasWall(x, y+dy) {
					continue
				}
				distance = math.Sqrt2
			}
			cost := g.CostGrid[i]
			edge := astar.Edge{i, distance + cost}
			edges = append(edges, edge)
		}
	}
	return edges
}

func (g *Grid) Estimate(src, dst int) float64 {
	x1, y1 := g.xy(src)
	x2, y2 := g.xy(dst)
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(float64(dx*dx + dy*dy))
}

func (g *Grid) Search(src, dst IntPoint, agents []*Agent) []IntPoint {
	s := g.i(src.X, src.Y)
	d := g.i(dst.X, dst.Y)
	key := IntPoint{s, d}
	if points, ok := g.cache[key]; ok {
		return points
	}
	for i := range g.CostGrid {
		g.CostGrid[i] = 0
	}
	for _, agent := range agents {
		p := agent.Position.IntPoint()
		if p.X < 0 || p.Y < 0 || p.X >= g.W || p.Y >= g.H {
			continue
		}
		g.CostGrid[g.i(p.X, p.Y)] += 0.25
	}
	result := astar.Search(g, s, d)
	points := make([]IntPoint, len(result.Nodes))
	for i, node := range result.Nodes {
		x, y := g.xy(node)
		points[i] = IntPoint{x, y}
	}
	if g.cache == nil {
		g.cache = make(map[IntPoint][]IntPoint)
	}
	g.cache[key] = points
	return points
}
