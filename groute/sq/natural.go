package sq

import (
	"container/heap"
	"math"
	"sort"

	"github.com/legamerdc/pathfinding/groute/grid"
)

const naturalMargin = 0.05

// SolveNatural returns a continuous path in grid-space.
//
// Coordinates use cell-space semantics: cell (x, y) covers
// [x, x+1) x [y, y+1), and the center of cell (x, y) is
// (float64(x)+0.5, float64(y)+0.5).
//
// The result is constrained to a corridor derived from the JPS path. This
// keeps the natural path as a post-process of the discrete solver rather than
// an independent full-map planner.
func (ws *WorkSpace) SolveNatural(sx, sy, ex, ey float64) ([]grid.PathPoint, bool) {
	if ws.Map == nil {
		return nil, false
	}

	startCellX, startCellY, ok := pointToWalkableGrid(ws.Map, sx, sy)
	if !ok {
		return nil, false
	}
	endCellX, endCellY, ok := pointToWalkableGrid(ws.Map, ex, ey)
	if !ok {
		return nil, false
	}

	gridPath, ok := ws.Solve(startCellX, startCellY, endCellX, endCellY)
	if !ok {
		return nil, false
	}

	corridor := buildPathCorridor(ws.Map, expandGridPath(gridPath))
	start := grid.PathPoint{X: sx, Y: sy}
	end := grid.PathPoint{X: ex, Y: ey}
	if segmentVisibleInCells(ws.Map, corridor, start, end) {
		return []grid.PathPoint{start, end}, true
	}

	nodes := make([]grid.PathPoint, 0, 2+len(corridor))
	nodes = append(nodes, start, end)
	nodes = append(nodes, corridorCornerPoints(corridor, naturalMargin)...)

	path, ok := shortestVisiblePathInCells(ws.Map, corridor, nodes)
	if !ok {
		return nil, false
	}
	return compressNaturalPath(path), true
}

type cellSet map[grid.Gpos]struct{}

func (s cellSet) add(x, y int32) {
	s[grid.Gpos{X: x, Y: y}] = struct{}{}
}

func (s cellSet) has(x, y int32) bool {
	_, ok := s[grid.Gpos{X: x, Y: y}]
	return ok
}

type shortestState struct {
	index int
	cost  float64
}

type shortestHeap []shortestState

func buildPathCorridor(m *grid.Local, path []grid.PathGrid) cellSet {
	cells := make(cellSet, len(path)*3)
	for _, p := range path {
		if cellInsideMap(m, p.X, p.Y) && m.Available(p.X, p.Y) {
			cells.add(p.X, p.Y)
		}
	}
	for i := 1; i < len(path); i++ {
		prev := path[i-1]
		cur := path[i]
		dx := cur.X - prev.X
		dy := cur.Y - prev.Y
		if dx == 0 || dy == 0 {
			continue
		}
		sideA := grid.Gpos{X: prev.X + dx, Y: prev.Y}
		sideB := grid.Gpos{X: prev.X, Y: prev.Y + dy}
		if cellInsideMap(m, sideA.X, sideA.Y) && m.Available(sideA.X, sideA.Y) {
			cells.add(sideA.X, sideA.Y)
		}
		if cellInsideMap(m, sideB.X, sideB.Y) && m.Available(sideB.X, sideB.Y) {
			cells.add(sideB.X, sideB.Y)
		}
	}
	return cells
}

type boundaryDir uint8

const (
	dirLeft boundaryDir = 1 << iota
	dirRight
	dirDown
	dirUp
)

func corridorCornerPoints(cells cellSet, margin float64) []grid.PathPoint {
	dirs := make(map[grid.Gpos]boundaryDir, len(cells)*2)
	addDir := func(x, y int32, dir boundaryDir) {
		p := grid.Gpos{X: x, Y: y}
		dirs[p] |= dir
	}

	for c := range cells {
		x, y := c.X, c.Y
		if !cells.has(x-1, y) {
			addDir(x, y, dirUp)
			addDir(x, y+1, dirDown)
		}
		if !cells.has(x+1, y) {
			addDir(x+1, y, dirUp)
			addDir(x+1, y+1, dirDown)
		}
		if !cells.has(x, y-1) {
			addDir(x, y, dirRight)
			addDir(x+1, y, dirLeft)
		}
		if !cells.has(x, y+1) {
			addDir(x, y+1, dirRight)
			addDir(x+1, y+1, dirLeft)
		}
	}

	points := make([]grid.PathPoint, 0, len(dirs))
	for p, mask := range dirs {
		if mask == dirLeft|dirRight || mask == dirUp|dirDown {
			continue
		}
		points = append(points, insetCornerPoint(cells, p, mask, margin))
	}
	sort.Slice(points, func(i, j int) bool {
		if points[i].X == points[j].X {
			return points[i].Y < points[j].Y
		}
		return points[i].X < points[j].X
	})
	return points
}

