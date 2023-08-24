package demo

import (
	"math"
)

type Vec2 struct {
	X, Y float64
}

const (
	Epsilon = 0.00001
	Eqs     = Epsilon * Epsilon
)

func equal(a, b float64) bool {
	return math.Abs(a-b) <= Epsilon
}

func (v Vec2) abs() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vec2) absSq() float64 {
	return v.X*v.X + v.Y*v.Y
}

func (v Vec2) minus() Vec2 {
	return Vec2{X: -v.X, Y: -v.Y}
}

func (v Vec2) mul(k float64) Vec2 {
	return Vec2{X: k * v.X, Y: k * v.Y}
}

func (v Vec2) div(k float64) Vec2 {
	return Vec2{X: v.X / k, Y: v.Y / k}
}

func (v Vec2) normalize() Vec2 {
	n := v.abs()
	return Vec2{X: v.X / n, Y: v.Y / n}
}

func vAdd(a, b Vec2) Vec2 {
	return Vec2{X: a.X + b.X, Y: a.Y + b.Y}
}

func vSub(a, b Vec2) Vec2 {
	return Vec2{X: a.X - b.X, Y: a.Y - b.Y}
}

func vDot(a, b Vec2) float64 {
	return a.X*b.X + a.Y*b.Y
}

func vCross(a, b Vec2) float64 {
	return a.X*b.Y - a.Y*b.X
}

func vDistSq(a, b Vec2) float64 {
	return vSub(a, b).absSq()
}

func vLeft(a, b, c Vec2) float64 {
	return vCross(vSub(b, a), vSub(c, a))
}

func collide(a, b, c Vec2, r float64) bool {
	if vDot(vSub(b, a), vSub(c, a)) < 0 {
		return false
	}
	return vDistSq(b, c) <= r*r
}

// angle a, b must be normalized
func angle(a, b Vec2) float64 {
	cos := vDot(a, b)
	sin := vCross(a, b)
	return math.Atan2(sin, cos)
}
