package main

import "github.com/askeladdk/pancake"

type Screen interface {
	Begin()
	End()
	Key(pancake.KeyEvent) error
	Draw(pancake.DrawEvent) error
	Frame(pancake.FrameEvent) (Screen, error)
}

type TransitionScreen struct {
	To Screen
}

func (s TransitionScreen) Begin()                                      {}
func (s TransitionScreen) End()                                        {}
func (s TransitionScreen) Key(ev pancake.KeyEvent) error               { return nil }
func (s TransitionScreen) Draw(ev pancake.DrawEvent) error             { return nil }
func (s TransitionScreen) Frame(ev pancake.FrameEvent) (Screen, error) { return s.To, nil }

type ScreenState struct {
	Screen Screen
}

func (gs *ScreenState) Do(event interface{}) error {
	switch ev := event.(type) {
	case pancake.KeyEvent:
		return gs.Screen.Key(ev)
	case pancake.FrameEvent:
		next, err := gs.Screen.Frame(ev)
		if next != nil {
			gs.Screen.End()
			next.Begin()
			gs.Screen = next
		}
		return err
	case pancake.DrawEvent:
		return gs.Screen.Draw(ev)
	default:
		return nil
	}
}
