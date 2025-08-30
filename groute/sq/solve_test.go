package sq

import (
	"context"
	"testing"
	"time"

	"github.com/legamerdc/pathfinding/groute/grid"
)

// 创建一个简单的测试网格
func createTestGrid(width, height int32) *grid.Local {
	nx, ny := (width+15)/16, (height+15)/16
	local := grid.NewLocal(nx, ny)

	// 初始化所有网格为空（可通行）
	for i := int32(0); i < nx; i++ {
		for j := int32(0); j < ny; j++ {
			local.SetGrid(i, j, &grid.Grid{})
		}
	}
	return local
}

// 在网格中设置障碍物
func setObstacles(local *grid.Local, obstacles []grid.PathGrid) {
	for _, obs := range obstacles {
		local.Set(obs.X, obs.Y)
	}
}

// 带超时的路径查找辅助函数
func solveWithTimeout(t *testing.T, ws *WorkSpace, sx, sy, ex, ey int32) ([]grid.PathGrid, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	type result struct {
		path []grid.PathGrid
		ok   bool
	}

	ch := make(chan result, 1)
	go func() {
		path, ok := ws.Solve(sx, sy, ex, ey)
		ch <- result{path, ok}
	}()

	select {
	case res := <-ch:
		return res.path, res.ok
	case <-ctx.Done():
		t.Fatal("测试超时，可能存在无限循环")
		return nil, false
	}
}

func TestWorkSpace_BasicPathfinding(t *testing.T) {
	// 创建一个简单的10x10网格
	local := createTestGrid(10, 10)

	ws := NewWorkSpace(100)
	ws.Reset(local)

	// 测试简单的直线路径
	path, ok := solveWithTimeout(t, ws, 0, 0, 5, 0)
	if !ok {
		t.Fatal("应该找到路径")
	}
	if len(path) == 0 {
		t.Fatal("路径不应该为空")
	}

	// 验证起点和终点
	if path[0].X != 0 || path[0].Y != 0 {
		t.Errorf("起点错误：期望 (0,0)，得到 (%d,%d)", path[0].X, path[0].Y)
	}
	if path[len(path)-1].X != 5 || path[len(path)-1].Y != 0 {
		t.Errorf("终点错误：期望 (5,0)，得到 (%d,%d)", path[len(path)-1].X, path[len(path)-1].Y)
	}
}

func TestWorkSpace_NoPath(t *testing.T) {
	// 创建网格并设置障碍物墙，创建16x16网格以确保边界检查正确
	local := createTestGrid(16, 16)

	// 创建一道完整的墙阻断路径（从y=0到y=15）
	obstacles := make([]grid.PathGrid, 16)
	for i := 0; i < 16; i++ {
		obstacles[i] = grid.PathGrid{X: 5, Y: int32(i)}
	}
	setObstacles(local, obstacles)

	ws := NewWorkSpace(256)
	ws.Reset(local)

	// 尝试穿过墙
	_, ok := solveWithTimeout(t, ws, 0, 0, 15, 0)
	if ok {
		t.Fatal("不应该找到路径，因为被墙阻断")
	}
}

func TestWorkSpace_SameStartEnd(t *testing.T) {
	local := createTestGrid(10, 10)

	ws := NewWorkSpace(100)
	ws.Reset(local)

	// 起点和终点相同
	path, ok := solveWithTimeout(t, ws, 5, 5, 5, 5)
	if !ok {
		t.Fatal("起点和终点相同时应该返回成功")
	}
	if len(path) != 1 {
		t.Errorf("路径长度应该为1，得到 %d", len(path))
	}
}

func TestWorkSpace_DiagonalPath(t *testing.T) {
	local := createTestGrid(20, 20)

	ws := NewWorkSpace(400)
	ws.Reset(local)

	// 测试对角路径
	path, ok := solveWithTimeout(t, ws, 0, 0, 10, 10)
	if !ok {
		t.Fatal("应该找到对角路径")
	}

	// JPS应该生成较短的路径（相比A*）
	if len(path) == 0 {
		t.Fatal("路径不应该为空")
	}

	// 验证起点和终点
	if path[0].X != 0 || path[0].Y != 0 {
		t.Errorf("起点错误：期望 (0,0)，得到 (%d,%d)", path[0].X, path[0].Y)
	}
	if path[len(path)-1].X != 10 || path[len(path)-1].Y != 10 {
		t.Errorf("终点错误：期望 (10,10)，得到 (%d,%d)", path[len(path)-1].X, path[len(path)-1].Y)
	}
}

