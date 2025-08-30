package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/fogleman/gg"
	"github.com/legamerdc/pathfinding/groute/grid"
	"github.com/legamerdc/pathfinding/groute/hex"
)

const (
	ox    = 100
	oy    = 100
	ratio = 0.277

	nx, ny = 3, 3
)

var (
	colorG = color.RGBA{R: 144, G: 238, B: 144, A: 128}
	colorB = color.RGBA{R: 169, G: 169, B: 169, A: 255}
	colorR = color.RGBA{R: 255, G: 0, B: 0, A: 255}
)

func hexDemo() {
	bg := gg.NewContext(1000, 1000)
	ws := CreateMap()
	for i := int32(0); i < 16*nx; i++ {
		for j := int32(0); j < 16*ny; j++ {
			cx, cy := center(i, j)
			if ws.Map.Available(i, j) {
				drawHexagon(bg, ox+cx, oy+cy, 10, colorG)
			} else {
				drawHexagon(bg, ox+cx, oy+cy, 10, colorB)
			}
		}
	}
	fmt.Println("start solve")
	start := time.Now()
	var (
		path []grid.PathGrid
		ok   bool
	)
	path, ok = ws.Solve(0, 0, 16*nx-1, 16*ny-1)

	if !ok {
		fmt.Println("no path: ", time.Since(start))
	} else {
		fmt.Println("end solve: ", time.Since(start))
		for i := 1; i < len(path); i++ {
			x1, y1 := center(path[i-1].X, path[i-1].Y)
			x2, y2 := center(path[i].X, path[i].Y)
			drawLine(bg, ox+x1, oy+y1, ox+x2, oy+y2, colorR)
		}
	}

	//drawHexagon(bg, 200, 200, 10, color.RGBA{R: 144, G: 238, B: 144, A: 128})
	_ = bg.SavePNG("out.png")
}

func CreateMap() *hex.WorkSpace {
	m := grid.NewLocal(nx, ny)
	for i := 0; i < nx; i++ {
		for j := 0; j < ny; j++ {
			m.SetGrid(int32(i), int32(j), new(grid.Grid))
		}
	}
	ws := hex.NewWorkSpace(1200)
	for i := int32(0); i < 16*nx; i++ {
		for j := int32(0); j < 16*ny; j++ {
			if (i == 0 && j == 0) || (i == 16*nx-1 && j == 16*ny-1) {
				continue
			}
			if rand.Float32() < ratio {
				m.Set(i, j)
			}
		}
	}
	ws.Reset(m)
	return ws
}

func center(x, y int32) (cx, cy float64) {
	cy = float64(y) * 15
	cx = math.Sqrt(3) * 5 * float64(2*x+(y&1))
	return
}

// drawHexagon draw a hexagon at pos (x, y) with size, fill it with color and draw line use color.Black
func drawHexagon(ctx *gg.Context, x, y, size float64, fill color.Color) {
	// Number of sides in a hexagon
	const sides = 6

	// Begin a new path
	ctx.ClearPath()

	// Draw the hexagon with flat top and bottom sides
	for i := 0; i < sides; i++ {
		// Calculate angle (starting from pi/6 to get flat sides at top/bottom)
		angle := math.Pi/6 + 2.0*math.Pi*float64(i)/float64(sides)
		vx := x + size*math.Cos(angle)
		vy := y + size*math.Sin(angle)

		if i == 0 {
			ctx.MoveTo(vx, vy)
		} else {
			ctx.LineTo(vx, vy)
		}
	}

	// Close the path
	ctx.ClosePath()

	// Fill with the provided color
	ctx.SetColor(fill)
	ctx.Fill()

	// Outline with black
	ctx.SetColor(color.Black)
	ctx.Stroke()
}
