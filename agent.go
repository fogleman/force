package force

import "math"

type Agent struct {
	Position Point
	Target   Point
	Padding  float64
	Speed    float64
	Reverse  bool
}

func (agent *Agent) desiredDirection(grid *Grid, agents []*Agent) Point {
	src := agent.Position.IntPoint()
	dst := agent.Target.IntPoint()
	path := grid.Search(src, dst, agents)
	if len(path) < 2 {
		return agent.Target.Sub(agent.Position).Normalize()
	}
	a := agent.Position
	b := path[1].Point()
	return b.Sub(a).Normalize()
}

func (agent *Agent) direction(grid *Grid, agents []*Agent, index int) Point {
	const e = 4
	desired := agent.desiredDirection(grid, agents)
	direction := desired
	for i, other := range agents {
		if other == agent {
			continue
		}
		d := agent.Position.Sub(other.Position)
		l := d.Length()
		if l > 5 {
			continue
		}
		p := agent.Padding + other.Padding
		m := math.Pow(p, e) / math.Pow(l, e)
		if i < index {
			m *= 4
		}
		direction = direction.Add(d.MulScalar(m))
	}
	for _, wall := range grid.WallList {
		d := agent.Position.Sub(wall.Point())
		l := d.Abs().SubScalar(0.5).Max(Point{}).Length()
		if l > 5 {
			continue
		}
		p := agent.Padding
		m := math.Pow(p, e) / math.Pow(l, e)
		direction = direction.Add(d.MulScalar(m * 4))
	}
	agent.Reverse = desired.Dot(direction) < 0
	l := direction.Length()
	l = math.Max(0.5, l)
	l = math.Min(1, l)
	return direction.Normalize().MulScalar(l)
}