func insetCornerPoint(cells cellSet, p grid.Gpos, mask boundaryDir, margin float64) grid.PathPoint {
	type quadrant struct {
		open bool
		dx   float64
		dy   float64
	}

	quads := []quadrant{
		{open: cells.has(p.X-1, p.Y-1), dx: -margin, dy: -margin},
		{open: cells.has(p.X, p.Y-1), dx: margin, dy: -margin},
		{open: cells.has(p.X-1, p.Y), dx: -margin, dy: margin},
		{open: cells.has(p.X, p.Y), dx: margin, dy: margin},
	}

	openCount := 0
	for _, q := range quads {
		if q.open {
			openCount++
		}
	}

	switch openCount {
	case 1:
		for _, q := range quads {
			if q.open {
				return grid.PathPoint{X: float64(p.X) + q.dx, Y: float64(p.Y) + q.dy}
			}
		}
	case 3:
		for _, q := range quads {
			if !q.open {
				return grid.PathPoint{X: float64(p.X) - q.dx, Y: float64(p.Y) - q.dy}
			}
		}
	}

	x := float64(p.X)
	y := float64(p.Y)
	if mask&dirUp != 0 {
		x += margin
	}
	if mask&dirDown != 0 {
		x -= margin
	}
	if mask&dirRight != 0 {
		y += margin
	}
	if mask&dirLeft != 0 {
		y -= margin
	}
	return grid.PathPoint{X: x, Y: y}
}

