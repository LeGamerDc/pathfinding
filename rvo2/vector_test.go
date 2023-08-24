package rvo2

import (
	"math/rand"
	"testing"
)

func randVec2() Vec2 {
	return Vec2{
		X: (rand.Float64() - 0.5) * 10,
		Y: (rand.Float64() - 0.5) * 10,
	}
}

func TestLeft(t *testing.T) {
	for i := 0; i < 10; i++ {
		a, b, c := randVec2(), randVec2(), randVec2()
		c1 := vCross(vSub(b, a), vSub(c, a))
		c2 := vCross(vSub(a, c), vSub(b, a))
		if !equal(c1, c2) {
			t.Errorf("cross not equal")
		}
	}
}
