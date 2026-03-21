这里实现一些游戏中常用的寻路算法，当前主要包含：

- `groute/hex`: 正六边形网格 JPS(Jump Point Search)
- `groute/sq`: 正方形网格 JPS
- `groute/sq`: 额外提供 `SolveNatural`，用于生成更自然的连续路径

## 快速开始

```bash
go test ./...
go run ./demo/groute
```

执行 `go run ./demo/groute` 后，会在仓库根目录生成：

- `out.png`: 六边形寻路示意图
- `sq_out.png`: 方格寻路示意图，其中红线是离散路径，蓝线是自然路径

## 用法概览

这个库的基本调用流程很固定：

1. 用 `grid.NewLocal(nx, ny)` 创建地图。
2. 为每个分块调用 `SetGrid(i, j, new(grid.Grid))` 初始化底层存储。
3. 用 `m.Set(x, y)` 标记障碍物；未标记的位置默认可通行。
4. 创建 `hex.WorkSpace` 或 `sq.WorkSpace`，然后调用 `Reset(m)` 绑定地图。
5. 调用 `Solve(...)` 或 `SolveNatural(...)` 求路径。

需要注意：

- `grid.Local` 按 `16x16` 分块存储，所以实际地图大小是 `nx*16` x `ny*16`。
- `Solve` 返回的是离散格子路径 `[]grid.PathGrid`。
- `sq.SolveNatural` 返回的是连续路径点 `[]grid.PathPoint`，适合角色平滑移动。

## 代码定位

- 六边形 JPS: `groute/hex`
- 方格 JPS: `groute/sq`
- 地图与路径类型: `groute/grid`

如果是 LLM agent 想快速掌握项目，建议阅读顺序：

1. `README.md`
2. `demo/groute/*.go`
3. `groute/sq/solve.go` 或 `groute/hex/solve.go`
4. `groute/grid/grid16.go`

## 正六边形网格 JPS

剪枝策略参考 [Aditya Subramanian](https://adityasubramanian.weebly.com/uploads/7/0/6/3/70633237/jump_point_search_on_hexagonal_grids.pdf)

代码位置：https://github.com/LeGamerDc/pathfinding/tree/main/groute/hex

demo 位置：https://github.com/LeGamerDc/pathfinding/tree/main/demo/groute

## 正方形网格 JPS

代码位置：https://github.com/LeGamerDc/pathfinding/tree/main/groute/sq