func TestWorkSpace_ComplexMaze(t *testing.T) {
	local := createTestGrid(20, 20)

	// 创建一个复杂的迷宫
	obstacles := []grid.PathGrid{
		// 创建一些障碍物形成迷宫
		{X: 2, Y: 1}, {X: 2, Y: 2}, {X: 2, Y: 3}, {X: 2, Y: 4}, {X: 2, Y: 5},
		{X: 4, Y: 3}, {X: 5, Y: 3}, {X: 6, Y: 3}, {X: 7, Y: 3}, {X: 8, Y: 3},
		{X: 8, Y: 1}, {X: 8, Y: 2}, {X: 8, Y: 4}, {X: 8, Y: 5}, {X: 8, Y: 6},
		{X: 10, Y: 2}, {X: 11, Y: 2}, {X: 12, Y: 2}, {X: 13, Y: 2}, {X: 14, Y: 2},
		{X: 12, Y: 0}, {X: 12, Y: 1}, {X: 12, Y: 3}, {X: 12, Y: 4}, {X: 12, Y: 5},
	}
	setObstacles(local, obstacles)

	ws := NewWorkSpace(400)
	ws.Reset(local)

	path, ok := solveWithTimeout(t, ws, 0, 0, 15, 5)
	if !ok {
		t.Fatal("应该找到穿过迷宫的路径")
	}

	// 验证路径的完整性
	if len(path) < 2 {
		t.Fatal("路径至少应该有2个点")
	}

	// 验证起点和终点
	if path[0].X != 0 || path[0].Y != 0 {
		t.Errorf("起点错误：期望 (0,0)，得到 (%d,%d)", path[0].X, path[0].Y)
	}
	if path[len(path)-1].X != 15 || path[len(path)-1].Y != 5 {
		t.Errorf("终点错误：期望 (15,5)，得到 (%d,%d)", path[len(path)-1].X, path[len(path)-1].Y)
	}
}

func TestWorkSpace_BoundaryCheck(t *testing.T) {
	local := createTestGrid(10, 10)

	ws := NewWorkSpace(100)
	ws.Reset(local)

	// 测试边界
	path, ok := solveWithTimeout(t, ws, 0, 0, 9, 9)
	if !ok {
		t.Fatal("应该找到到边界的路径")
	}

	// 验证起点和终点
	if path[0].X != 0 || path[0].Y != 0 {
		t.Errorf("起点错误：期望 (0,0)，得到 (%d,%d)", path[0].X, path[0].Y)
	}
	if path[len(path)-1].X != 9 || path[len(path)-1].Y != 9 {
		t.Errorf("终点错误：期望 (9,9)，得到 (%d,%d)", path[len(path)-1].X, path[len(path)-1].Y)
	}
}

func TestMove(t *testing.T) {
	tests := []struct {
		name                 string
		x, y, d              int32
		expectedX, expectedY int32
	}{
		{"Direction 0", 5, 5, 0, 5, 6}, // y++
		{"Direction 1", 5, 5, 1, 6, 6}, // x++, y++
		{"Direction 2", 5, 5, 2, 6, 5}, // x++
		{"Direction 3", 5, 5, 3, 6, 4}, // x++, y--
		{"Direction 4", 5, 5, 4, 5, 4}, // y--
		{"Direction 5", 5, 5, 5, 4, 4}, // x--, y--
		{"Direction 6", 5, 5, 6, 4, 5}, // x--
		{"Direction 7", 5, 5, 7, 4, 6}, // x--, y++
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := move(tt.x, tt.y, tt.d)
			if gotX != tt.expectedX || gotY != tt.expectedY {
				t.Errorf("Move(%d, %d, %d) = (%d, %d), 期望 (%d, %d)",
					tt.x, tt.y, tt.d, gotX, gotY, tt.expectedX, tt.expectedY)
			}
		})
	}
}

