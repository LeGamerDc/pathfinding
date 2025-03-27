package hex

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestTransform(t *testing.T) {
	for i := 0; i < 100; i++ {
		x := rand.Int31n(200) - 100
		y := rand.Int31n(200) - 100
		q, r := xy2qr(x, y)
		x1, y1 := qr2xy(q, r)
		assert.Equal(t, x1, x, "x")
		assert.Equal(t, y1, y, "y")
	}
}
