package main

import (
	"github.com/askeladdk/pancake"
	"github.com/askeladdk/pancake/graphics"
	gl "github.com/askeladdk/pancake/graphics/opengl"
	"github.com/askeladdk/pancake/graphics2d"
)

type TitleScreen struct {
	Background StaticImage
	Title      StaticImage
	Drawer     *graphics2d.Drawer
	Shader     *graphics.ShaderProgram
	Start      bool
}

func (s *TitleScreen) Begin() {}
func (s *TitleScreen) End()   {}

func (s *TitleScreen) Key(ev pancake.KeyEvent) error {
	s.Start = true
	return nil
}

func (s *TitleScreen) Frame(ev pancake.FrameEvent) (Screen, error) {
	if s.Start {
		return gameScreen, nil
	}
	return nil, nil
}

func (s *TitleScreen) Draw(ev pancake.DrawEvent) error {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	s.Shader.Begin()
	s.Drawer.Draw(
		s.Background,
		s.Title,
	)
	s.Shader.End()
	return nil
}
