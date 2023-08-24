package rvo2

import "math"

type Line struct {
	Point, Direction Vec2
}

func (a *Agent) linearProgram(point, direction Vec2, velocity Vec2) (newVelocity Vec2, update, ok bool) {
	newVelocity = velocity
	if vCross(direction, vSub(point, newVelocity)) > 0.0 {
		dotProduct := vDot(point, direction)
		discriminant := dotProduct*dotProduct + a.Speed*a.Speed - point.absSq()
		if discriminant < 0.0 {
			return
		}
		sqrtDiscriminant := math.Sqrt(discriminant)
		tLeft := -dotProduct - sqrtDiscriminant
		tRight := -dotProduct + sqrtDiscriminant
		for _, line := range a.lines {
			denominator := vCross(direction, line.Direction)
			numerator := vCross(line.Direction, vSub(point, line.Point))

			if math.Abs(denominator) <= Epsilon {
				if numerator < 0.0 {
					return
				}
				continue
			}

			t := numerator / denominator
			if denominator >= 0.0 {
				tRight = math.Min(tRight, t)
			} else {
				tLeft = math.Max(tLeft, t)
			}
			if tLeft > tRight {
				return
			}
		}
		t := vDot(direction, vSub(a.PrefVelocity, point))
		if t < tLeft {
			t = tLeft
		} else if t > tRight {
			t = tRight
		}
		newVelocity = vAdd(point, direction.mul(t))
		a.lines = append(a.lines, Line{Point: point, Direction: direction})
		update = true
	}
	return newVelocity, update, true
}
