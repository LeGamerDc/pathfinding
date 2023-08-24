package demo

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"log"
	"pathfinding/demo/res"
	"pathfinding/grid"
	"pathfinding/rvo2"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	blockSize        = 22
	offsetX, offsetY = 100, 100
	timePerTick      = 0.25
)

var (
	agentImg   *ebiten.Image
	blockImg   *ebiten.Image
	background *ebiten.Image

	ScreenWidth           int
	ScreenHeight          int
	gridWidth, gridHeight int

	defaultDir = Vec2{X: 0, Y: -1}
)

func Size() (x, y int) {
	nx := 3*16*blockSize + offsetX*2
	ny := 3*16*blockSize + offsetY*2
	return nx, ny
}

func init() {
	// init agentImg
	img, _, err := image.Decode(bytes.NewReader(res.AgentPng))
	if err != nil {
		log.Fatal(err)
	}
	tmp := ebiten.NewImageFromImage(img)
	bound := tmp.Bounds()
	agentImg = ebiten.NewImage(blockSize, blockSize)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-float64(bound.Min.X), -float64(bound.Min.Y))
	op.GeoM.Scale(blockSize/float64(bound.Dx()), blockSize/float64(bound.Dy()))
	agentImg.DrawImage(tmp, op)

	// init blockImg
	img, _, err = image.Decode(bytes.NewReader(res.BlockPng))
	if err != nil {
		log.Fatal(err)
	}
	tmp = ebiten.NewImageFromImage(img)
	bound = tmp.Bounds()
	blockImg = ebiten.NewImage(blockSize, blockSize)
	op.GeoM.Reset()
	op.GeoM.Scale(blockSize/float64(bound.Dy()), blockSize/float64(bound.Dy()))
	blockImg.DrawImage(tmp.SubImage(image.Rect(0, 0, 10, 10)).(*ebiten.Image), op)
}

type World struct {
	Map    *grid.Terrain
	Agents []*Agent

	ws        *grid.WorkSpace
	lastFrame float64
	lastTick  float64
	start     bool
	t         int
	// rvo related
	rvo *rvo2.RvoConfig
}

