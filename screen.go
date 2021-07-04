package main

import "github.com/askeladdk/pancake"

type screen interface {
	Begin()
	End()
	Key(pancake.KeyEvent) error
	Draw(pancake.DrawEvent) error
	Frame(pancake.FrameEvent) (screen, error)
}

type transitionScreen struct {
	To screen
}

func (s transitionScreen) Begin()                                      {}
func (s transitionScreen) End()                                        {}
func (s transitionScreen) Key(ev pancake.KeyEvent) error               { return nil }
func (s transitionScreen) Draw(ev pancake.DrawEvent) error             { return nil }
func (s transitionScreen) Frame(ev pancake.FrameEvent) (screen, error) { return s.To, nil }

type screenState struct {
	Screen screen
}

func (gs *screenState) Do(event interface{}) error {
	switch ev := event.(type) {
	case pancake.QuitEvent:
		return pancake.ErrQuit
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
