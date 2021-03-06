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

type gameScreen struct {
	Sim        *theSimulation
	Text       *text.Text
	Drawer     *graphics2d.Drawer
	Shader     *graphics.ShaderProgram
	Background staticImage
	Keys       uint32
}

func (g *gameScreen) Begin() {
	g.Keys = 0
	g.Sim.Reset()
}

func (g *gameScreen) End() {}

func (g *gameScreen) Key(ev pancake.KeyEvent) error {
	switch ev.Key {
	case input.KeyEscape:
		return pancake.ErrQuit
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
			g.Sim.Action(shipID, actionFire, 0)
		}
	}
	return nil
}

func (g *gameScreen) Frame(ev pancake.FrameEvent) (screen, error) {
	switch g.Sim.State {
	case stateGAMEOVER:
		return globalGameOverScreen, nil
	case stateNEXTLEVEL:
		return globalNextScreen, nil
	}

	if g.Keys&3 == 1 {
		g.Sim.Action(shipID, actionTurn, -1)
	} else if g.Keys&3 == 2 {
		g.Sim.Action(shipID, actionTurn, +1)
	}

	if g.Keys&4 != 0 {
		g.Sim.Action(shipID, actionForward, 1)
	}

	g.Sim.Frame(ev.DeltaTime)

	g.Text.Clear()
	fmt.Fprintf(g.Text, "Level: %d\nScore: %d", 1+g.Sim.Level, g.Sim.Score)

	return nil, nil
}

func (g *gameScreen) Draw(ev pancake.DrawEvent) error {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	g.Shader.Begin()
	g.Sim.Alpha = ev.Alpha
	g.Drawer.Draw(g.Background)
	g.Drawer.Draw(g.Sim)
	g.Drawer.Draw(g.Text)
	g.Shader.End()
	return nil
}

func toggleFlag(flags uint32, flag uint32, state bool) uint32 {
	if state {
		return flags | flag
	}
	return flags &^ flag
}
