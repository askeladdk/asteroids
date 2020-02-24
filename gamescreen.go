package main

import (
	"fmt"

	"github.com/askeladdk/pancake"
	"github.com/askeladdk/pancake/graphics"
	gl "github.com/askeladdk/pancake/graphics/opengl"
	"github.com/askeladdk/pancake/graphics2d"
	"github.com/askeladdk/pancake/input"
	"github.com/askeladdk/pancake/text"
)

type GameScreen struct {
	Sim        *Simulation
	Text       *text.Text
	Drawer     *graphics2d.Drawer
	Shader     *graphics.ShaderProgram
	Background StaticImage
	Keys       uint32
}

func (g *GameScreen) Begin() {
	g.Keys = 0
	g.Sim.Reset()
}

func (g *GameScreen) End() {}

func (g *GameScreen) Key(ev pancake.KeyEvent) error {
	switch ev.Key {
	case input.KeyEscape:
		return pancake.Quit
	case input.KeyA:
		fallthrough
	case input.KeyLeft:
		g.Keys = toggleFlag(g.Keys, 1, ev.Flags.Down())
	case input.KeyD:
		fallthrough
	case input.KeyRight:
		g.Keys = toggleFlag(g.Keys, 2, ev.Flags.Down())
	case input.KeyW:
		fallthrough
	case input.KeyUp:
		g.Keys = toggleFlag(g.Keys, 4, ev.Flags.Down())
	case input.KeyP:
		if ev.Flags.Pressed() {
			g.Sim.SpawnAsteroid()
		}
	case input.KeySpace:
		if ev.Flags.Pressed() {
			g.Sim.Action(SHIPID, FIRE, 0)
		}
	}
	return nil
}

func (g *GameScreen) Frame(ev pancake.FrameEvent) (Screen, error) {
	switch g.Sim.State {
	case GAMEOVER:
		return gameOverScreen, nil
	case NEXTLEVEL:
		return nextScreen, nil
	}

	if g.Keys&3 == 1 {
		g.Sim.Action(SHIPID, TURN, -1)
	} else if g.Keys&3 == 2 {
		g.Sim.Action(SHIPID, TURN, +1)
	}

	if g.Keys&4 != 0 {
		g.Sim.Action(SHIPID, FORWARD, 1)
	}

	g.Sim.Frame(float32(ev.DeltaTime))

	g.Text.Clear()
	fmt.Fprintf(g.Text, "Level: %d\nScore: %d", 1+g.Sim.Level, g.Sim.Score)

	return nil, nil
}

func (g *GameScreen) Draw(ev pancake.DrawEvent) error {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	g.Shader.Begin()
	g.Sim.Alpha = float32(ev.Alpha)
	g.Drawer.Draw(g.Background,
		g.Sim,
		g.Text,
	)
	g.Shader.End()
	return nil
}

func toggleFlag(flags uint32, flag uint32, state bool) uint32 {
	if state {
		return flags | flag
	} else {
		return flags &^ flag
	}
}
