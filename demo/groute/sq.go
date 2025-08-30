package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/fogleman/gg"
	"github.com/legamerdc/pathfinding/groute/grid"
	"github.com/legamerdc/pathfinding/groute/sq"
)

const (
	sqOx    = 100
	sqOy    = 100
	sqRatio = 0.277
	sqSize  = 15 // 正方形的边长

	sqNx, sqNy = 3, 3
)

var (
	sqColorG = color.RGBA{R: 144, G: 238, B: 144, A: 128}
	sqColorB = color.RGBA{R: 169, G: 169, B: 169, A: 255}
	sqColorR = color.RGBA{R: 255, G: 0, B: 0, A: 255}
)

func sqDemo() {
	bg := gg.NewContext(1000, 1000)
	ws := createSqMap()
	for i := int32(0); i < 16*sqNx; i++ {
		for j := int32(0); j < 16*sqNy; j++ {
			cx, cy := sqCenter(i, j)
			if ws.Map.Available(i, j) {
				drawSquare(bg, sqOx+cx, sqOy+cy, sqSize, sqColorG)
			} else {
				drawSquare(bg, sqOx+cx, sqOy+cy, sqSize, sqColorB)
			}
		}
	}
	fmt.Println("start solve")
	start := time.Now()
	var (
		path []grid.PathGrid
		ok   bool
	)
	path, ok = ws.Solve(0, 0, 16*sqNx-1, 16*sqNy-1)

	if !ok {
		fmt.Println("no path: ", time.Since(start))
	} else {
		fmt.Println("end solve: ", time.Since(start))
		for i := 1; i < len(path); i++ {
			x1, y1 := sqCenter(path[i-1].X, path[i-1].Y)
			x2, y2 := sqCenter(path[i].X, path[i].Y)
			drawLine(bg, sqOx+x1, sqOy+y1, sqOx+x2, sqOy+y2, sqColorR)
		}
	}

	_ = bg.SavePNG("sq_out.png")
}

func createSqMap() *sq.WorkSpace {
	m := grid.NewLocal(sqNx, sqNy)
	for i := 0; i < sqNx; i++ {
		for j := 0; j < sqNy; j++ {
			m.SetGrid(int32(i), int32(j), new(grid.Grid))
		}
	}
	ws := sq.NewWorkSpace(1200)
	for i := int32(0); i < 16*sqNx; i++ {
		for j := int32(0); j < 16*sqNy; j++ {
			if (i == 0 && j == 0) || (i == 16*sqNx-1 && j == 16*sqNy-1) {
				continue
			}
			if rand.Float32() < sqRatio {
				m.Set(i, j)
			}
		}
	}
	ws.Reset(m)
	return ws
}

func sqCenter(x, y int32) (cx, cy float64) {
	cx = float64(x) * float64(sqSize)
	cy = float64(y) * float64(sqSize)
	return
}

// drawSquare 绘制一个正方形
func drawSquare(ctx *gg.Context, x, y, size float64, fill color.Color) {
	ctx.SetColor(fill)
	ctx.DrawRectangle(x-size/2, y-size/2, size, size)
	ctx.Fill()

	// 绘制边框
	ctx.SetColor(color.Black)
	ctx.SetLineWidth(1)
	ctx.DrawRectangle(x-size/2, y-size/2, size, size)
	ctx.Stroke()
}
