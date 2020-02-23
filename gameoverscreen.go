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

type GameOverScreen struct {
	Sim        *Simulation
	Text       *text.Text
	Drawer     *graphics2d.Drawer
	Shader     *graphics.ShaderProgram
	Background StaticImage
	Title      StaticImage
	Restart    bool
}

func (s *GameOverScreen) Begin() {
	s.Restart = false
	s.Text.Clear()
	fmt.Fprintf(s.Text, "Final level: %d\nFinal score: %d\nPress Enter to restart or ESC to quit.", 1+s.Sim.Level, s.Sim.Score)
}

func (s *GameOverScreen) End() {}

func (s *GameOverScreen) Key(ev pancake.KeyEvent) error {
	switch ev.Key {
	case input.KeyEscape:
		return pancake.Quit
	case input.KeyEnter:
		s.Restart = true
	}
	return nil
}

func (s *GameOverScreen) Frame(ev pancake.FrameEvent) (Screen, error) {
	if s.Restart {
		s.Sim.Level = 0
		s.Sim.Score = 0
		return gameScreen, nil
	}

	return nil, nil
}

func (s *GameOverScreen) Draw(ev pancake.DrawEvent) error {
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
