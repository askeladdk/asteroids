package main

import (
	"fmt"

	"github.com/askeladdk/pancake/input"

	"github.com/askeladdk/pancake"
	"github.com/askeladdk/pancake/graphics"
	gl "github.com/askeladdk/pancake/graphics/opengl"
	"github.com/askeladdk/pancake/graphics2d"
	"github.com/askeladdk/pancake/text"
)

type gameOverScreen struct {
	Sim        *theSimulation
	Text       *text.Text
	Drawer     *graphics2d.Drawer
	Shader     *graphics.ShaderProgram
	Background staticImage
	Title      staticImage
	Restart    bool
}

func (s *gameOverScreen) Begin() {
	s.Restart = false
	s.Text.Clear()
	fmt.Fprintf(s.Text, "Final level: %d\nFinal score: %d\nPress Enter to restart or ESC to quit.", 1+s.Sim.Level, s.Sim.Score)
}

func (s *gameOverScreen) End() {}

func (s *gameOverScreen) Key(ev pancake.KeyEvent) error {
	switch ev.Key {
	case input.KeyEscape:
		return pancake.ErrQuit
	case input.KeyEnter:
		s.Restart = true
	}
	return nil
}

func (s *gameOverScreen) Frame(ev pancake.FrameEvent) (screen, error) {
	if s.Restart {
		s.Sim.Level = 0
		s.Sim.Score = 0
		return globalGameScreen, nil
	}

	return nil, nil
}

func (s *gameOverScreen) Draw(ev pancake.DrawEvent) error {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	s.Shader.Begin()
	s.Drawer.Draw(s.Background)
	s.Drawer.Draw(s.Title)
	s.Drawer.Draw(s.Text)
	s.Shader.End()
	return nil
}
