package force

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"sync"

	"github.com/fogleman/gg"
)

type Model struct {
	Grid    *Grid
	Agents  []*Agent
	Spots   []IntPoint
	Targets int
}

func NewModel() *Model {
	grid := NewGrid(32, 32)

	spots := []IntPoint{
		{2, 2},
		{29, 2},
		{2, 29},
		{29, 29},
		{16, 16},
		{2, 16},
		{16, 2},
		{16, 29},
		{29, 16},
	}

	for x := 0; x < grid.W; x++ {
		grid.AddWall(x, 0)
		grid.AddWall(x, grid.H-1)
	}
	for y := 0; y < grid.H; y++ {
		grid.AddWall(0, y)
		grid.AddWall(grid.W-1, y)
	}
	for i := 0; i < 100; i++ {
		x, y := grid.RandomEmptyCell()
		ok := true
		for _, spot := range spots {
			if math.Abs(float64(x-spot.X)) <= 2 && math.Abs(float64(y-spot.Y)) <= 2 {
				ok = false
			}
		}
		if !ok {
			i--
			continue
		}
		grid.AddWall(x, y)
	}
	var agents []*Agent
	seen := make(map[IntPoint]bool)
	for i := 0; i < 400; i++ {
		var x, y int
		for {
			x, y = grid.RandomEmptyCell()
			if _, ok := seen[IntPoint{x, y}]; !ok {
				break
			}
		}
		seen[IntPoint{x, y}] = true
		agent := Agent{}
		agent.Position = Point{float64(x) + rand.Float64() - 0.5, float64(y) + rand.Float64() - 0.5}
		agent.Target = spots[rand.Intn(len(spots))].Point()
		agent.Padding = 0.2
		agent.Speed = 2
		agents = append(agents, &agent)
	}
	return &Model{grid, agents, spots, 0}
}

func (model *Model) Step(t, dt float64) {
	tps := float64(model.Targets) / t
	fmt.Printf("%.1f, %.1f\n", t, tps)
	const sps = 60
	n := int(math.Ceil(sps * dt))
	for i := 0; i < n; i++ {
		model.step(dt / float64(n))
	}
}

func (model *Model) step(dt float64) {
	const alpha = 0.08
	vectors := make([]Point, len(model.Agents))
	var wg sync.WaitGroup
	for i, agent := range model.Agents {
		wg.Add(1)
		go func(i int, agent *Agent) {
			desired, actual := agent.direction(model.Grid, model.Agents, i)
			agent.Direction = agent.Direction.Sub(agent.Direction.Sub(actual).MulScalar(alpha))
			vectors[i] = agent.Direction
			pointer := desired.MulScalar(1).Add(actual).Normalize()
			agent.Pointer = agent.Pointer.Sub(agent.Pointer.Sub(pointer).MulScalar(alpha))
			wg.Done()
		}(i, agent)
	}
	wg.Wait()
	for i, agent := range model.Agents {
		v := vectors[i].MulScalar(dt * agent.Speed)
		agent.Position = agent.Position.Add(v)
		if agent.Target.Sub(agent.Position).Length() < 0.75 {
			agent.Target = model.Spots[rand.Intn(len(model.Spots))].Point()
			model.Targets++
		}
	}
	model.Grid.UpdateCost(model.Agents)
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

	// draw spots
	for _, spot := range model.Spots {
		p := spot.Point()
		dc.DrawCircle(p.X*s+s/2, p.Y*s+s/2, s/2)
	}
	dc.SetRGBA(1, 0, 0, 0.5)
	dc.Fill()

	// draw paths
	for _, agent := range model.Agents {
		break
		points := g.Search(agent.Position.IntPoint(), agent.Target.IntPoint(), model.Agents)
		for _, point := range points {
			p := point.Point()
			x, y := p.X*s+s/2, p.Y*s+s/2
			dc.LineTo(x, y)
		}
		dc.SetRGBA(1, 0, 0, 0.25)
		dc.SetLineWidth(s / 2)
		dc.Stroke()
		break
	}

	// draw agents
	for i, agent := range model.Agents {
		point := agent.Position
		radius := agent.Padding * s
		x, y := point.X*s+s/2, point.Y*s+s/2
		dc.DrawCircle(x, y, radius)
		dc.SetRGB(0, 0, 1)
		// if agent.Reverse {
		// 	dc.SetRGB(1, 0, 0)
		// }
		if i == 0 {
			dc.SetRGB(1, 0, 0)
		}
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
	dc.SetLineWidth(s / 12)
	dc.Stroke()

	return dc.Image()
}
