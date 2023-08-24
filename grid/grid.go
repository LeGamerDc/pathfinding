package grid

import "fmt"

// map are always little-endian

type Grid struct {
	Bits [16]uint16
}

type Terrain struct {
	Grids  [][]Grid
	Nx, Ny int32
}

type WorkSpace struct {
	Map [144]uint16
	T   *Terrain

	pool       *gNodePool
	queue      *gNodeQueue
	endX, endY int32
}

func NewWorkSpace(t *Terrain) *WorkSpace {
	return &WorkSpace{
		Map:   [144]uint16{},
		T:     t,
		pool:  newNodePool(256),
		queue: newNodeQueue(256),
	}
}

func (t *Terrain) Set(x, y int32) {
	if x < 0 || y < 0 {
		return
	}
	xx, yy := x/16, y/16
	rx, ry := x%16, y%16
	if xx >= t.Nx || yy >= t.Ny {
		return
	}
	t.Grids[xx][yy].Bits[ry] |= 1 << rx
}

func (t *Terrain) IsSet(x, y int32) bool {
	if x < 0 || y < 0 {
		return false
	}
	xx, yy := x/16, y/16
	rx, ry := x%16, y%16
	if xx >= t.Nx || yy >= t.Ny {
		return false
	}
	return t.Grids[xx][yy].Bits[ry]&(1<<rx) != 0
}

func (t *Terrain) Los(ax, ay, bx, by int32) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%f %f %f %f\n", ax, ay, bx, by)
			panic("xxx")
		}
	}()
	var (
		x, y, ix, iy, k int32
		dx, dy          int32 = 1, 1
	)
	x = bx - ax
	y = by - ay
	if x < 0 {
		x, dx = -x, -1
	}
	if y < 0 {
		y, dy = -y, -1
	}
	for ix < x || iy < y {
		k = (1+2*ix)*y - (1+2*iy)*x
		if k < 0 {
			ax += dx
			ix++
		} else if k > 0 {
			ay += dy
			iy++
		} else {
			if t.IsSet(ax, ay+dy) || t.IsSet(ax+dx, ay) {
				return false
			}
			ax += dx
			ay += dy
			ix++
			iy++
		}
		if t.IsSet(ax, ay) {
			return false
		}
	}
	return true
}

func (w *WorkSpace) Fill(ox, oy int32) {
	for i := int32(0); i < 3; i++ {
		for j := int32(0); j < 3; j++ {
			x, y := ox+i, oy+j
			if x < 0 || x >= w.T.Nx || y < 0 || y >= w.T.Ny {
				set1(&w.Map, i, j)
			} else {
				set(&w.Map, i, j, w.T.Grids[x][y].Bits)
			}
		}
	}
}

func (w *WorkSpace) IsSet(x, y int32) bool {
	if x < 0 || y < 0 || x >= 48 || y >= 48 {
		return true
	}
	return check(&w.Map, x/16, y/16, x%16, y%16)
}

func (w *WorkSpace) Set(x, y int32) {
	lock(&w.Map, x/16, y/16, x%16, y%16)
}

func (w *WorkSpace) Unset(x, y int32) {
	unlock(&w.Map, x/16, y/16, x%16, y%16)
}

func (w *WorkSpace) walkable(x, y, curDir, nextDir int32) bool {
	x, y = move(x, y, (curDir+nextDir)%8)
	return inWorkSpace(x, y) && !w.IsSet(x, y)
}

func set1(m *[144]uint16, i, j int32) {
	base := j*48 + i
	for p := int32(0); p < 16; p++ {
		m[base+3*p] = 0xffff
	}
}

func set(m *[144]uint16, i, j int32, g [16]uint16) {
	base := j*48 + i
	for p := int32(0); p < 16; p++ {
		m[base+3*p] = g[p]
	}
}

func lock(m *[144]uint16, i, j, x, y int32) {
	p := j*48 + y*3 + i
	m[p] |= 1 << x
}

func unlock(m *[48 * 3]uint16, i, j, x, y int32) {
	p := j*48 + y*3 + i
	m[p] &^= 1 << x
}

func check(m *[48 * 3]uint16, i, j, x, y int32) bool {
	p := j*48 + y*3 + i
	return m[p]&(1<<x) != 0
}

func inWorkSpace(x, y int32) bool {
	return x >= 0 && x < 48 && y >= 0 && y < 48
}