func TestDist(t *testing.T) {
	tests := []struct {
		name           string
		x1, y1, x2, y2 int32
		expectedDist   int32
	}{
		{"Same point", 0, 0, 0, 0, 0},
		{"Horizontal", 0, 0, 5, 0, 25},     // 5 * 5
		{"Vertical", 0, 0, 0, 5, 25},       // 5 * 5
		{"Diagonal equal", 0, 0, 3, 3, 21}, // 3 * 7
		{"Mixed", 0, 0, 2, 5, 29},          // 2*7 + 3*5 = 14 + 15 = 29
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dist(tt.x1, tt.y1, tt.x2, tt.y2)
			if got != tt.expectedDist {
				t.Errorf("dist(%d, %d, %d, %d) = %d, 期望 %d",
					tt.x1, tt.y1, tt.x2, tt.y2, got, tt.expectedDist)
			}
		})
	}
}

func TestDiagonal(t *testing.T) {
	tests := []struct {
		direction  int32
		isDiagonal bool
	}{
		{0, false}, // N
		{1, true},  // NE
		{2, false}, // E
		{3, true},  // SE
		{4, false}, // S
		{5, true},  // SW
		{6, false}, // W
		{7, true},  // NW
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := diagonal(tt.direction)
			if got != tt.isDiagonal {
				t.Errorf("diagonal(%d) = %v, 期望 %v", tt.direction, got, tt.isDiagonal)
			}
		})
	}
}

func TestDirSet(t *testing.T) {
	var s dirSet

	// 测试添加方向
	s.dirAdd(0)
	s.dirAdd(2)
	s.dirAdd(4)

	// 测试迭代
	var directions []int32
	s.dirIter(func(d int32) bool {
		directions = append(directions, d)
		return true
	})

	expected := []int32{0, 2, 4}
	if len(directions) != len(expected) {
		t.Errorf("方向数量错误，期望 %d，得到 %d", len(expected), len(directions))
	}

	for i, d := range expected {
		if i >= len(directions) || directions[i] != d {
			t.Errorf("方向 %d：期望 %d，得到 %d", i, d, directions[i])
		}
	}
}

// 测试提前终止迭代
func TestDirSet_EarlyTermination(t *testing.T) {
	var s dirSet
	s.dirAdd(0)
	s.dirAdd(1)
	s.dirAdd(2)

	var count int
	s.dirIter(func(d int32) bool {
		count++
		return count < 2 // 只处理前2个方向
	})

	if count != 2 {
		t.Errorf("应该只处理2个方向，实际处理了 %d 个", count)
	}
}

// ==================== 基准测试 ====================

// 带超时的基准测试辅助函数
func benchSolveWithTimeout(b *testing.B, ws *WorkSpace, sx, sy, ex, ey int32) ([]grid.PathGrid, bool) {
	start := time.Now()
	path, ok := ws.Solve(sx, sy, ex, ey)
	if time.Since(start) > 2*time.Second {
		b.Fatal("基准测试超时，可能存在无限循环")
	}
	return path, ok
}

// 基准测试：简单直线路径
func BenchmarkWorkSpace_SimplePath(b *testing.B) {
	local := createTestGrid(50, 50)
	ws := NewWorkSpace(2500)
	ws.Reset(local)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := benchSolveWithTimeout(b, ws, 0, 0, 49, 0)
		if !ok {
			b.Fatal("应该找到路径")
		}
	}
}

// 基准测试：对角线路径
func BenchmarkWorkSpace_DiagonalPath(b *testing.B) {
	local := createTestGrid(50, 50)
	ws := NewWorkSpace(2500)
	ws.Reset(local)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := benchSolveWithTimeout(b, ws, 0, 0, 49, 49)
		if !ok {
			b.Fatal("应该找到路径")
		}
	}
}

// 基准测试：复杂迷宫
func BenchmarkWorkSpace_ComplexMaze(b *testing.B) {
	local := createTestGrid(100, 100)

	// 创建复杂迷宫
	obstacles := make([]grid.PathGrid, 0, 1000)

	// 添加随机障碍物
	for i := int32(10); i < 90; i += 5 {
		for j := int32(10); j < 90; j += 3 {
			if (i+j)%7 != 0 { // 留出通道
				obstacles = append(obstacles, grid.PathGrid{X: i, Y: j})
				obstacles = append(obstacles, grid.PathGrid{X: i + 1, Y: j})
				obstacles = append(obstacles, grid.PathGrid{X: i, Y: j + 1})
			}
		}
	}

	// 创建几道墙，但留出缺口
	for i := int32(20); i < 80; i++ {
		if i%15 != 0 { // 每15个单位留一个缺口
			obstacles = append(obstacles, grid.PathGrid{X: i, Y: 25})
			obstacles = append(obstacles, grid.PathGrid{X: i, Y: 50})
			obstacles = append(obstacles, grid.PathGrid{X: i, Y: 75})
		}
	}

	setObstacles(local, obstacles)

	ws := NewWorkSpace(10000)
	ws.Reset(local)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := benchSolveWithTimeout(b, ws, 5, 5, 95, 95)
		if !ok {
			b.Fatal("应该找到路径")
		}
	}
}

