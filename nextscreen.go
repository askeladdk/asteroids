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

type nextScreen struct {
	Sim        *theSimulation
	Text       *text.Text
	Background staticImage
	Title      staticImage
	Drawer     *graphics2d.Drawer
	Shader     *graphics.ShaderProgram
	Start      bool
}

func (s *nextScreen) Begin() {
	s.Start = false
	s.Sim.Level++

	s.Text.Clear()
	fmt.Fprintf(s.Text, "Level: %d\nScore: %d\nPress Enter to continue.", 1+s.Sim.Level, s.Sim.Score)
}

func (s *nextScreen) End() {}

func (s *nextScreen) Key(ev pancake.KeyEvent) error {
	if ev.Key == input.KeyEnter {
		s.Start = true
	}
	return nil
}

func (s *nextScreen) Frame(ev pancake.FrameEvent) (screen, error) {
	if s.Start {
		return globalGameScreen, nil
	}
	return nil, nil
}

func (s *nextScreen) Draw(ev pancake.DrawEvent) error {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	s.Shader.Begin()
	s.Drawer.Draw(s.Background)
	s.Drawer.Draw(s.Title)
	s.Drawer.Draw(s.Text)
	s.Shader.End()
	return nil
}
