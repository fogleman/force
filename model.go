package force

import (
	"fmt"
	"image"
	"math"
	"math/rand"

	"github.com/fogleman/gg"
)

type Model struct {
	Grid    *Grid
	Agents  []*Agent
	Targets int
}

func NewModel() *Model {
	grid := NewGrid(16, 16)

	grid.AddWall(3, 3)
	grid.AddWall(4, 3)
	grid.AddWall(5, 3)
	grid.AddWall(6, 3)
	grid.AddWall(9, 3)
	grid.AddWall(10, 3)
	grid.AddWall(11, 3)
	grid.AddWall(12, 3)

	grid.AddWall(3, 12)
	grid.AddWall(4, 12)
	grid.AddWall(5, 12)
	grid.AddWall(6, 12)
	grid.AddWall(9, 12)
	grid.AddWall(10, 12)
	grid.AddWall(11, 12)
	grid.AddWall(12, 12)

	grid.AddWall(3, 4)
	grid.AddWall(3, 5)
	grid.AddWall(3, 6)
	grid.AddWall(3, 9)
	grid.AddWall(3, 10)
	grid.AddWall(3, 11)

	grid.AddWall(12, 4)
	grid.AddWall(12, 5)
	grid.AddWall(12, 6)
	grid.AddWall(12, 9)
	grid.AddWall(12, 10)
	grid.AddWall(12, 11)

	grid.AddWall(3, 1)
	grid.AddWall(3, 2)
	grid.AddWall(3, 3)
	grid.AddWall(3, 12)
	grid.AddWall(3, 13)
	grid.AddWall(3, 14)

	grid.AddWall(12, 1)
	grid.AddWall(12, 2)
	grid.AddWall(12, 3)
	grid.AddWall(12, 12)
	grid.AddWall(12, 13)
	grid.AddWall(12, 14)

	for x := 0; x < grid.W; x++ {
		grid.AddWall(x, 0)
		grid.AddWall(x, grid.H-1)
	}
	for y := 0; y < grid.H; y++ {
		grid.AddWall(0, y)
		grid.AddWall(grid.W-1, y)
	}
	// for y := 0; y < grid.H; y++ {
	// 	for x := 0; x < grid.W; x++ {
	// 		if y%4 == 0 && x%4 == 0 {
	// 			grid.AddWall(x, y)
	// 		}
	// 	}
	// }
	// for i := 0; i < 50; i++ {
	// 	x, y := grid.RandomEmptyCell()
	// 	grid.AddWall(x, y)
	// }
	var agents []*Agent
	seen := make(map[IntPoint]bool)
	for i := 0; i < 100; i++ {
		var x, y int
		for {
			x, y = grid.RandomEmptyCell()
			if _, ok := seen[IntPoint{x, y}]; !ok {
				break
			}
		}
		seen[IntPoint{x, y}] = true
		tx, ty := grid.RandomEmptyCell()
		agent := Agent{}
		agent.Position = Point{float64(x) + rand.Float64() - 0.5, float64(y) + rand.Float64() - 0.5}
		agent.Target = Point{float64(tx), float64(ty)}
		agent.Padding = 0.2
		agent.Speed = 1
		agents = append(agents, &agent)
	}
	return &Model{grid, agents, 0}
}

func (model *Model) Step(t, dt float64) {
	tps := float64(model.Targets) / t
	fmt.Printf("%.1f, %.1f\n", t, tps)
	const sps = 120
	n := int(math.Ceil(sps * dt))
	for i := 0; i < n; i++ {
		model.step(dt / float64(n))
	}
}

func (model *Model) step(dt float64) {
	const alpha = 0.1 / 3
	vectors := make([]Point, len(model.Agents))
	for i, agent := range model.Agents {
		desired, actual := agent.direction(model.Grid, model.Agents, i)

		agent.Direction = agent.Direction.Sub(agent.Direction.Sub(actual).MulScalar(alpha))
		vectors[i] = agent.Direction

		pointer := desired.Add(actual).Normalize()
		agent.Pointer = agent.Pointer.Sub(agent.Pointer.Sub(pointer).MulScalar(alpha))
	}
	for i, agent := range model.Agents {
		v := vectors[i].MulScalar(dt * agent.Speed)
		agent.Position = agent.Position.Add(v)
		if agent.Target.Sub(agent.Position).Length() < agent.Padding*2 {
			tx, ty := model.Grid.RandomEmptyCell()
			agent.Target = Point{float64(tx), float64(ty)}
			model.Targets++
		}
	}
}

func (model *Model) Draw(w, h int, mx, my float64) image.Image {
	g := model.Grid
	sw := float64(w / g.W)
	sh := float64(h / g.H)
	s := sw
	if sh < sw {
		s = sh
	}

	// create context
	dc := gg.NewContext(w, h)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// draw walls
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			if g.HasWall(x, y) {
				fx := float64(x) * s
				fy := float64(y) * s
				dc.DrawRectangle(fx, fy, s, s)
			}
		}
	}
	dc.SetRGB(0, 0, 0)
	dc.Fill()

	// draw paths
	// for _, agent := range model.Agents {
	// 	points := g.Search(agent.Position.IntPoint(), agent.Target.IntPoint(), model.Agents)
	// 	for _, point := range points {
	// 		p := point.Point()
	// 		x, y := p.X*s+s/2, p.Y*s+s/2
	// 		dc.LineTo(x, y)
	// 	}
	// 	dc.SetRGBA(1, 0, 0, 0.025)
	// 	dc.SetLineWidth(s / 2)
	// 	dc.Stroke()
	// }

	// draw agents
	for _, agent := range model.Agents {
		point := agent.Position
		radius := agent.Padding * s
		x, y := point.X*s+s/2, point.Y*s+s/2
		dc.DrawCircle(x, y, radius)
		dc.SetRGB(0, 0, 1)
		// if agent.Reverse {
		// 	dc.SetRGB(1, 0, 0)
		// }
		dc.Fill()
	}

	// draw directions
	for _, agent := range model.Agents {
		point := agent.Position
		radius := agent.Padding * s
		direction := agent.Pointer.Normalize()
		x1, y1 := point.X*s+s/2, point.Y*s+s/2
		x2 := x1 + direction.X*radius
		y2 := y1 + direction.Y*radius
		dc.DrawLine(x1, y1, x2, y2)
	}
	dc.SetRGB(1, 1, 1)
	dc.SetLineWidth(3)
	dc.Stroke()

	return dc.Image()
}