func shortestVisiblePathInCells(m *grid.Local, cells cellSet, nodes []grid.PathPoint) ([]grid.PathPoint, bool) {
	const eps = 1e-9

	nodes = dedupePoints(nodes)
	visible := make([][]int, len(nodes))
	for i := 0; i < len(nodes); i++ {
		for j := i + 1; j < len(nodes); j++ {
			if segmentVisibleInCells(m, cells, nodes[i], nodes[j]) {
				visible[i] = append(visible[i], j)
				visible[j] = append(visible[j], i)
			}
		}
	}

	dist := make([]float64, len(nodes))
	prev := make([]int, len(nodes))
	for i := range dist {
		dist[i] = math.Inf(1)
		prev[i] = -1
	}
	dist[0] = 0

	pq := shortestHeap{{index: 0, cost: 0}}
	heap.Init(&pq)
	for pq.Len() > 0 {
		cur := heap.Pop(&pq).(shortestState)
		if cur.cost > dist[cur.index]+eps {
			continue
		}
		if cur.index == 1 {
			break
		}
		for _, next := range visible[cur.index] {
			alt := cur.cost + pointDistance(nodes[cur.index], nodes[next])
			if alt+eps < dist[next] {
				dist[next] = alt
				prev[next] = cur.index
				heap.Push(&pq, shortestState{index: next, cost: alt})
			}
		}
	}

	if math.IsInf(dist[1], 1) {
		return nil, false
	}

	path := make([]grid.PathPoint, 0, len(nodes))
	for at := 1; at != -1; at = prev[at] {
		path = append(path, nodes[at])
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path, true
}

func segmentVisibleInCells(m *grid.Local, cells cellSet, a, b grid.PathPoint) bool {
	return segmentVisibleWith(
		m.Nx*16,
		m.Ny*16,
		func(x, y int32) bool { return cells.has(x, y) },
		a,
		b,
	)
}

func segmentVisibleInMap(m *grid.Local, a, b grid.PathPoint) bool {
	return segmentVisibleWith(m.Nx*16, m.Ny*16, m.Available, a, b)
}

func segmentVisibleWith(width, height int32, open func(int32, int32) bool, a, b grid.PathPoint) bool {
	if !pointInOpenSpace(width, height, open, a.X, a.Y) || !pointInOpenSpace(width, height, open, b.X, b.Y) {
		return false
	}
	if nearlyEqual(a.X, b.X) && nearlyEqual(a.Y, b.Y) {
		return true
	}
	if nearlyEqual(a.X, b.X) && isIntegerCoord(a.X) {
		return verticalBoundaryVisible(width, height, open, int32(math.Round(a.X)), a.Y, b.Y)
	}
	if nearlyEqual(a.Y, b.Y) && isIntegerCoord(a.Y) {
		return horizontalBoundaryVisible(width, height, open, int32(math.Round(a.Y)), a.X, b.X)
	}

	ts := segmentBreakpoints(a, b)
	cells := make([]grid.Gpos, 0, len(ts)-1)
	for i := 1; i < len(ts); i++ {
		tm := 0.5 * (ts[i-1] + ts[i])
		x, y := segmentPoint(a, b, tm)
		cell, ok := interiorCell(width, height, x, y, b.X-a.X, b.Y-a.Y)
		if !ok || !open(cell.X, cell.Y) {
			return false
		}
		cells = append(cells, cell)
	}

	for i := 1; i < len(ts)-1; i++ {
		x, y := segmentPoint(a, b, ts[i])
		if !isIntegerCoord(x) || !isIntegerCoord(y) {
			continue
		}
		before := cells[i-1]
		after := cells[i]
		if before == after || before.X == after.X || before.Y == after.Y {
			continue
		}
		sideA := grid.Gpos{X: before.X, Y: after.Y}
		sideB := grid.Gpos{X: after.X, Y: before.Y}
		if !openCell(width, height, open, sideA.X, sideA.Y) && !openCell(width, height, open, sideB.X, sideB.Y) {
			return false
		}
	}

	return segmentHasMargin(width, height, open, a, b, naturalMargin)
}

func verticalBoundaryVisible(width, height int32, open func(int32, int32) bool, x int32, y1, y2 float64) bool {
	breaks := axisBreakpoints(y1, y2)
	for i := 1; i < len(breaks); i++ {
		ym := 0.5 * (breaks[i-1] + breaks[i])
		row := int32(math.Floor(ym))
		leftOpen := openCell(width, height, open, x-1, row)
		rightOpen := openCell(width, height, open, x, row)
		if !leftOpen && !rightOpen {
			return false
		}
	}

	for i := 1; i < len(breaks)-1; i++ {
		vy := int32(math.Round(breaks[i]))
		leftBelow := openCell(width, height, open, x-1, vy-1)
		rightBelow := openCell(width, height, open, x, vy-1)
		leftAbove := openCell(width, height, open, x-1, vy)
		rightAbove := openCell(width, height, open, x, vy)

		if leftBelow && rightAbove && !rightBelow && !leftAbove {
			return false
		}
		if rightBelow && leftAbove && !leftBelow && !rightAbove {
			return false
		}
	}
	return true
}

func horizontalBoundaryVisible(width, height int32, open func(int32, int32) bool, y int32, x1, x2 float64) bool {
	breaks := axisBreakpoints(x1, x2)
	for i := 1; i < len(breaks); i++ {
		xm := 0.5 * (breaks[i-1] + breaks[i])
		col := int32(math.Floor(xm))
		bottomOpen := openCell(width, height, open, col, y-1)
		topOpen := openCell(width, height, open, col, y)
		if !bottomOpen && !topOpen {
			return false
		}
	}

	for i := 1; i < len(breaks)-1; i++ {
		vx := int32(math.Round(breaks[i]))
		leftBelow := openCell(width, height, open, vx-1, y-1)
		leftAbove := openCell(width, height, open, vx-1, y)
		rightBelow := openCell(width, height, open, vx, y-1)
		rightAbove := openCell(width, height, open, vx, y)

		if leftBelow && rightAbove && !leftAbove && !rightBelow {
			return false
		}
		if leftAbove && rightBelow && !leftBelow && !rightAbove {
			return false
		}
	}
	return true
}

func axisBreakpoints(a, b float64) []float64 {
	lo, hi := a, b
	if lo > hi {
		lo, hi = hi, lo
	}
	points := []float64{lo, hi}
	start := int32(math.Ceil(lo))
	end := int32(math.Floor(hi))
	for v := start; v <= end; v++ {
		fv := float64(v)
		if fv > lo && fv < hi {
			points = append(points, fv)
		}
	}
	sort.Float64s(points)
	return dedupeFloats(points)
}

func segmentHasMargin(width, height int32, open func(int32, int32) bool, a, b grid.PathPoint, margin float64) bool {
	if margin <= 0 {
		return true
	}

	minX := clamp32(int32(math.Floor(minFloat(a.X, b.X)-margin))-1, 0, width-1)
	maxX := clamp32(int32(math.Ceil(maxFloat(a.X, b.X)+margin))+1, 0, width-1)
	minY := clamp32(int32(math.Floor(minFloat(a.Y, b.Y)-margin))-1, 0, height-1)
	maxY := clamp32(int32(math.Ceil(maxFloat(a.Y, b.Y)+margin))+1, 0, height-1)

	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			if open(x, y) {
				continue
			}
			if segmentHitsExpandedCell(a, b, x, y, margin) {
				return false
			}
		}
	}
	return true
}

