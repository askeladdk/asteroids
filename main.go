//go:generate go-bindata -nocompress assets

package main

import (
	"bytes"
	"fmt"
	"image"
	"math/rand"
	"time"

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

var (
	gameScreen     *GameScreen
	gameOverScreen *GameOverScreen
	titleScreen    *TitleScreen
)

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
	var gameover *graphics.Texture
	var title *graphics.Texture

	rand.Seed(time.Now().Unix())

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

	if tex, err := loadTexture("assets/gameover.png"); err != nil {
		return err
	} else {
		gameover = tex
	}

	if tex, err := loadTexture("assets/title.png"); err != nil {
		return err
	} else {
		title = tex
	}

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
		ImageAtlas: sheet,
		Images: []graphics.Image{
			sheet.SubImage(image.Rect(0, 0, 32, 32)),       // spaceship
			sheet.SubImage(image.Rect(64, 192, 128, 256)),  // asteroid
			sheet.SubImage(image.Rect(112, 64, 128, 80)),   // bullet
			sheet.SubImage(image.Rect(128, 192, 160, 224)), // debris
			sheet.SubImage(image.Rect(160, 192, 192, 224)),
			sheet.SubImage(image.Rect(128, 224, 160, 256)),
			sheet.SubImage(image.Rect(160, 224, 192, 256)),
		},
		Bounds: mathx.Rectangle{
			mathx.Vec2{},
			mathx.FromPoint(resolution),
		},
	}

	gameScreen = &GameScreen{
		Sim:    &simulation,
		Text:   text.NewText(thefont),
		Drawer: drawer,
		Shader: shader,
		Background: StaticImage{
			Image:    background,
			Position: midscreen,
		},
	}

	gameOverScreen = &GameOverScreen{
		Sim:    &simulation,
		Text:   text.NewText(thefont),
		Drawer: drawer,
		Shader: shader,
		Background: StaticImage{
			Image:    background,
			Position: midscreen,
		},
		Title: StaticImage{
			Image:    gameover,
			Position: midscreen,
		},
	}

	titleScreen = &TitleScreen{
		Drawer: drawer,
		Shader: shader,
		Background: StaticImage{
			Image:    background,
			Position: midscreen,
		},
		Title: StaticImage{
			Image:    title,
			Position: midscreen,
		},
	}

	screenState := ScreenState{
		Screen: TransitionScreen{
			To: titleScreen,
		},
	}

	return app.Events(func(event interface{}) error {
		app.SetTitle(fmt.Sprintf("Asteroids (%d FPS)", app.FrameRate()))
		return screenState.Do(event)
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
