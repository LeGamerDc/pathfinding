package rvo2

import (
	"math"
)

type RvoConfig struct {
	TimeStep    float64 // 每次模拟时步进的时间 etc 0.25
	TimeHorizon float64 // 单位提前几s避免碰撞 etc 0.5-1

	MaxNeighbor int // 最多考虑n个碰撞邻居
}

type Agent struct {
	Position     Vec2 // 当前坐标
	PrefVelocity Vec2 // 寻路速度
	Velocity     Vec2 // 收敛速度

	Speed  float64 // 单位允许的最大速度
	Radius float64 // 单位碰撞半径

	lines []Line
}

func (c *RvoConfig) Solve(a *Agent, iterNeighbor func() (*Agent, bool)) (newVelocity Vec2, update, ok bool) {
	if a.PrefVelocity.absSq() > a.Speed*a.Speed {
		newVelocity = a.PrefVelocity.normalize().mul(a.Speed)
	} else {
		newVelocity = a.PrefVelocity
	}
	a.lines = a.lines[0:0]
	invTimeHorizon := 1.0 / c.TimeHorizon
	for i := 0; i < c.MaxNeighbor; i++ {
		var other *Agent
		if other, ok = iterNeighbor(); !ok {
			break
		}
		relativePosition := vSub(other.Position, a.Position)
		relativeVelocity := vSub(a.Velocity, other.Velocity)
		distSq := relativePosition.absSq()
		combinedRadius := a.Radius + other.Radius
		combinedRadiusSq := combinedRadius * combinedRadius

		var lineDir, lineU Vec2
		if distSq > combinedRadiusSq {
			// not collide
			w := vSub(relativeVelocity, relativePosition.mul(invTimeHorizon))
			wLengthSq := w.absSq()
			dotProduct := vDot(w, relativePosition)
			if dotProduct < 0.0 && dotProduct*dotProduct > combinedRadiusSq*wLengthSq {
				// Project on cut-off circle
				// (w * p)^2 = |w|^2 * |p|^2 * cos^2(θ) > r*r * |w|^2
				// => θ < arcsin(r/p)
				wLength := math.Sqrt(wLengthSq)
				unitW := w.div(wLength)

				lineDir = Vec2{X: unitW.Y, Y: -unitW.X}
				lineU = unitW.mul(combinedRadius*invTimeHorizon - wLength)
			} else {
				// Project on legs
				leg := math.Sqrt(distSq - combinedRadiusSq)
				if vCross(relativePosition, w) > 0.0 {
					// left leg
					lineDir = Vec2{
						X: relativePosition.X*leg - relativePosition.Y*combinedRadius,
						Y: relativePosition.X*combinedRadius + relativePosition.Y*leg,
					}.div(distSq)
				} else {
					// right leg
					lineDir = Vec2{
						X: relativePosition.X*leg + relativePosition.Y*combinedRadius,
						Y: relativePosition.Y*leg - relativePosition.X*combinedRadius,
					}.div(distSq)
				}
				// lineDir is normalized lineDir * (lineDir * v) = lineDir * (|v| * sin θ)
				// = v + u
				lineU = vSub(lineDir.mul(vDot(relativeVelocity, lineDir)), relativeVelocity)
			}
		} else {
			// already collide
			invTimeStep := 1.0 / c.TimeStep
			w := vSub(relativeVelocity, relativePosition.mul(invTimeStep))
			wLength := w.abs()
			unitW := w.div(wLength)
			lineDir = Vec2{X: unitW.Y, Y: -unitW.X}
			lineU = unitW.mul(combinedRadius*invTimeStep - wLength)
		}

		linePoint := vAdd(a.Velocity, lineU.mul(0.5))
		var u bool
		if newVelocity, u, ok = a.linearProgram(linePoint, lineDir, newVelocity); !ok {
			return Vec2{}, false, false
		} else if u {
			update = true
		}
	}
	// TODO consider terrain obstacle
	return newVelocity, update, true
}