// 基准测试：大型网格
func BenchmarkWorkSpace_LargeGrid(b *testing.B) {
	local := createTestGrid(200, 200)
	ws := NewWorkSpace(40000)
	ws.Reset(local)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := benchSolveWithTimeout(b, ws, 0, 0, 199, 199)
		if !ok {
			b.Fatal("应该找到路径")
		}
	}
}

// 基准测试：多次短路径搜索
func BenchmarkWorkSpace_MultipleShortPaths(b *testing.B) {
	local := createTestGrid(20, 20)
	ws := NewWorkSpace(400)
	ws.Reset(local)

	paths := [][4]int32{
		{0, 0, 5, 5},
		{1, 1, 8, 8},
		{2, 2, 10, 3},
		{3, 5, 7, 15},
		{5, 0, 15, 10},
		{10, 10, 19, 19},
		{0, 19, 19, 0},
		{10, 0, 0, 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		path := paths[i%len(paths)]
		benchSolveWithTimeout(b, ws, path[0], path[1], path[2], path[3])
	}
}

// 基准测试：无解路径（测试失败情况的性能）
func BenchmarkWorkSpace_NoSolution(b *testing.B) {
	local := createTestGrid(64, 64)

	// 创建完全分隔的两个区域（完整的墙从y=0到y=63）
	obstacles := make([]grid.PathGrid, 64)
	for i := 0; i < 64; i++ {
		obstacles[i] = grid.PathGrid{X: 32, Y: int32(i)}
	}
	setObstacles(local, obstacles)

	ws := NewWorkSpace(4096)
	ws.Reset(local)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, ok := benchSolveWithTimeout(b, ws, 0, 0, 63, 63)
		if ok {
			b.Fatal("不应该找到路径")
		}
	}
}

// 基准测试：Move函数性能
func BenchmarkMove(b *testing.B) {
	x, y := int32(100), int32(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for d := int32(0); d < 8; d++ {
			move(x, y, d)
		}
	}
}

// 基准测试：距离计算函数
func BenchmarkDist(b *testing.B) {
	coords := [][4]int32{
		{0, 0, 10, 10},
		{5, 3, 15, 20},
		{100, 50, 200, 150},
		{0, 0, 1000, 1000},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coord := coords[i%len(coords)]
		dist(coord[0], coord[1], coord[2], coord[3])
	}
}

// 基准测试：节点池操作
func BenchmarkNodePool_Operations(b *testing.B) {
	pool := grid.NewNodePool(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 清空池
		pool.Clear()

		// 获取一些节点
		for j := int32(0); j < 100; j++ {
			node := pool.GetNode(j%10, j/10)
			if node == nil {
				b.Fatal("应该获取到节点")
			}
		}

		// 查找节点
		for j := int32(0); j < 50; j++ {
			pool.FindNode(j%10, j/10)
		}
	}
}

// 基准测试：优先队列操作
func BenchmarkNodeQueue_Operations(b *testing.B) {
	queue := grid.NewNodeQueue(1000)
	pool := grid.NewNodePool(1000)

	// 预创建一些节点
	nodes := make([]*grid.Gnode, 100)
	for i := 0; i < 100; i++ {
		nodes[i] = pool.GetNode(int32(i%10), int32(i/10))
		nodes[i].Total = int32(i * 3) // 设置优先级
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 清空队列
		queue.Clear()

		// 添加节点
		for j := 0; j < 50; j++ {
			queue.Push(nodes[j])
		}

		// 弹出一些节点
		for j := 0; j < 25; j++ {
			if queue.Empty() {
				break
			}
			queue.Pop()
		}

		// 修改优先级
		for j := 25; j < 50; j++ {
			if queue.Empty() {
				break
			}
			nodes[j].Total = int32(j * 2)
			queue.Fix(nodes[j])
		}
	}
}
