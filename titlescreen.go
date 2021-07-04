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

type titleScreen struct {
	Background staticImage
	Title      staticImage
	Drawer     *graphics2d.Drawer
	Shader     *graphics.ShaderProgram
	Text       *text.Text
	Start      bool
}

func (s *titleScreen) Begin() {
	s.Text.Clear()
	s.Text.Pos = mathx.Vec2{4, 360 - 5*s.Text.LineHeight - 4}
	fmt.Fprintf(s.Text, "Music by Eric Matyas (www.soundimage.org)\n")
	fmt.Fprintf(s.Text, "Sprites by CDmir (www.opengameart.org)\n")
	fmt.Fprintf(s.Text, "Background by OdinTdh (www.opengameart.org)\n")
	fmt.Fprintf(s.Text, "Sound effects made with JFXR (jfxr.frozenfractal.com)\n")
	fmt.Fprintf(s.Text, "Programmed by Askeladd (github.com/askeladdk/asteroids)")
}

func (s *titleScreen) End() {
	s.Text.Pos[1] = 0
}

func (s *titleScreen) Key(ev pancake.KeyEvent) error {
	switch ev.Key {
	case input.KeyEscape:
		return pancake.ErrQuit
	default:
		s.Start = true
		return nil
	}
}

func (s *titleScreen) Frame(ev pancake.FrameEvent) (screen, error) {
	if s.Start {
		return globalGameScreen, nil
	}
	return nil, nil
}

func (s *titleScreen) Draw(ev pancake.DrawEvent) error {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	s.Shader.Begin()
	s.Drawer.Draw(s.Background)
	s.Drawer.Draw(s.Title)
	s.Drawer.Draw(s.Text)
	s.Shader.End()
	return nil
}
