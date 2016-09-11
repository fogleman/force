package force

import (
	"image"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/fogleman/gg"
	"github.com/ojrac/opensimplex-go"
)

func noise(source *opensimplex.Noise, n int, x, y, scale float64) float64 {
	maxAmp := 0.0
	amp := 1.0
	freq := scale
	noise := 0.0

	for i := 0; i < n; i++ {
		noise += (source.Eval2(x*freq, y*freq) + 0.8) / 1.6 * amp
		maxAmp += amp
		amp *= 0.5
		freq *= 2
	}

	noise /= maxAmp
	return noise
}

type Spot struct {
	Point
	Time time.Time
}

type Model struct {
	Grid   *Grid
	Agents []*Agent
	Spots  []*Spot
}

func NewModel() *Model {
	const w = 64
	const h = 64
	grid := NewGrid(w, h)

	const inset = 6
	// spots := []*Spot{
	// 	{Point{inset, inset}, time.Now()},
	// 	{Point{w - inset - 1, inset}, time.Now()},
	// 	{Point{inset, h - inset - 1}, time.Now()},
	// 	{Point{w - inset - 1, h - inset - 1}, time.Now()},
	// 	{Point{w / 2, h / 2}, time.Now()},
	// 	{Point{inset, h / 2}, time.Now()},
	// 	{Point{w / 2, inset}, time.Now()},
	// 	{Point{w / 2, h - inset - 1}, time.Now()},
	// 	{Point{w - inset - 1, h / 2}, time.Now()},
	// }

	var spots []*Spot
	for i := 0; i < 4; i++ {
		x := rand.Intn(w)
		y := rand.Intn(h)
		if x < inset || y < inset || x >= w-inset || y >= h-inset {
			i--
			continue
		}
		spots = append(spots, &Spot{Point{float64(x), float64(y)}, time.Now()})
	}

	for x := 0; x < grid.W; x++ {
		grid.AddWall(x, 0)
		grid.AddWall(x, grid.H-1)
	}
	for y := 0; y < grid.H; y++ {
		grid.AddWall(0, y)
		grid.AddWall(grid.W-1, y)
	}
	source := opensimplex.NewWithSeed(time.Now().UTC().UnixNano())
	for x := 1; x < grid.W-1; x++ {
		for y := 1; y < grid.H-1; y++ {
			ok := true
			for _, spot := range spots {
				if math.Abs(float64(x)-spot.Point.X) <= 4 && math.Abs(float64(y)-spot.Point.Y) <= 4 {
					ok = false
				}
			}
			if !ok {
				continue
			}
			w := noise(source, 3, float64(x), float64(y), 0.3)
			if w > 0.7 {
				grid.AddWall(x, y)
			}
		}
	}
	// for i := 0; i < 100; i++ {
	// 	x, y := grid.RandomEmptyCell()
	// 	ok := true
	// 	for _, spot := range spots {
	// 		if math.Abs(float64(x)-spot.Point.X) <= 2 && math.Abs(float64(y)-spot.Point.Y) <= 2 {
	// 			ok = false
	// 		}
	// 	}
	// 	if !ok {
	// 		i--
	// 		continue
	// 	}
	// 	grid.AddWall(x, y)
	// }
	var agents []*Agent
	seen := make(map[IntPoint]bool)
	for i := 0; i < 200; i++ {
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
		agent.Target = spots[rand.Intn(len(spots))]
		agent.Padding = 0.4
		agent.Speed = 4
		agents = append(agents, &agent)
	}
	return &Model{grid, agents, spots}
}

func (model *Model) Step(t, dt float64) {
	const sps = 60
	n := int(math.Ceil(sps * dt))
	for i := 0; i < n; i++ {
		model.step(dt / float64(n))
	}
}

func (model *Model) step(dt float64) {
	const alpha = 0.1
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
		if agent.Target.Sub(agent.Position).Length() < 1 {
			agent.Target.Time = time.Now()
			agent.Target = model.Spots[rand.Intn(len(model.Spots))]
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
		p := spot.Point
		dc.DrawCircle(p.X*s+s/2, p.Y*s+s/2, s/2)
		d := time.Since(spot.Time).Seconds()
		d = math.Max(0, 1-d)
		d = math.Pow(d, 2)
		dc.SetRGBA(1, 0, 0, 1)
		dc.FillPreserve()
		dc.SetRGBA(1, 0, 0, d)
		dc.SetLineWidth(s / 4)
		dc.Stroke()
	}

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
	dc.SetLineWidth(s / 6)
	dc.Stroke()

	return dc.Image()
}
