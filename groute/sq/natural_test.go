package sq

import (
	"math"
	"testing"

	"github.com/legamerdc/pathfinding/groute/grid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