func segmentHitsExpandedCell(a, b grid.PathPoint, x, y int32, margin float64) bool {
	t0, t1, ok := segmentRectIntersection(
		a,
		b,
		float64(x)-margin,
		float64(x+1)+margin,
		float64(y)-margin,
		float64(y+1)+margin,
	)
	if !ok {
		return false
	}
	return t1 > 1e-9 && t0 < 1-1e-9
}

func segmentRectIntersection(a, b grid.PathPoint, minX, maxX, minY, maxY float64) (float64, float64, bool) {
	t0, t1 := 0.0, 1.0
	dx := b.X - a.X
	dy := b.Y - a.Y
	if !clipAxis(a.X, dx, minX, maxX, &t0, &t1) {
		return 0, 0, false
	}
	if !clipAxis(a.Y, dy, minY, maxY, &t0, &t1) {
		return 0, 0, false
	}
	return t0, t1, t0 <= t1+1e-12
}

func clipAxis(origin, delta, minV, maxV float64, t0, t1 *float64) bool {
	const eps = 1e-12

	if math.Abs(delta) <= eps {
		return origin >= minV-eps && origin <= maxV+eps
	}

	a := (minV - origin) / delta
	b := (maxV - origin) / delta
	if a > b {
		a, b = b, a
	}
	if a > *t0 {
		*t0 = a
	}
	if b < *t1 {
		*t1 = b
	}
	return *t0 <= *t1+eps
}

func segmentBreakpoints(a, b grid.PathPoint) []float64 {
	points := []float64{0, 1}
	dx := b.X - a.X
	dy := b.Y - a.Y

	if !nearlyZero(dx) {
		lo := minFloat(a.X, b.X)
		hi := maxFloat(a.X, b.X)
		start := int32(math.Ceil(lo))
		end := int32(math.Floor(hi))
		for v := start; v <= end; v++ {
			fv := float64(v)
			if fv <= lo || fv >= hi {
				continue
			}
			t := (fv - a.X) / dx
			if t > 0 && t < 1 {
				points = append(points, t)
			}
		}
	}
	if !nearlyZero(dy) {
		lo := minFloat(a.Y, b.Y)
		hi := maxFloat(a.Y, b.Y)
		start := int32(math.Ceil(lo))
		end := int32(math.Floor(hi))
		for v := start; v <= end; v++ {
			fv := float64(v)
			if fv <= lo || fv >= hi {
				continue
			}
			t := (fv - a.Y) / dy
			if t > 0 && t < 1 {
				points = append(points, t)
			}
		}
	}

	sort.Float64s(points)
	return dedupeFloats(points)
}

func segmentPoint(a, b grid.PathPoint, t float64) (float64, float64) {
	return a.X + (b.X-a.X)*t, a.Y + (b.Y-a.Y)*t
}

func interiorCell(width, height int32, x, y, dx, dy float64) (grid.Gpos, bool) {
	if isIntegerCoord(x) {
		x = math.Nextafter(x, x+math.Copysign(1, dx))
	}
	if isIntegerCoord(y) {
		y = math.Nextafter(y, y+math.Copysign(1, dy))
	}
	cx := int32(math.Floor(x))
	cy := int32(math.Floor(y))
	if !cellInsideBounds(width, height, cx, cy) {
		return grid.Gpos{}, false
	}
	return grid.Gpos{X: cx, Y: cy}, true
}

func pointInOpenSpace(width, height int32, open func(int32, int32) bool, x, y float64) bool {
	if math.IsNaN(x) || math.IsNaN(y) || math.IsInf(x, 0) || math.IsInf(y, 0) {
		return false
	}
	if x < 0 || y < 0 || x > float64(width) || y > float64(height) {
		return false
	}

	anyOpen := false
	for _, c := range pointAdjacentCells(x, y) {
		if !cellInsideBounds(width, height, c.X, c.Y) {
			continue
		}
		if open(c.X, c.Y) {
			anyOpen = true
			continue
		}
		if x > float64(c.X) && x < float64(c.X+1) && y > float64(c.Y) && y < float64(c.Y+1) {
			return false
		}
	}
	return anyOpen
}

