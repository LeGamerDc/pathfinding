package demo

import (
	"math"
	"math/rand"
	"pathfinding/grid"
)

type Status int

const (
	StatusStop = Status(iota)
	StatusMoving
	StatusTerminate
)

type Agent struct {
	// param
	Pos    Vec2
	Target Vec2
	Speed  float64

	// inner
	st               Status
	lastPos, nextPos Vec2
	path             []Vec2
	dir              Vec2
	velocity         Vec2 // when status = Collide, use velocity as moving directions
	prefVelocity     Vec2
	wait             int // when status = stop, wait n tick for next route
}

func (a *Agent) frame(ti float64) {
	switch a.st {
	case StatusStop:
		// do nothing
	case StatusMoving:
		d := vSub(a.nextPos, a.Pos)
		l2 := d.absSq()
		if l2 <= Epsilon {
			// arrived
			return
		}
		l := math.Sqrt(l2)
		if ti*a.Speed > l {
			a.Pos = a.nextPos
		} else {
			a.dir = d.div(l)
			a.Pos = vAdd(a.Pos, a.dir.mul(ti*a.Speed))
		}
	}
}

func (w *World) setObstacle(me *Agent) {
	w.ws.Fill(0, 0)
	for _, a := range w.Agents {
		if a == me {
			continue
		}
		if a.st == StatusStop || a.st == StatusTerminate {
			x, y := pos2grid(a.Pos)
			w.ws.Set(x, y)
		}
	}
}

func (a *Agent) route(w *World) (path []Vec2, ok bool) {
	fx, fy := pos2grid(a.Pos)
	tx, ty := pos2grid(a.Target)
	//fmt.Printf("%d,%d -> %d,%d\n", fx, fy, tx, ty)
	//defer fmt.Printf("%d %v\n", len(path), ok)
	if fx == tx && fy == ty {
		// 1. same grid
		return []Vec2{a.Pos, a.Target}, true
	}
	w.setObstacle(a)
	//if w.ws.Los(a.Pos.X, a.Pos.Y, a.Target.X, a.Target.Y) {
	//	// 2. direct move
	//	return []Vec2{a.Pos, a.Target}, true
	//}
	// 3. grid pathfinding
	var p []grid.PathNode
	p, ok = w.ws.Solve(fx, fy, tx, ty)
	if !ok {
		return nil, false
	}
	// todo smooth path
	path = make([]Vec2, 0, len(p)-1)
	for i := 1; i < len(p)-1; i++ {
		path = append(path, gridCenter(p[i].X, p[i].Y))
	}
	path = append(path, a.Target)
	return path, true
}

func (a *Agent) Stop() {
	a.st = StatusStop
	a.Pos = occupy(a.Pos)
	a.nextPos = a.Pos
	a.wait = nextWait()
}

func pos2grid(pos Vec2) (x, y int32) {
	return int32(math.Floor(pos.X)), int32(math.Floor(pos.Y))
}

func sameGrid(a, b Vec2) bool {
	ax, ay := pos2grid(a)
	bx, by := pos2grid(b)
	return ax == bx && ay == by
}

func gridCenter(x, y int32) Vec2 {
	return Vec2{
		X: float64(x) + 0.5,
		Y: float64(y) + 0.5,
	}
}

func occupy(pos Vec2) Vec2 {
	xx, yy := math.Floor(pos.X), math.Floor(pos.Y)
	rx, ry := pos.X-xx, pos.Y-yy
	if rx > 0.75 {
		rx = 0.75
	}
	if rx < 0.25 {
		rx = 0.25
	}
	if ry > 0.75 {
		ry = 0.75
	}
	if ry < 0.25 {
		ry = 0.25
	}
	return Vec2{X: xx + rx, Y: yy + ry}
}

func (a *Agent) getPathNextPoint() (Vec2, bool) {
	for len(a.path) > 1 {
		if sameGrid(a.Pos, a.path[0]) {
			a.path = a.path[1:]
		} else {
			break
		}
	}
	if len(a.path) > 0 && vDistSq(a.Pos, a.path[0]) > Epsilon {
		return a.path[0], true
	}
	return defaultDir, false
}

func (a *Agent) setLocalTarget(pos Vec2, ti float64) {
	d := vSub(pos, a.Pos)
	move := a.Speed * ti
	if d.absSq() <= move*move {
		a.nextPos = pos
		a.prefVelocity = d.div(ti)
	} else {
		a.nextPos = vAdd(a.Pos, d.normalize().mul(move))
		a.prefVelocity = vSub(a.nextPos, a.Pos).div(ti)
	}
}

func nextWait() int {
	return rand.Intn(4)
}
