package sq

import (
	"slices"

	"github.com/legamerdc/pathfinding/groute/grid"
	"github.com/legamerdc/pathfinding/utils/heap"
)

/*
N, NE, E, SE, S, SW, W, NW

	7  0  1		6  7  0
	6  *  2		5  *  1
	5  4  3		4  3  2
*/

type dirSet uint32

const (
	_fullDirSetdirSet dirSet = (1 << 8) - 1
	_emptyDirSet      dirSet = 0

	_noDir = 0xff

	avoidCorner = true
)

func (s *dirSet) dirAdd(d int32) {
	*s |= 1 << d
}
func (s *dirSet) dirIter(f func(d int32) bool) {
	for d := int32(0); d < 8; d++ {
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
	heap       *heap.Heap[*grid.Gnode]
	endX, endY int32
}

func NewWorkSpace(size int) *WorkSpace {
	return &WorkSpace{
		pool: grid.NewNodePool(int32(size)),
		heap: heap.NewHeap[*grid.Gnode](size),
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
		set := ws.naturalDir(x, y, d) | ws.forceDir(x, y, d)
		set.dirIter(func(nd int32) bool {
			return !ws.jump(x, y, x, y, nd, c)
		})
	}
	return nil, false
}

func (ws *WorkSpace) jump(x, y, fx, fy, d, c int32) bool {
	for {
		x, y = move(x, y, d)
		if !ws.Map.Available(x, y) {
			return false
		}
		if avoidCorner && diagonal(d) {
			if !(ws.walkable(x, y, d, 3) && ws.walkable(x, y, d, 5)) {
				return false
			}
		}
		if x == ws.endX && y == ws.endY {
			ws.putInOpenSet(x, y, d, fx, fy, c)
			return true
		}
		if ws.forceDir(x, y, d) > 0 {
			ws.putInOpenSet(x, y, d, fx, fy, c)
			return false
		}
		if diagonal(d) {
			if ws.jump(x, y, fx, fy, (d+7)%8, c) {
				return true
			}
			if ws.jump(x, y, fx, fy, (d+1)%8, c) {
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

func (ws *WorkSpace) naturalDir(x, y, curDir int32) (s dirSet) {
	if curDir == _noDir {
		return _fullDirSetdirSet
	}
	if avoidCorner {
		if diagonal(curDir) {
			if !ws.walkable(x, y, curDir, 7) {
				s.dirAdd((curDir + 1) % 8)
			} else if !ws.walkable(x, y, curDir, 1) {
				s.dirAdd((curDir + 7) % 8)
			} else {
				s.dirAdd(curDir)
				s.dirAdd((curDir + 1) % 8)
				s.dirAdd((curDir + 7) % 8)
			}
		} else {
			s.dirAdd(curDir)
		}
	} else {
		s.dirAdd(curDir)
		if diagonal(curDir) {
			s.dirAdd((curDir + 1) % 8)
			s.dirAdd((curDir + 7) % 8)
		}
	}
	return s
}

func (ws *WorkSpace) forceDir(x, y, curDir int32) (s dirSet) {
	if curDir == _noDir {
		return _emptyDirSet
	}
	if avoidCorner {
		if !diagonal(curDir) {
			if ws.walkable(x, y, curDir, 2) && !ws.walkable(x, y, curDir, 3) {
				s.dirAdd((curDir + 2) % 8)
				s.dirAdd((curDir + 1) % 8)
			}
			if ws.walkable(x, y, curDir, 6) && !ws.walkable(x, y, curDir, 5) {
				s.dirAdd((curDir + 6) % 8)
				s.dirAdd((curDir + 7) % 8)
			}
		}
	} else {
		if diagonal(curDir) {
			if ws.walkable(x, y, curDir, 6) && !ws.walkable(x, y, curDir, 5) {
				s.dirAdd((curDir + 6) % 8)
			}
			if ws.walkable(x, y, curDir, 2) && !ws.walkable(x, y, curDir, 3) {
				s.dirAdd((curDir + 2) % 8)
			}
		} else {
			if ws.walkable(x, y, curDir, 1) && !ws.walkable(x, y, curDir, 2) {
				s.dirAdd((curDir + 1) % 8)
			}
			if ws.walkable(x, y, curDir, 7) && !ws.walkable(x, y, curDir, 6) {
				s.dirAdd((curDir + 7) % 8)
			}
		}
	}
	return
}

func (ws *WorkSpace) walkable(x, y, curDir, nextDir int32) bool {
	x, y = move(x, y, (curDir+nextDir)%8)
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

func move(x, y, d int32) (int32, int32) {
	switch d {
	case 1, 2, 3:
		x++
	case 5, 6, 7:
		x--
	}
	switch d {
	case 0, 1, 7:
		y++
	case 3, 4, 5:
		y--
	}
	return x, y
}

func diagonal(d int32) bool {
	return d&0x1 == 1
}

func dist(x1, y1, x2, y2 int32) int32 {
	dx, dy := x2-x1, y2-y1
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dy >= dx {
		return dx*7 + (dy-dx)*5
	} else {
		return dy*7 + (dx-dy)*5
	}
}

func midPoint(x, y, fx, fy int32) (mx, my int32, ok bool) {
	dx, dy := x-fx, y-fy
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dx == 0 || dy == 0 || dx == dy {
		return
	}
	span := min(dx, dy)
	switch {
	case x > fx && y > fy: // top-right
		mx, my = fx+span, fy+span
	case x > fx && y < fy: // bot-right
		mx, my = fx+span, fy-span
	case x < fx && y > fy: // top-left
		mx, my = fx-span, fy+span
	case x < fx && y < fy: // bot-left
		mx, my = fx-span, fy-span
	default:
		panic("unreachable")
	}
	return mx, my, true
}