func (w *World) Init() {
	ScreenWidth, ScreenHeight = Size()
	gridWidth, gridHeight = 48, 48
	background = ebiten.NewImage(ScreenWidth, ScreenHeight)
	background.Fill(color.Black)
	// draw board
	tmp := ebiten.NewImage(48*blockSize, 48*blockSize)
	tmp.Fill(color.Gray16{Y: 0xfff})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(offsetX, offsetY)
	background.DrawImage(tmp, op)
	// draw vertical lines
	for i := 1; i < gridWidth; i++ {
		x := offsetX + blockSize*i
		for j := offsetY; j <= offsetY+blockSize*gridHeight; j++ {
			background.Set(x, j, color.Gray16{Y: 0x1fff})
		}
	}
	// draw horizon lines
	for j := 1; j < gridHeight; j++ {
		y := offsetY + blockSize*j
		for i := offsetX; i <= offsetX+blockSize*gridWidth; i++ {
			background.Set(i, y, color.Gray16{Y: 0x1fff})
		}
	}
	// draw blocks
	for i := 0; i < gridWidth; i++ {
		for j := 0; j < gridHeight; j++ {
			if w.Map.IsSet(int32(i), int32(j)) {
				op.GeoM.Reset()
				op.GeoM.Translate(float64(offsetX+i*blockSize), float64(offsetY+j*blockSize))
				background.DrawImage(blockImg, op)
			}
		}
	}
	w.ws = grid.NewWorkSpace(w.Map)
	w.rvo = &rvo2.RvoConfig{
		TimeStep:    timePerTick,
		TimeHorizon: timePerTick * 4,
		MaxNeighbor: 8,
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(background, op)
	for _, a := range w.Agents {
		op.GeoM.Reset()
		op.GeoM.Translate(-0.5*blockSize, -0.5*blockSize)
		op.GeoM.Rotate(angle(defaultDir, a.dir))
		op.GeoM.Translate(offsetX+a.Pos.X*blockSize, offsetY+a.Pos.Y*blockSize)
		screen.DrawImage(agentImg, op)
	}
}

func (w *World) Layout(_, _ int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (w *World) Update() error {
	now := float64(time.Now().UnixNano()) / 1e9
	if !w.start {
		w.start = true
		w.lastFrame = now
		w.lastTick = now
		return nil
	}
	if gap := now - w.lastTick; gap > timePerTick {
		w.tick(gap, w.t)
		w.t++
		w.lastTick = now
	} else {
		gap = now - w.lastFrame
		for _, a := range w.Agents {
			a.frame(gap)
		}
		w.lastFrame = now
	}
	return nil
}

func (w *World) tick(ti float64, t int) {
	// prepare nextPos and velocity/prefVelocity
	for _, agent := range w.Agents {
		switch agent.st {
		case StatusTerminate:
			continue
		case StatusStop:
			if agent.wait > 0 {
				agent.wait--
				continue
			}
			var ok bool
			agent.path, ok = agent.route(w)
			fmt.Printf("route ack %d %v [%v] %v\n", t, agent.Pos, agent.path, ok)
			if !ok {
				agent.wait = nextWait()
				continue
			}
			agent.st = StatusMoving
		case StatusMoving:
			agent.Pos = agent.nextPos
			to, ok := agent.getPathNextPoint()
			if !ok {
				agent.path = nil
				agent.st = StatusTerminate
			} else {
				//w.setObstacle(agent)
				ax, ay := pos2grid(agent.Pos)
				bx, by := pos2grid(to)
				if w.Map.Los(ax, ay, bx, by) {
					agent.setLocalTarget(to, ti)
				} else { // must re-route now !!!
					agent.Stop()
				}
			}
		}
	}
	// avoid collide
	for idx, agent := range w.Agents {
		if agent.st == StatusMoving {
			jdx := -1
			v, u, ok := w.rvo.Solve(rvoAgent(agent), func() (*rvo2.Agent, bool) {
				for jdx+1 < len(w.Agents) {
					jdx++
					if jdx == idx || agent.st != StatusMoving {
						continue
					}
					return rvoAgent(w.Agents[jdx]), true
				}
				return nil, false
			})
			if !ok {
				fmt.Println("stop solve")
				agent.Stop()
			} else {
				agent.velocity = Vec2{X: v.X, Y: v.Y}
				fmt.Printf("%v %v -> %v\n", u, agent.prefVelocity, agent.velocity)
				agent.nextPos = vAdd(agent.Pos, agent.velocity.mul(ti))
				//if agent.velocity.absSq() < Epsilon {
				//	fmt.Println("stop no speed")
				//	agent.Stop()
				//} else {
				//	agent.nextPos = vAdd(agent.Pos, agent.velocity.mul(ti))
				//}
			}
		}
	}
	// check collide
	for _, agent := range w.Agents {
		if agent.st != StatusMoving {
			continue
		}
		x, y := pos2grid(agent.nextPos)
		if w.ws.IsSet(x, y) {
			fmt.Printf("stop wall %v -> ", agent.Pos)
			agent.Stop()
			fmt.Printf("%v \n", agent.Pos)
		} else {
			for _, b := range w.Agents {
				if b.st != StatusMoving && sameGrid(agent.nextPos, b.Pos) {
					fmt.Println("stop collide")
					agent.Stop()
				}
			}
		}
	}
}

func (w *World) SetAgent(pos, target Vec2, speed float64) {
	w.Agents = append(w.Agents, &Agent{
		Pos:     pos,
		Target:  target,
		Speed:   speed,
		st:      StatusStop,
		lastPos: pos,
		nextPos: pos,
		dir:     defaultDir,
	})
}

func rvoAgent(a *Agent) *rvo2.Agent {
	r := &rvo2.Agent{
		Position:     rvo2.Vec2{X: a.Pos.X, Y: a.Pos.Y},
		PrefVelocity: rvo2.Vec2{X: a.prefVelocity.X, Y: a.prefVelocity.Y},
		Velocity:     rvo2.Vec2{X: a.velocity.X, Y: a.velocity.Y},

		Speed:  a.Speed,
		Radius: 0.5,
	}
	return r
}
