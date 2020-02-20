package main

import (
	"fmt"
	"image/color"

	"github.com/askeladdk/pancake"
	"github.com/askeladdk/pancake/graphics"
	gl "github.com/askeladdk/pancake/graphics/opengl"
	"github.com/askeladdk/pancake/graphics2d"
	"github.com/askeladdk/pancake/input"
	"github.com/askeladdk/pancake/text"
)

type GameScreen struct {
	app        pancake.App
	sim        *Simulation
	fpstext    *text.Text
	drawer     *graphics2d.Drawer
	shader     *graphics.ShaderProgram
	background StaticImage
	keys       uint32
}

func (g *GameScreen) Begin() {
	g.sim.Reset()
	g.sim.SpawnSpaceship()
}

func (g *GameScreen) End() {}

func (g *GameScreen) Key(ev pancake.KeyEvent) error {
	switch ev.Key {
	case input.KeyEscape:
		return pancake.Quit
	case input.KeyA:
		fallthrough
	case input.KeyLeft:
		g.keys = toggleFlag(g.keys, 1, ev.Flags.Down())
	case input.KeyD:
		fallthrough
	case input.KeyRight:
		g.keys = toggleFlag(g.keys, 2, ev.Flags.Down())
	case input.KeyW:
		fallthrough
	case input.KeyUp:
		g.keys = toggleFlag(g.keys, 4, ev.Flags.Down())
	case input.KeyP:
		if ev.Flags.Pressed() {
			g.sim.SpawnAsteroid()
		}
	case input.KeySpace:
		if ev.Flags.Pressed() {
			e := g.sim.At(SHIPID)
			g.sim.SpawnBullet(e.Pos, e.Rot)
		}
	}
	return nil
}

func (g *GameScreen) Frame(ev pancake.FrameEvent) (Screen, error) {
	if g.keys&3 == 1 {
		g.sim.Action(SHIPID, TURN, -1)
	} else if g.keys&3 == 2 {
		g.sim.Action(SHIPID, TURN, +1)
	}

	if g.keys&4 != 0 {
		g.sim.Action(SHIPID, FORWARD, 1)
	}

	g.sim.Frame(float32(ev.DeltaTime))

	g.fpstext.Clear()
	g.fpstext.Color = color.NRGBA{255, 255, 255, 255}
	fmt.Fprintf(g.fpstext, "FPS: ")
	g.fpstext.Color = color.NRGBA{255, 0, 0, 255}
	fmt.Fprintf(g.fpstext, "%d", g.app.FrameRate())

	return nil, nil
}

func (g *GameScreen) Draw(ev pancake.DrawEvent) error {
	gl.ClearColor(0, 0, 1, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	g.shader.Begin()
	g.drawer.Draw(g.background)
	g.sim.Alpha = float32(ev.Alpha)
	g.drawer.Draw(g.sim)
	g.drawer.Draw(g.fpstext)
	g.shader.End()
	return nil
}

func toggleFlag(flags uint32, flag uint32, state bool) uint32 {
	if state {
		return flags | flag
	} else {
		return flags &^ flag
	}
}