func pointToWalkableGrid(m *grid.Local, x, y float64) (gx, gy int32, ok bool) {
	if !pointInOpenSpace(m.Nx*16, m.Ny*16, m.Available, x, y) {
		return 0, 0, false
	}
	for _, c := range pointAdjacentCells(x, y) {
		if cellInsideMap(m, c.X, c.Y) && m.Available(c.X, c.Y) {
			return c.X, c.Y, true
		}
	}
	return 0, 0, false
}

func pointAdjacentCells(x, y float64) []grid.Gpos {
	fx := int32(math.Floor(x))
	fy := int32(math.Floor(y))
	xs := []int32{fx}
	ys := []int32{fy}

	if isIntegerCoord(x) {
		xs = append(xs, fx-1)
	}
	if isIntegerCoord(y) {
		ys = append(ys, fy-1)
	}

	cells := make([]grid.Gpos, 0, len(xs)*len(ys))
	seen := make(map[grid.Gpos]struct{}, len(xs)*len(ys))
	for _, cx := range xs {
		for _, cy := range ys {
			p := grid.Gpos{X: cx, Y: cy}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			cells = append(cells, p)
		}
	}
	return cells
}

func openCell(width, height int32, open func(int32, int32) bool, x, y int32) bool {
	return cellInsideBounds(width, height, x, y) && open(x, y)
}

func cellInsideMap(m *grid.Local, x, y int32) bool {
	return cellInsideBounds(m.Nx*16, m.Ny*16, x, y)
}

func cellInsideBounds(width, height, x, y int32) bool {
	return uint32(x) < uint32(width) && uint32(y) < uint32(height)
}

func expandGridPath(path []grid.PathGrid) []grid.PathGrid {
	if len(path) <= 1 {
		return append([]grid.PathGrid(nil), path...)
	}

	out := make([]grid.PathGrid, 0, len(path))
	out = append(out, path[0])
	for i := 1; i < len(path); i++ {
		x, y := path[i-1].X, path[i-1].Y
		tx, ty := path[i].X, path[i].Y
		dx, dy := sign32(tx-x), sign32(ty-y)
		for x != tx || y != ty {
			if x != tx {
				x += dx
			}
			if y != ty {
				y += dy
			}
			out = append(out, grid.PathGrid{X: x, Y: y})
		}
	}
	return out
}

func compressNaturalPath(path []grid.PathPoint) []grid.PathPoint {
	if len(path) < 3 {
		return path
	}
	out := make([]grid.PathPoint, 0, len(path))
	out = append(out, path[0])
	for i := 1; i < len(path)-1; i++ {
		prev := out[len(out)-1]
		cur := path[i]
		next := path[i+1]
		if collinear(prev, cur, next) {
			continue
		}
		out = append(out, cur)
	}
	out = append(out, path[len(path)-1])
	return out
}

func collinear(a, b, c grid.PathPoint) bool {
	const eps = 1e-9
	return math.Abs((b.X-a.X)*(c.Y-a.Y)-(b.Y-a.Y)*(c.X-a.X)) <= eps
}

func dedupePoints(points []grid.PathPoint) []grid.PathPoint {
	out := make([]grid.PathPoint, 0, len(points))
	seen := make(map[grid.PathPoint]struct{}, len(points))
	for _, p := range points {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func dedupeFloats(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	out := values[:1]
	for _, v := range values[1:] {
		if nearlyEqual(v, out[len(out)-1]) {
			continue
		}
		out = append(out, v)
	}
	return out
}

func pointDistance(a, b grid.PathPoint) float64 {
	return math.Hypot(a.X-b.X, a.Y-b.Y)
}

func isIntegerCoord(v float64) bool {
	return nearlyEqual(v, math.Round(v))
}

func nearlyEqual(a, b float64) bool {
	const eps = 1e-9
	return math.Abs(a-b) <= eps
}

func nearlyZero(v float64) bool {
	return nearlyEqual(v, 0)
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func clamp32(v, lo, hi int32) int32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func sign32(v int32) int32 {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	default:
		return 0
	}
}

func (h shortestHeap) Len() int { return len(h) }

func (h shortestHeap) Less(i, j int) bool { return h[i].cost < h[j].cost }

func (h shortestHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *shortestHeap) Push(x any) { *h = append(*h, x.(shortestState)) }

func (h *shortestHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}
