package grid

import (
	"fmt"
	"golang.org/x/exp/slices"
)

/*
N, NE, E, SE, S, SW, W, NW

	7  0  1		6  7  0
	6  *  2		5  *  1
	5  4  3		4  3  2
*/

type dirSet uint32

const (
	_fullDirSet  = 0xff
	_emptyDirSet = 0
	_noDir       = 8

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

func encodePos(x, y int32) int32 {
	return x<<16 + y
}
func decodePos(pos int32) (x, y int32) {
	return pos >> 16, pos & 0xffff
}

type PathNode struct {
	X, Y int32
}

func (w *WorkSpace) Dump() {
	for y := int32(47); y >= 0; y-- {
		for x := int32(0); x < 48; x++ {
			if w.IsSet(x, y) {
				fmt.Print("X")
			} else {
				fmt.Print("O")
			}
		}
		fmt.Println()
	}
}

func (w *WorkSpace) Solve(sx, sy, ex, ey int32) (path []PathNode, success bool) {
	w.pool.clear()
	w.queue.clear()
	w.endX = ex
	w.endY = ey
	w.putInOpenSet(sx, sy, _noDir, sx, sy, 0)
	for {
		x, y, d, c, ok := w.getOutOpenSet()
		if !ok {
			break
		}
		//fmt.Printf("%d %d c=%d\n", x, y, c) // debug
		//time.Sleep(time.Second)             // debug
		if x == ex && y == ey {
			path, success = w.genPath(sx, sy)
			return
		}
		dSet := w.naturalDir(x, y, d) | w.forceDir(x, y, d)
		dSet.dirIter(func(nd int32) bool {
			return !w.jump(x, y, nd, x, y, c)
		})
	}
	return nil, false
}

func (w *WorkSpace) genPath(sx, sy int32) (p []PathNode, ok bool) {
	x, y := w.endX, w.endY
	for !(x == sx && y == sy) {
		p = append(p, PathNode{X: x, Y: y})
		node := w.pool.findNode(encodePos(x, y))
		if node == nil {
			return
		}
		fx, fy := decodePos(node.fPos)
		if n, ok1 := midPoint(x, y, fx, fy); ok1 {
			p = append(p, n)
		}
		x, y = fx, fy
	}
	p = append(p, PathNode{X: sx, Y: sy})
	slices.Reverse(p)
	return p, true
}

func (w *WorkSpace) jump(x, y, d, fx, fy, c int32) (arrive bool) {
	for {
		x, y = move(x, y, d)
		if !inWorkSpace(x, y) || w.IsSet(x, y) { // path end
			return false
		}
		if avoidCorner && diagonal(d) {
			if !(w.walkable(x, y, d, 3) && w.walkable(x, y, d, 5)) {
				return false
			}
		}
		if x == w.endX && y == w.endY {
			w.putInOpenSet(x, y, d, fx, fy, c)
			return true
		}
		if w.forceDir(x, y, d) > 0 {
			w.putInOpenSet(x, y, d, fx, fy, c)
			return false
		}
		if diagonal(d) {
			if w.jump(x, y, (d+7)%8, fx, fy, c) {
				return true
			}
			if w.jump(x, y, (d+1)%8, fx, fy, c) {
				return true
			}
		}
	}
}

func (w *WorkSpace) getOutOpenSet() (x, y, d, c int32, ok bool) {
	if w.queue.empty() {
		return 0, 0, 0, 0, false
	}
	node := w.queue.pop()
	node.status = nodeClose
	x, y = decodePos(node.pos)
	return x, y, node.dir, node.cost, true
}

func (w *WorkSpace) putInOpenSet(x, y, d, fx, fy, c int32) {
	//fmt.Printf("push (%d %d) from (%d %d)\n", x, y, fx, fy) // debug
	pos := encodePos(x, y)
	node := w.pool.getNode(pos)
	switch node.status {
	case nodeNew:
		node.pos = pos
		node.fPos = encodePos(fx, fy)
		node.dir = d
		node.cost = c + dist(x, y, fx, fy)
		node.total = node.cost + dist(x, y, w.endX, w.endY)
		node.status = nodeOpen
		w.queue.push(node)
	case nodeOpen:
		cost := c + dist(x, y, fx, fy)
		if cost < node.cost {
			node.pos = pos
			node.fPos = encodePos(fx, fy)
			node.dir = d
			node.cost = c + dist(x, y, fx, fy)
			node.total = node.cost + dist(x, y, w.endX, w.endY)
			w.queue.fix(node)
		}
	case nodeClose:
		return
	}
}

func (w *WorkSpace) naturalDir(x, y, curDir int32) (s dirSet) {
	if curDir == _noDir {
		return _fullDirSet
	}
	if avoidCorner {
		if diagonal(curDir) {
			if !w.walkable(x, y, curDir, 7) {
				s.dirAdd((curDir + 1) % 8)
			} else if !w.walkable(x, y, curDir, 1) {
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
			s.dirAdd((curDir + 7) % 8)
			s.dirAdd((curDir + 1) % 8)
		}
	}
	return s
}

func (w *WorkSpace) forceDir(x, y, curDir int32) (s dirSet) {
	if curDir == _noDir {
		return _emptyDirSet
	}
	if avoidCorner {
		if !diagonal(curDir) {
			if w.walkable(x, y, curDir, 2) && !w.walkable(x, y, curDir, 3) {
				s.dirAdd((curDir + 2) % 8)
				s.dirAdd((curDir + 1) % 8)
			}
			if w.walkable(x, y, curDir, 6) && !w.walkable(x, y, curDir, 5) {
				s.dirAdd((curDir + 6) % 8)
				s.dirAdd((curDir + 7) % 8)
			}
		}
	} else {
		if diagonal(curDir) {
			if w.walkable(x, y, curDir, 6) && !w.walkable(x, y, curDir, 5) {
				s.dirAdd((curDir + 6) % 8)
			}
			if w.walkable(x, y, curDir, 2) && !w.walkable(x, y, curDir, 3) {
				s.dirAdd((curDir + 2) % 8)
			}
		} else {
			if w.walkable(x, y, curDir, 1) && !w.walkable(x, y, curDir, 2) {
				s.dirAdd((curDir + 1) % 8)
			}
			if w.walkable(x, y, curDir, 7) && !w.walkable(x, y, curDir, 6) {
				s.dirAdd((curDir + 7) % 8)
			}
		}
	}
	return
}

func (w *WorkSpace) Los(ax, ay, bx, by int32) bool {
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
			if w.IsSet(ax, ay+dy) || w.IsSet(ax+dx, ay) {
				return false
			}
			ax += dx
			ay += dy
			ix++
			iy++
		}
		if w.IsSet(ax, ay) {
			return false
		}
	}
	return true
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

func midPoint(x, y, fx, fy int32) (n PathNode, ok bool) {
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
	span := dx
	if dy < dx {
		span = dy
	}
	var mx, my int32
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
	return PathNode{X: mx, Y: my}, true
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
