package main

import (
	"fmt"
	"image/color"

	"github.com/fogleman/gg"
)

func main() {
	fmt.Println("Running hex demo...")
	hexDemo()

	fmt.Println("Running square demo...")
	sqDemo()
}

// drawLine draw a line from (x1,y1) to (x2,y2) use color c.
func drawLine(ctx *gg.Context, x1, y1, x2, y2 float64, c color.Color) {
	ctx.SetColor(c)
	ctx.SetLineWidth(3)
	ctx.DrawLine(x1, y1, x2, y2)
	ctx.Stroke()
}
