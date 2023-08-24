package main

import (
	"fmt"
	"pathfinding/grid"
	"time"
)

func main() {
	g := make([][]grid.Grid, 3)
	for i := 0; i < 3; i++ {
		g[i] = make([]grid.Grid, 3)
	}
	t := &grid.Terrain{
		Grids: g,
		Nx:    3,
		Ny:    3,
	}
	ws := grid.NewWorkSpace(t)
	ws.Fill(0, 0)
	ws.Set(1, 3)
	ws.Set(2, 0)
	ws.Set(2, 1)
	ws.Set(2, 3)
	ws.Set(3, 3)
	ws.Set(5, 1)
	ws.Dump()
	fmt.Println(ws.Solve(0, 0, 7, 0))
	ts(func() {
		ws.Fill(0, 0)
		ws.Set(1, 3)
		ws.Set(2, 0)
		ws.Set(2, 1)
		ws.Set(2, 3)
		ws.Set(3, 3)
		ws.Set(5, 1)
		ws.Solve(0, 0, 7, 0)
	}, 1000)

}

func ts(f func(), n int) {
	s := time.Now()
	for i := 0; i < n; i++ {
		f()
	}
	fmt.Println(time.Since(s))
}
