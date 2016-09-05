package force

import (
	"image"
	"math/rand"

	"github.com/fogleman/gg"
)

type Model struct {
	Grid   *Grid
	Agents []*Agent
}

func NewModel() *Model {
	grid := NewGrid(32, 32)
	for x := 0; x < grid.W; x++ {
		grid.AddWall(x, 0)
		grid.AddWall(x, grid.H-1)
	}
	for y := 0; y < grid.W; y++ {
		grid.AddWall(0, y)
		grid.AddWall(grid.W-1, y)
	}
	for i := 0; i < 100; i++ {
		x, y := grid.RandomEmptyCell()
		grid.AddWall(x, y)
	}
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
		agent.Speed = 2
		agents = append(agents, &agent)
	}
	return &Model{grid, agents}
}

func (model *Model) Step(t, dt float64) {
	const n = 1
	for i := 0; i < n; i++ {
		model.step(dt / n)
	}
}

func (model *Model) step(dt float64) {
	vectors := make([]Point, len(model.Agents))
	for i, agent := range model.Agents {
		vectors[i] = agent.direction(model.Grid, model.Agents, i)
	}
	for i, agent := range model.Agents {
		v := vectors[i].MulScalar(dt * agent.Speed)
		agent.Position = agent.Position.Add(v)
		if agent.Target.Sub(agent.Position).Length() < agent.Padding*2 {
			tx, ty := model.Grid.RandomEmptyCell()
			agent.Target = Point{float64(tx), float64(ty)}
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
	for _, agent := range model.Agents {
		points := g.Search(agent.Position.IntPoint(), agent.Target.IntPoint(), model.Agents)
		for _, point := range points {
			p := point.Point()
			x, y := p.X*s+s/2, p.Y*s+s/2
			dc.LineTo(x, y)
		}
		dc.SetRGBA(1, 0, 0, 0.05)
		dc.SetLineWidth(s / 2)
		dc.Stroke()
	}

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

	return dc.Image()
}
