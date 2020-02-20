//go:generate go-bindata assets

package main

import (
	"bytes"
	"fmt"
	"image"

	_ "image/png"

	"github.com/askeladdk/pancake/graphics2d"
	"github.com/askeladdk/pancake/text"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/askeladdk/pancake/graphics"
	gl "github.com/askeladdk/pancake/graphics/opengl"
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
	var background *graphics.Texture

	if tex, err := loadTexture("assets/asteroids-arcade.png"); err != nil {
		return err
	} else {
		sheet = tex
	}

	if tex, err := loadTexture("assets/background.png"); err != nil {
		return err
	} else {
		background = tex
	}

	ship := sheet.SubImage(image.Rect(0, 0, 32, 32))
	asteroid := sheet.SubImage(image.Rect(64, 192, 128, 256))
	bullet := sheet.SubImage(image.Rect(112, 64, 128, 80))

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	resolution := app.Resolution()
	midscreen := mathx.FromPoint(resolution.Div(2))

	drawer := graphics2d.NewDrawer(1024, graphics2d.Quad)
	shader := graphics2d.DefaultShader()
	shader.Begin()
	shader.SetUniform("u_Projection", mathx.Ortho2D(
		0,
		float32(resolution.X),
		float32(resolution.Y),
		0,
	))
	shader.End()

	// load the font
	ttf, _ := truetype.Parse(goregular.TTF)
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	thefont := text.NewFontFromFace(face, text.ASCII)

	simulation := Simulation{
		sprites: []graphics.Image{
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
				mask:    SPACESHIP,
				radius:  14,
			},
		},
	}

	gameScreen = &GameScreen{
		app:     app,
		sim:     &simulation,
		fpstext: text.NewText(thefont),
		drawer:  drawer,
		shader:  shader,
		background: StaticImage{
			Image:    background,
			Position: midscreen,
		},
	}

	screenState := ScreenState{
		Screen: TransitionScreen{
			To: gameScreen,
		},
	}

	return app.Events(screenState.Do)
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
