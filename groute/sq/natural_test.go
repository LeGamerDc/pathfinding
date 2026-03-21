package sq

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/legamerdc/pathfinding/groute/grid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	demoSqRatio       = 0.277
	demoSqNx    int32 = 3
	demoSqNy    int32 = 3
)

func TestWorkSpace_SolveNatural_DirectLine(t *testing.T) {
	local := createTestGrid(10, 10)
	ws := NewWorkSpace(100)
	ws.Reset(local)

	path, ok := ws.SolveNatural(0.2, 0.2, 9.8, 9.8)
	require.True(t, ok)
	require.Len(t, path, 2)
	assert.Equal(t, grid.PathPoint{X: 0.2, Y: 0.2}, path[0])
	assert.Equal(t, grid.PathPoint{X: 9.8, Y: 9.8}, path[1])
}

func TestWorkSpace_SolveNatural_SameCell(t *testing.T) {
	local := createTestGrid(10, 10)
	ws := NewWorkSpace(100)
	ws.Reset(local)

	path, ok := ws.SolveNatural(3.2, 3.2, 3.8, 3.6)
	require.True(t, ok)
	require.Len(t, path, 2)
	assert.Equal(t, grid.PathPoint{X: 3.2, Y: 3.2}, path[0])
	assert.Equal(t, grid.PathPoint{X: 3.8, Y: 3.6}, path[1])
}

func TestWorkSpace_SolveNatural_WithObstacle(t *testing.T) {
	local := createTestGrid(5, 5)
	setObstacles(local, []grid.PathGrid{{X: 1, Y: 1}})

	ws := NewWorkSpace(64)
	ws.Reset(local)

	path, ok := ws.SolveNatural(0.2, 0.2, 2.8, 2.8)
	require.True(t, ok)
	assert.Greater(t, len(path), 2)
	assert.Equal(t, grid.PathPoint{X: 0.2, Y: 0.2}, path[0])
	assert.Equal(t, grid.PathPoint{X: 2.8, Y: 2.8}, path[len(path)-1])

	gridPath, ok := ws.Solve(0, 0, 2, 2)
	require.True(t, ok)
	corridor := buildPathCorridor(local, expandGridPath(gridPath))
	sawInset := false
	for i := 1; i < len(path); i++ {
		assert.True(t, segmentVisibleInCells(local, corridor, path[i-1], path[i]))
		assert.True(t, segmentVisibleInMap(local, path[i-1], path[i]))
	}
	for _, p := range path[1 : len(path)-1] {
		if !isIntegerCoord(p.X) || !isIntegerCoord(p.Y) {
			sawInset = true
		}
	}
	assert.True(t, sawInset)
}

func TestWorkSpace_SolveNatural_BlockedEndpoint(t *testing.T) {
	local := createTestGrid(5, 5)
	local.Set(1, 1)

	ws := NewWorkSpace(64)
	ws.Reset(local)

	_, ok := ws.SolveNatural(1.2, 1.2, 4.8, 4.8)
	assert.False(t, ok)
}

func TestBuildPathCorridor_AddsDiagonalSideCells(t *testing.T) {
	local := createTestGrid(3, 3)
	corridor := buildPathCorridor(local, []grid.PathGrid{
		{X: 0, Y: 0},
		{X: 1, Y: 1},
	})

	assert.True(t, corridor.has(0, 0))
	assert.True(t, corridor.has(1, 1))
	assert.True(t, corridor.has(1, 0))
	assert.True(t, corridor.has(0, 1))
}

func TestCorridorCornerPoints_LShape(t *testing.T) {
	corridor := make(cellSet)
	for _, c := range []grid.PathGrid{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 1, Y: 1},
	} {
		corridor.add(c.X, c.Y)
	}

	points := corridorCornerPoints(corridor, naturalMargin)
	require.Len(t, points, 6)
	assert.Contains(t, points, grid.PathPoint{X: 1.05, Y: 0.95})
	for _, p := range points {
		fracX := p.X - math.Floor(p.X)
		fracY := p.Y - math.Floor(p.Y)
		assert.True(t, nearlyEqual(fracX, naturalMargin) || nearlyEqual(fracX, 1-naturalMargin))
		assert.True(t, nearlyEqual(fracY, naturalMargin) || nearlyEqual(fracY, 1-naturalMargin))
	}
}

func TestSegmentVisibleInCells_RejectsLeavingCorridor(t *testing.T) {
	local := createTestGrid(3, 3)
	corridor := make(cellSet)
	for _, c := range []grid.PathGrid{
		{X: 0, Y: 0},
		{X: 1, Y: 1},
	} {
		corridor.add(c.X, c.Y)
	}

	assert.False(t, segmentVisibleInCells(
		local,
		corridor,
		grid.PathPoint{X: 0.2, Y: 0.2},
		grid.PathPoint{X: 1.8, Y: 1.8},
	))
}

func TestSegmentVisibleInMap_DiagonalAcrossObstacle(t *testing.T) {
	local := createTestGrid(3, 3)
	setObstacles(local, []grid.PathGrid{{X: 1, Y: 1}})

	assert.False(t, segmentVisibleInMap(
		local,
		grid.PathPoint{X: 0.2, Y: 0.2},
		grid.PathPoint{X: 2.8, Y: 2.8},
	))
	assert.True(t, segmentVisibleInMap(
		local,
		grid.PathPoint{X: 0.2, Y: 0.2},
		grid.PathPoint{X: 0.95, Y: 1.95},
	))
}

