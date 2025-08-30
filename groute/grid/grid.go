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

func (w *Local) SetGrid(nx, ny int32, g *Grid) {
	w.Grids[nx][ny] = g
}

func (w *Local) GetGrid(nx, ny int32) *Grid {
	return w.Grids[nx][ny]
}

func (w *Local) Available(x, y int32) bool {
	if x < 0 || x >= w.Nx*g16 || y < 0 || y >= w.Ny*g16 {
		return false
	}
	var (
		nx, ny = x / g16, y / g16
		ix, iy = x % g16, y % g16
	)
	g := w.Grids[nx][ny]
	return g.Bits[iy]>>ix&1 == 0
}

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
