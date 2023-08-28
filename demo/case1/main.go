package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	_ "image/png"
	"log"
	"pathfinding/demo"
	"pathfinding/grid"
)

func main() {
	world := newWorld()
	world.Init()
	ebiten.SetWindowSize(demo.ScreenWidth, demo.ScreenHeight)
	ebiten.SetWindowTitle("agent demo")
	//world.ws.Fill(0, 0)
	//world.ws.Dump()
	if err := ebiten.RunGame(world); err != nil {
		log.Fatal(err)
	}
}

func newWorld() *demo.World {
	g := make([][]grid.Grid, 3)
	for i := 0; i < 3; i++ {
		g[i] = make([]grid.Grid, 3)
	}
	t := &grid.Terrain{
		Grids: g,
		Nx:    3,
		Ny:    3,
	}
	//for i := 0; i < 48; i++ {
	//	for j := 0; j < 48; j++ {
	//		t.Set(int32(i), int32(j))
	//	}
	//}
	t.Set(1, 3)
	t.Set(2, 0)
	t.Set(2, 1)
	t.Set(2, 3)
	t.Set(3, 3)
	t.Set(5, 1)
	w := &demo.World{
		Map: t,
	}
	test3(w)
	return w
}

func test1(w *demo.World) {
	w.SetAgent(demo.Vec2{X: 6.9, Y: 3.8}, demo.Vec2{X: 1.2, Y: 0.4}, 0.95)
	w.SetAgent(demo.Vec2{X: 0.5, Y: 0.5}, demo.Vec2{X: 5.5, Y: 0.5}, 0.8)
}

func test2(w *demo.World) {
	w.SetAgent(demo.Vec2{X: 8.5, Y: 8.4}, demo.Vec2{X: 31.5, Y: 8.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 32.5, Y: 8.6}, demo.Vec2{X: 9.5, Y: 8.5}, 1.2)
}

func test3(w *demo.World) {
	w.SetAgent(demo.Vec2{X: 25.1, Y: 20}, demo.Vec2{X: 25, Y: 28}, 1.2)
	w.SetAgent(demo.Vec2{X: 24.9, Y: 30}, demo.Vec2{X: 25, Y: 22}, 1.2)
	w.SetAgent(demo.Vec2{X: 20, Y: 24.9}, demo.Vec2{X: 28, Y: 25}, 1.2)
	w.SetAgent(demo.Vec2{X: 30, Y: 25.1}, demo.Vec2{X: 22, Y: 25}, 1.2)
}

func test4(w *demo.World) {
	w.SetAgent(demo.Vec2{X: 20.5, Y: 20.5}, demo.Vec2{X: 20.5, Y: 20.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 21.5, Y: 20.5}, demo.Vec2{X: 21.5, Y: 20.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 22.5, Y: 21.5}, demo.Vec2{X: 22.5, Y: 21.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 23.5, Y: 21.5}, demo.Vec2{X: 23.5, Y: 21.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 24.5, Y: 22.5}, demo.Vec2{X: 24.5, Y: 22.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 24.5, Y: 25.5}, demo.Vec2{X: 24.5, Y: 25.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 23.5, Y: 26.5}, demo.Vec2{X: 23.5, Y: 26.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 22.5, Y: 26.5}, demo.Vec2{X: 22.5, Y: 26.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 21.5, Y: 27.5}, demo.Vec2{X: 21.5, Y: 27.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 20.5, Y: 27.5}, demo.Vec2{X: 20.5, Y: 27.5}, 1.2)

	w.SetAgent(demo.Vec2{X: 25.5, Y: 22.5}, demo.Vec2{X: 24.5, Y: 23.5}, 1.2)
	w.SetAgent(demo.Vec2{X: 25.5, Y: 25.5}, demo.Vec2{X: 24.5, Y: 24.5}, 1.2)

	w.SetAgent(demo.Vec2{X: 11.7, Y: 23.5}, demo.Vec2{X: 31.1, Y: 23.5}, 1.7)
}