func TestSegmentVisibleInMap_DiagonalSlit(t *testing.T) {
	local := createTestGrid(3, 3)
	setObstacles(local, []grid.PathGrid{
		{X: 0, Y: 1},
		{X: 1, Y: 0},
	})

	assert.False(t, segmentVisibleInMap(
		local,
		grid.PathPoint{X: 0.2, Y: 0.2},
		grid.PathPoint{X: 1.8, Y: 1.8},
	))
}

func TestSegmentVisibleInMap_UsesReferenceVertexLOS(t *testing.T) {
	local := createTestGrid(4, 4)
	setObstacles(local, []grid.PathGrid{
		{X: 1, Y: 1},
		{X: 1, Y: 2},
	})

	assert.False(t, segmentVisibleInMap(
		local,
		grid.PathPoint{X: 1, Y: 2},
		grid.PathPoint{X: 3, Y: 2},
	))
}

func TestWorkSpace_SolveNatural_RandomDemoMaps(t *testing.T) {
	const seedCount = 256

	for seed := int64(0); seed < seedCount; seed++ {
		local := createDemoSquareMap(seed)
		ws := NewWorkSpace(1200)
		ws.Reset(local)

		start := grid.PathPoint{X: 0.2, Y: 0.2}
		end := grid.PathPoint{
			X: float64(demoSqNx*16) - 0.2,
			Y: float64(demoSqNy*16) - 0.2,
		}

		gridPath, gridOK := ws.Solve(0, 0, demoSqNx*16-1, demoSqNy*16-1)
		naturalPath, naturalOK := ws.SolveNatural(start.X, start.Y, end.X, end.Y)

		if !gridOK {
			if naturalOK {
				t.Fatalf("seed %d: natural path exists while grid path does not: %v", seed, naturalPath)
			}
			continue
		}

		if !naturalOK {
			t.Fatalf("seed %d: natural solver failed while grid solver succeeded", seed)
		}

		corridor := buildPathCorridor(local, expandGridPath(gridPath))
		validateNaturalPath(t, local, corridor, naturalPath, seed)
	}
}

func createDemoSquareMap(seed int64) *grid.Local {
	local := grid.NewLocal(demoSqNx, demoSqNy)
	for i := int32(0); i < demoSqNx; i++ {
		for j := int32(0); j < demoSqNy; j++ {
			local.SetGrid(i, j, &grid.Grid{})
		}
	}

	rng := rand.New(rand.NewSource(seed))
	for x := int32(0); x < demoSqNx*16; x++ {
		for y := int32(0); y < demoSqNy*16; y++ {
			if (x == 0 && y == 0) || (x == demoSqNx*16-1 && y == demoSqNy*16-1) {
				continue
			}
			if rng.Float32() < demoSqRatio {
				local.Set(x, y)
			}
		}
	}
	return local
}

func validateNaturalPath(t *testing.T, local *grid.Local, corridor cellSet, path []grid.PathPoint, seed int64) {
	t.Helper()

	require.GreaterOrEqual(t, len(path), 2, "seed %d: natural path must contain at least start/end", seed)
	for i := 1; i < len(path); i++ {
		require.Truef(
			t,
			segmentVisibleInCells(local, corridor, path[i-1], path[i]),
			"seed %d: segment %d leaves corridor: %s",
			seed,
			i-1,
			formatPath(path),
		)
		require.Truef(
			t,
			segmentVisibleInMap(local, path[i-1], path[i]),
			"seed %d: segment %d crosses blocked cells: %s",
			seed,
			i-1,
			formatPath(path),
		)
	}

	for _, issue := range findRedundantTurns(local, corridor, path) {
		t.Fatalf(
			"seed %d: natural path keeps a redundant turn at waypoint %d: %s",
			seed,
			issue,
			formatPath(path),
		)
	}
	for _, issue := range findVisibleZigZags(local, corridor, path) {
		t.Fatalf("seed %d: natural path contains a local zig-zag around waypoints %d-%d: %s", seed, issue, issue+1, formatPath(path))
	}
}

func findRedundantTurns(local *grid.Local, corridor cellSet, path []grid.PathPoint) []int {
	issues := make([]int, 0)
	for i := 1; i < len(path)-1; i++ {
		if collinear(path[i-1], path[i], path[i+1]) {
			continue
		}
		if segmentVisibleInCells(local, corridor, path[i-1], path[i+1]) &&
			segmentVisibleInMap(local, path[i-1], path[i+1]) {
			issues = append(issues, i)
		}
	}
	return issues
}

func findVisibleZigZags(local *grid.Local, corridor cellSet, path []grid.PathPoint) []int {
	issues := make([]int, 0)
	for i := 1; i+2 < len(path); i++ {
		turnA := turnOrientation(path[i-1], path[i], path[i+1])
		turnB := turnOrientation(path[i], path[i+1], path[i+2])
		if turnA == 0 || turnB == 0 || turnA == turnB {
			continue
		}
		if segmentVisibleInCells(local, corridor, path[i-1], path[i+2]) &&
			segmentVisibleInMap(local, path[i-1], path[i+2]) {
			issues = append(issues, i)
		}
	}
	return issues
}

func turnOrientation(a, b, c grid.PathPoint) int {
	cross := (b.X-a.X)*(c.Y-b.Y) - (b.Y-a.Y)*(c.X-b.X)
	switch {
	case cross > 1e-9:
		return 1
	case cross < -1e-9:
		return -1
	default:
		return 0
	}
}

func formatPath(path []grid.PathPoint) string {
	return fmt.Sprintf("%v", path)
}
