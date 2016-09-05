package force

import "math"

type IntPoint struct {
	X, Y int
}

func (p IntPoint) Point() Point {
	return Point{float64(p.X), float64(p.Y)}
}

type Point struct {
	X, Y float64
}

func (p Point) IntPoint() IntPoint {
	return IntPoint{Round(p.X), Round(p.Y)}
}

func (a Point) Normalize() Point {
	return a.DivScalar(a.Length())
}

func (a Point) Abs() Point {
	return Point{math.Abs(a.X), math.Abs(a.Y)}
}

func (a Point) Length() float64 {
	return math.Sqrt(a.X*a.X + a.Y*a.Y)
}

func (a Point) Dot(b Point) float64 {
	return a.X*b.X + a.Y*b.Y
}

func (a Point) Add(b Point) Point {
	return Point{a.X + b.X, a.Y + b.Y}
}

func (a Point) Sub(b Point) Point {
	return Point{a.X - b.X, a.Y - b.Y}
}

func (a Point) SubScalar(b float64) Point {
	return Point{a.X - b, a.Y - b}
}

func (a Point) MulScalar(b float64) Point {
	return Point{a.X * b, a.Y * b}
}

func (a Point) DivScalar(b float64) Point {
	return Point{a.X / b, a.Y / b}
}

func (a Point) Max(b Point) Point {
	return Point{math.Max(a.X, b.X), math.Max(a.Y, b.Y)}
}

func (a Point) Rotate(theta float64) Point {
	return Point{
		a.X*math.Cos(theta) - a.Y*math.Sin(theta),
		a.X*math.Sin(theta) + a.Y*math.Cos(theta),
	}
}
