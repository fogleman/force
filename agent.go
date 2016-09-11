package force

import "math"

type Agent struct {
	Position  Point
	Target    Point
	Direction Point
	Pointer   Point
	Padding   float64
	Speed     float64
	Reverse   bool
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

func (agent *Agent) direction(grid *Grid, agents []*Agent, index int) (Point, Point) {
	const e = 3
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
		if i > index {
			m *= 4
		} else {
			m *= 6
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
		direction = direction.Add(d.MulScalar(m * 2))
	}
	agent.Reverse = desired.Dot(direction) < 0
	l := direction.Length()
	l = math.Min(1, l)
	l = math.Max(0.2, l)
	direction = direction.Normalize().MulScalar(l)
	return desired, direction
}
