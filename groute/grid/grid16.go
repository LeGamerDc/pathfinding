package grid

const g16 = 16

type (
	Grid struct {
		Bits [g16]uint16
	}

	Local struct {
		Grids  [][]*Grid
		Nx, Ny int32
	}

	PathGrid struct {
		X, Y int32
	}
)

// NewLocal allocates a map composed of nx by ny 16x16 grid blocks.
func NewLocal(nx, ny int32) *Local {
	grids := make([][]*Grid, nx)
	for i := range grids {
		grids[i] = make([]*Grid, ny)
	}
	return &Local{
		Grids: grids,
		Nx:    nx,
		Ny:    ny,
	}
}

// SetGrid assigns the storage block at block coordinate (nx, ny).
func (w *Local) SetGrid(nx, ny int32, g *Grid) {
	w.Grids[nx][ny] = g
}

// GetGrid returns the storage block at block coordinate (nx, ny).
func (w *Local) GetGrid(nx, ny int32) *Grid {
	return w.Grids[nx][ny]
}

// Available reports whether map cell (x, y) is inside bounds and not blocked.
func (w *Local) Available(x, y int32) bool {
	//  x < 0 || x >= w.Nx*g16 || y < 0 || y >= w.Ny*g16
	if uint32(x) >= uint32(w.Nx*g16) || uint32(y) >= uint32(w.Ny*g16) {
		return false
	}
	var (
		nx, ny = x / g16, y / g16
		ix, iy = x % g16, y % g16
	)
	g := w.Grids[nx][ny]
	return g.Bits[iy]&(1<<ix) == 0
}

// Set marks map cell (x, y) as blocked.
func (w *Local) Set(x, y int32) {
	if x < 0 || x >= w.Nx*g16 || y < 0 || y >= w.Ny*g16 {
		return
	}
	var (
		nx, ny = x / g16, y / g16
		ix, iy = x % g16, y % g16
	)
	g := w.Grids[nx][ny]
	g.Bits[iy] |= 1 << ix
}
