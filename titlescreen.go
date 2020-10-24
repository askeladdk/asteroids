package main

import (
	"fmt"

	"github.com/askeladdk/pancake/input"
	"github.com/askeladdk/pancake/mathx"

	"github.com/askeladdk/pancake"
	"github.com/askeladdk/pancake/graphics"
	gl "github.com/askeladdk/pancake/graphics/opengl"
	"github.com/askeladdk/pancake/graphics2d"
	"github.com/askeladdk/pancake/text"
)

type TitleScreen struct {
	Background StaticImage
	Title      StaticImage
	Drawer     *graphics2d.Drawer
	Shader     *graphics.ShaderProgram
	Text       *text.Text
	Start      bool
}

func (s *TitleScreen) Begin() {
	s.Text.Clear()
	s.Text.Pos = mathx.Vec2{4, 360 - 5*s.Text.LineHeight - 4}
	fmt.Fprintf(s.Text, "Music by Eric Matyas (www.soundimage.org)\n")
	fmt.Fprintf(s.Text, "Sprites by CDmir (www.opengameart.org)\n")
	fmt.Fprintf(s.Text, "Background by OdinTdh (www.opengameart.org)\n")
	fmt.Fprintf(s.Text, "Sound effects made with JFXR (jfxr.frozenfractal.com)\n")
	fmt.Fprintf(s.Text, "Programmed by Askeladd (github.com/askeladdk/asteroids)")
}

func (s *TitleScreen) End() {
	s.Text.Pos[1] = 0
}

func (s *TitleScreen) Key(ev pancake.KeyEvent) error {
	switch ev.Key {
	case input.KeyEscape:
		return pancake.Quit
	default:
		s.Start = true
		return nil
	}
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
	s.Drawer.Draw(s.Background)
	s.Drawer.Draw(s.Title)
	s.Drawer.Draw(s.Text)
	s.Shader.End()
	return nil
}
