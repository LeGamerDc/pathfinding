package hex

import (
	"github.com/legamerdc/pathfinding/grid"

	"slices"
)

/*
  ↑ y → x
	  4s  5
	3   *   0s
	  2s  1
*/

type dirSet uint32

const (
	_fullDirSetdirSet dirSet = (1 << 6) - 1
	_emptyDirSet      dirSet = 0

	_noDir = 0xff
)

func (s *dirSet) dirAdd(d int32) {
	*s |= 1 << d
}

func (s *dirSet) dirIter(f func(d int32) bool) {
	for d := int32(0); d < 6; d++ {
		if (*s)&(1<<d) != 0 {
			if !f(d) {
				return
			}
		}
	}
}

type WorkSpace struct {
	Map *grid.Local

	pool       *grid.NodePool
	heap       *grid.NodeQueue
	endX, endY int32
}

func NewWorkSpace(size int) *WorkSpace {
	return &WorkSpace{
		pool: grid.NewNodePool(int32(size)),
		heap: grid.NewNodeQueue(size),
	}
}

func (ws *WorkSpace) Reset(m *grid.Local) {
	ws.Map = m
}

func (ws *WorkSpace) Solve(sx, sy, ex, ey int32) (p []grid.PathGrid, ok bool) {
	ws.pool.Clear()
	ws.heap.Clear()
	ws.endX, ws.endY = ex, ey
	ws.putInOpenSet(sx, sy, _noDir, sx, sy, 0)
	for {
		x, y, d, c, ok1 := ws.getOutOpenSet()
		if !ok1 {
			break
		}
		if x == ex && y == ey {
			return ws.path(sx, sy)
		}
		set := ws.naturalDir(d) | ws.forceDir(x, y, d)
		set.dirIter(func(nd int32) bool {
			return !ws.jump(x, y, x, y, nd, c)
		})
	}
	return nil, false
}

func (ws *WorkSpace) jump(x, y, fx, fy, d, c int32) bool {
	for {
		x, y = Move(x, y, d)
		if !ws.Map.Available(x, y) {
			return false
		}
		if x == ws.endX && y == ws.endY {
			ws.putInOpenSet(x, y, d, fx, fy, c)
			return true
		}
		if ws.forceDir(x, y, d) > 0 {
			ws.putInOpenSet(x, y, d, fx, fy, c)
			return false
		}
		if spread(d) {
			if ws.jump(x, y, fx, fy, (d+1)%6, c) || ws.jump(x, y, fx, fy, (d+5)%6, c) {
				return true
			}
		}
	}
}

func (ws *WorkSpace) getOutOpenSet() (x, y, d, c int32, ok bool) {
	if ws.heap.Empty() {
		return
	}
	node := ws.heap.Pop()
	node.Status = grid.NodeClose
	return node.Pos.X, node.Pos.Y, node.Dir, node.Cost, true
}

func (ws *WorkSpace) putInOpenSet(x, y, d, fx, fy, c int32) {
	node := ws.pool.GetNode(x, y)
	if node == nil {
		return
	}
	switch node.Status {
	case grid.NodeNew:
		node.FPos = grid.Gpos{X: fx, Y: fy}
		node.Dir = d
		node.Cost = c + dist(x, y, fx, fy)
		node.Total = node.Cost + dist(x, y, ws.endX, ws.endY)
		node.Status = grid.NodeOpen
		ws.heap.Push(node)
	case grid.NodeOpen:
		cost := c + dist(x, y, fx, fy)
		if cost < node.Cost {
			node.FPos = grid.Gpos{X: fx, Y: fy}
			node.Dir = d
			node.Cost = cost
			node.Total = cost + dist(x, y, ws.endX, ws.endY)
			ws.heap.Fix(node)
		}
	case grid.NodeClose:
		return
	}
}

func (ws *WorkSpace) naturalDir(curDir int32) (s dirSet) {
	if curDir == _noDir {
		return _fullDirSetdirSet
	}
	s.dirAdd(curDir)
	if spread(curDir) {
		s.dirAdd((curDir + 1) % 6)
		s.dirAdd((curDir + 5) % 6)
	}

	return s
}

func (ws *WorkSpace) forceDir(x, y, curDir int32) (s dirSet) {
	if curDir == _noDir || spread(curDir) {
		return _emptyDirSet
	}
	if !ws.walkable(x, y, curDir, 2) {
		s.dirAdd((curDir + 1) % 6)
	}
	if !ws.walkable(x, y, curDir, 4) {
		s.dirAdd((curDir + 5) % 6)
	}
	return s
}

func (ws *WorkSpace) walkable(x, y, curDir, nextDir int32) bool {
	x, y = Move(x, y, (curDir+nextDir)%6)
	return ws.Map.Available(x, y)
}

func (ws *WorkSpace) path(sx, sy int32) (p []grid.PathGrid, ok bool) {
	var (
		fx, fy int32
		x, y   = ws.endX, ws.endY
		node   *grid.Gnode
	)
	for !(x == sx && y == sy) {
		p = append(p, grid.PathGrid{X: x, Y: y})
		if node = ws.pool.FindNode(x, y); node == nil {
			return
		}
		fx, fy = node.FPos.X, node.FPos.Y
		if mx, my, ok1 := midPoint(x, y, fx, fy); ok1 {
			p = append(p, grid.PathGrid{X: mx, Y: my})
		}
		x, y = fx, fy
	}
	p = append(p, grid.PathGrid{X: sx, Y: sy})
	slices.Reverse(p)
	return p, true
}

func Move(x, y, d int32) (int32, int32) {
	switch d {
	case 0:
		x++
	case 3:
		x--
	case 1, 5:
		x += y & 1
	case 2, 4:
		x -= 1 - y&1
	}
	switch d {
	case 1, 2:
		y--
	case 4, 5:
		y++
	}
	return x, y
}

func dist(x, y, fx, fy int32) int32 {
	q, r := xy2qr(x-fx, y-fy)
	return (abs(q) + abs(r) + abs(q+r)) / 2
}

func xy2qr(x, y int32) (q, r int32) {
	return x - (y-(y&1))/2, y
}
func qr2xy(q, r int32) (x, y int32) {
	return q + (r-(r&1))/2, r
}

func midPoint(x, y, fx, fy int32) (mx, my int32, ok bool) {
	var (
		q, r       = xy2qr(x, y)
		fq, fr     = xy2qr(fx, fy)
		s, fs      = -q - r, -fq - fr
		dq, dr, ds = q - fq, r - fr, s - fs
		mq, mr, ms int32
	)
	if dq == 0 || dr == 0 || ds == 0 {
		return // no midpoint
	}
	switch {
	case dr > 0 && dq > 0:
		mq, ms = fq+dq, fs-dq
		mr = -mq - ms
	case dr > 0 && ds > 0:
		mr, mq = fr+dr, fq-dr
		ms = -mq - mr
	case dr > 0:
		mq, mr = fq+dq, fr-dq
		ms = -mq - mr
	case dq < 0:
		mr, ms = fr+dr, fs-dr
		mq = -mr - ms
	case ds < 0:
		ms, mq = fs+ds, fq-ds
		mr = -mq - ms
	default:
		ms, mr = fs+ds, fr-ds
		mq = -mr - ms
	}
	mx, my = qr2xy(mq, mr)
	return mx, my, true
}

func spread(d int32) bool {
	return d%2 == 0
}

func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
