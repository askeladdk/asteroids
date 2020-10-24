package main

import (
	"image/color"

	"github.com/askeladdk/pancake/graphics"
	"github.com/askeladdk/pancake/mathx"
)

type StaticImage struct {
	Image    graphics.Image
	Position mathx.Vec2
}

func (si StaticImage) Len() int {
	return 1
}

func (si StaticImage) TintColorAt(i int) color.Color {
	return color.RGBA{255, 255, 255, 255}
}

func (si StaticImage) TextureAt(i int) *graphics.Texture {
	return si.Image.Texture()
}

func (si StaticImage) TextureRegionAt(i int) graphics.TextureRegion {
	return si.Image.TextureRegion()
}

func (si StaticImage) ModelViewAt(i int) mathx.Aff3 {
	return mathx.
		ScaleAff3(si.Image.Scale()).
		Translated(si.Position)
}

func (si StaticImage) OriginAt(i int) mathx.Vec2 {
	return mathx.Vec2{}
}

func (si StaticImage) ZOrderAt(i int) float64 {
	return 0
}
