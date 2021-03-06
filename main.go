package main

import (
	"embed"
	"fmt"
	"image"
	"math/rand"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"

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
	globalGameScreen     *gameScreen
	globalGameOverScreen *gameOverScreen
	globalTitleScreen    *titleScreen
	globalNextScreen     *nextScreen

	//go:embed assets/*
	assets embed.FS
)

func loadTexture(filename string) (*graphics.Texture, error) {
	if f, err := assets.Open(filename); err != nil {
		return nil, err
	} else if img, _, err := image.Decode(f); err != nil {
		return nil, err
	} else {
		return graphics.NewTextureFromImage(img, graphics.FilterNearest), nil
	}
}

func loadWav(filename string) (*beep.Buffer, error) {
	if f, err := assets.Open(filename); err != nil {
		return nil, err
	} else if wav, fmt, err := wav.Decode(f); err != nil {
		return nil, err
	} else {
		buffer := beep.NewBuffer(fmt)
		buffer.Append(wav)
		return buffer, nil
	}
}

func loadMp3(filename string) (*beep.Buffer, error) {
	if f, err := assets.Open(filename); err != nil {
		return nil, err
	} else if wav, fmt, err := mp3.Decode(f); err != nil {
		return nil, err
	} else {
		buffer := beep.NewBuffer(fmt)
		buffer.Append(wav)
		return buffer, nil
	}
}

func run(app pancake.App) error {
	var sheet *graphics.Texture
	var background *graphics.Texture
	var gameover *graphics.Texture
	var title *graphics.Texture
	var nextlevel *graphics.Texture
	var sfxLaser *beep.Buffer
	var sfxExplosion *beep.Buffer
	var sfxBoing *beep.Buffer
	var mp3 *beep.Buffer
	var err error

	speaker.Init(44100, beep.SampleRate(44100).N(time.Second/10))

	rand.Seed(time.Now().Unix())

	if sheet, err = loadTexture("assets/asteroids-arcade.png"); err != nil {
		return err
	}

	if background, err = loadTexture("assets/background.png"); err != nil {
		return err
	}

	if gameover, err = loadTexture("assets/gameover.png"); err != nil {
		return err
	}

	if title, err = loadTexture("assets/title.png"); err != nil {
		return err
	}

	if nextlevel, err = loadTexture("assets/nextlevel.png"); err != nil {
		return err
	}

	if sfxLaser, err = loadWav("assets/Laser.wav"); err != nil {
		return err
	}

	if sfxExplosion, err = loadWav("assets/Explosion.wav"); err != nil {
		return err
	}

	if sfxBoing, err = loadWav("assets/Boing.wav"); err != nil {
		return err
	}

	if mp3, err = loadMp3("assets/Bonkers-for-Arcades.mp3"); err != nil {
		return err
	}
	stream := mp3.Streamer(0, mp3.Len())
	speaker.Play(beep.Loop(-1, stream))

	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	resolution := app.Resolution()
	midscreen := mathx.FromPoint(resolution.Div(2))

	drawer := graphics2d.NewDrawer(1024, nil)
	shader := graphics2d.DefaultShader()
	shader.Begin()
	shader.SetUniform("u_Projection", mathx.Ortho2D(
		0,
		float64(resolution.X),
		float64(resolution.Y),
		0,
	))
	shader.End()

	// load the font
	ttf, _ := truetype.Parse(goregular.TTF)
	face16 := truetype.NewFace(ttf, &truetype.Options{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	font16 := text.NewFontFromFace(face16, text.ASCII)

	face12 := truetype.NewFace(ttf, &truetype.Options{
		Size:    12,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	font12 := text.NewFontFromFace(face12, text.ASCII)

	text16 := text.NewText(font16)
	text12 := text.NewText(font12)

	simulation := theSimulation{
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
		Sounds: []*beep.Buffer{
			sfxLaser,
			sfxExplosion,
			sfxBoing,
		},
		Bounds: mathx.Rectangle{
			Min: mathx.Vec2{},
			Max: mathx.FromPoint(resolution),
		},
	}

	globalGameScreen = &gameScreen{
		Sim:    &simulation,
		Text:   text16,
		Drawer: drawer,
		Shader: shader,
		Background: staticImage{
			Image:    background,
			Position: midscreen,
		},
	}

	globalGameOverScreen = &gameOverScreen{
		Sim:    &simulation,
		Text:   text16,
		Drawer: drawer,
		Shader: shader,
		Background: staticImage{
			Image:    background,
			Position: midscreen,
		},
		Title: staticImage{
			Image:    gameover,
			Position: midscreen,
		},
	}

	globalTitleScreen = &titleScreen{
		Drawer: drawer,
		Shader: shader,
		Text:   text12,
		Background: staticImage{
			Image:    background,
			Position: midscreen,
		},
		Title: staticImage{
			Image:    title,
			Position: midscreen,
		},
	}

	globalNextScreen = &nextScreen{
		Sim:    &simulation,
		Text:   text16,
		Drawer: drawer,
		Shader: shader,
		Background: staticImage{
			Image:    background,
			Position: midscreen,
		},
		Title: staticImage{
			Image:    nextlevel,
			Position: midscreen,
		},
	}

	sscreenState := screenState{
		Screen: transitionScreen{
			To: globalTitleScreen,
		},
	}

	return app.Events(func(event interface{}) error {
		app.SetTitle(fmt.Sprintf("Asteroids (%d FPS)", app.FrameRate()))
		return sscreenState.Do(event)
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
