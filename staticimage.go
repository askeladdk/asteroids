package main

import (
	"image/color"

	"github.com/askeladdk/pancake/graphics"
	"github.com/askeladdk/pancake/mathx"
)

type StaticImage struct {
	Image    *graphics.Texture
	Position mathx.Vec2
}

func (si StaticImage) Len() int {
	return 1
}

func (si StaticImage) ColorAt(i int) color.NRGBA {
	return color.NRGBA{255, 255, 255, 255}
}

func (si StaticImage) Texture() *graphics.Texture {
	return si.Image
}

func (si StaticImage) TextureRegionAt(i int) mathx.Aff3 {
	return mathx.IdentAff3()
}

func (si StaticImage) ModelViewAt(i int) mathx.Aff3 {
	return mathx.
		ScaleAff3(mathx.FromPoint(si.Image.Bounds().Size())).
		Translated(si.Position)
}

func (si StaticImage) PivotAt(i int) mathx.Vec2 {
	return mathx.Vec2{}
}
