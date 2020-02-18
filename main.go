//go:generate go-bindata assets

package main

import (
	"bytes"
	"fmt"
	"image"

	"image/color"
	_ "image/png"

	"github.com/askeladdk/pancake/graphics2d"
	"github.com/askeladdk/pancake/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/askeladdk/pancake/graphics"
	gl "github.com/askeladdk/pancake/graphics/opengl"
	"github.com/askeladdk/pancake/input"
	"github.com/askeladdk/pancake/mathx"

	"github.com/askeladdk/pancake"
)

func toggleFlag(flags uint32, flag uint32, state bool) uint32 {
	if state {
		return flags | flag
	} else {
		return flags &^ flag
	}
}

func loadTexture(filename string) (*graphics.Texture, error) {
	if data, err := Asset(filename); err != nil {
		return nil, err
	} else if img, _, err := image.Decode(bytes.NewBuffer(data)); err != nil {
		return nil, err
	} else {
		return graphics.NewTextureFromImage(img, graphics.FilterNearest), nil
	}
}

func run(app pancake.App) error {
	var sheet *graphics.Texture
	var background graphics2d.Sprite

	if tex, err := loadTexture("assets/asteroids-arcade.png"); err != nil {
		return err
	} else {
		sheet = tex
	}

	if tex, err := loadTexture("assets/background.png"); err != nil {
		return err
	} else {
		background = graphics2d.NewSprite(tex, tex.Bounds())
	}

	ship := graphics2d.NewSprite(sheet, image.Rect(0, 0, 32, 32))
	asteroid := graphics2d.NewSprite(sheet, image.Rect(64, 192, 128, 256))
	bullet := graphics2d.NewSprite(sheet, image.Rect(112, 64, 128, 80))

	resolution := app.Bounds().Size()
	projection := mathx.Ortho2D(
		0,
		float32(resolution.X),
		float32(resolution.Y),
		0,
	)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	shipbrain := inputBrain{}

	midscreen := mathx.FromPoint(app.Bounds().Size().Div(2))

	drawer := graphics2d.NewDrawer(1024, graphics2d.Quad)
	shader := graphics2d.DefaultShader()

	// load the font
	ttf, _ := truetype.Parse(goregular.TTF)
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	thefont := text.NewFontFromFace(face, text.ASCII)

	fpstext := text.NewText(thefont)

	simulation := simulation{
		sprites: []graphics2d.Sprite{
			background,
			ship,
			asteroid,
			bullet,
		},
		bounds: mathx.Rectangle{
			mathx.Vec2{},
			mathx.FromPoint(resolution),
		},
		entities: []entity{
			entity{
				sprite: background,
				pos0:   midscreen,
				pos:    midscreen,
			},
			entity{
				sprite:  ship,
				pos0:    midscreen,
				pos:     midscreen,
				rot:     -mathx.Tau / 4,
				minrotv: 1,
				maxv:    300,
				turn:    mathx.Tau / 4,
				thrust:  100,
				dampenr: 0.95,
				dampenv: 0.99,
				brain:   shipbrain.frame,
				mask:    SPACESHIP,
				radius:  14,
			},
		},
	}

	var keys uint32

	return app.Events(func(event interface{}) error {
		switch ev := event.(type) {
		case pancake.QuitEvent:
			return pancake.Quit
		case pancake.KeyEvent:
			switch ev.Key {
			case input.KeyEscape:
				return pancake.Quit
			case input.KeyA:
				fallthrough
			case input.KeyLeft:
				keys = toggleFlag(keys, 1, ev.Flags.Down())
			case input.KeyD:
				fallthrough
			case input.KeyRight:
				keys = toggleFlag(keys, 2, ev.Flags.Down())
			case input.KeyW:
				fallthrough
			case input.KeyUp:
				keys = toggleFlag(keys, 4, ev.Flags.Down())
			case input.KeyP:
				if ev.Flags.Pressed() {
					simulation.spawnAsteroid()
				}
			case input.KeySpace:
				if ev.Flags.Pressed() {
					e := simulation.at(1)
					simulation.spawnBullet(e.pos, e.rot)
				}
			}
		case pancake.FrameEvent:
			if keys&3 == 1 {
				shipbrain.action(TURN, -1)
			} else if keys&3 == 2 {
				shipbrain.action(TURN, +1)
			}

			if keys&4 != 0 {
				shipbrain.action(FORWARD, 1)
			}

			simulation.frame(float32(ev.DeltaTime))

			fpstext.Clear()
			fpstext.Color = color.NRGBA{255, 255, 255, 255}
			fmt.Fprintf(fpstext, "FPS: ")
			fpstext.Color = color.NRGBA{255, 0, 0, 255}
			fmt.Fprintf(fpstext, "%d", app.FrameRate())
		case pancake.DrawEvent:
			var batches []graphics2d.Batch
			batches = append(batches, simulation.batches(float32(ev.Alpha))...)
			batches = append(batches, fpstext)

			app.Begin()
			gl.ClearColor(0, 0, 1, 0)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			shader.Begin()
			shader.SetUniform("u_Projection", projection)
			drawer.DrawBatches(batches)
			shader.End()

			app.End()
		}

		return nil
	})
}

func main() {
	opt := pancake.Options{
		WindowSize: image.Point{960, 540},
		Resolution: image.Point{640, 360},
		Title:      "Asteroids",
		FrameRate:  60,
	}

	if err := pancake.Main(opt, run); err != nil {
		fmt.Println(err)
	}
}
