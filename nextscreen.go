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

type NextScreen struct {
	Sim        *Simulation
	Text       *text.Text
	Background StaticImage
	Title      StaticImage
	Drawer     *graphics2d.Drawer
	Shader     *graphics.ShaderProgram
	Start      bool
}

func (s *NextScreen) Begin() {
	s.Start = false
	s.Sim.Level++

	s.Text.Clear()
	fmt.Fprintf(s.Text, "Level: %d\nScore: %d\nPress Enter to start.", 1+s.Sim.Level, s.Sim.Score)
}

func (s *NextScreen) End() {}

func (s *NextScreen) Key(ev pancake.KeyEvent) error {
	if ev.Key == input.KeyEnter {
		s.Start = true
	}
	return nil
}

func (s *NextScreen) Frame(ev pancake.FrameEvent) (Screen, error) {
	if s.Start {
		return gameScreen, nil
	}
	return nil, nil
}

func (s *NextScreen) Draw(ev pancake.DrawEvent) error {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	s.Shader.Begin()
	s.Drawer.Draw(
		s.Background,
		s.Title,
		s.Text,
	)
	s.Shader.End()
	return nil
}
