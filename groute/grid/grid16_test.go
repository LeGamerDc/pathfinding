package grid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocal(t *testing.T) {
	zero := Grid{}
	w := NewLocal(3, 3)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			w.SetGrid(int32(i), int32(j), &zero)
		}
	}
	for x := int32(0); x < 48; x++ {
		for y := int32(0); y < 48; y++ {
			assert.True(t, w.Available(x, y))
		}
	}
	assert.False(t, w.Available(-1, 0))
	assert.False(t, w.Available(0, -1))
	assert.False(t, w.Available(48, 0))
	assert.False(t, w.Available(0, 48))
	assert.False(t, w.Available(48, 48))
}

func TestHash(t *testing.T) {
	var a = Gpos{X: 1, Y: 2}
	var b = Gpos{X: 1, Y: 2}
	assert.Equal(t, a.Hash(), b.Hash())
}
