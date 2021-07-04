package main

import (
	"image/color"

	"github.com/askeladdk/pancake/graphics"
	"github.com/askeladdk/pancake/mathx"
)

type staticImage struct {
	Image    graphics.Image
	Position mathx.Vec2
}

func (si staticImage) Len() int {
	return 1
}

func (si staticImage) TintColorAt(i int) color.Color {
	return color.RGBA{255, 255, 255, 255}
}

func (si staticImage) TextureAt(i int) *graphics.Texture {
	return si.Image.Texture()
}

func (si staticImage) TextureRegionAt(i int) graphics.TextureRegion {
	return si.Image.TextureRegion()
}

func (si staticImage) ModelViewAt(i int) mathx.Aff3 {
	return mathx.
		ScaleAff3(si.Image.Scale()).
		Translated(si.Position)
}

func (si staticImage) OriginAt(i int) mathx.Vec2 {
	return mathx.Vec2{}
}

func (si staticImage) ZOrderAt(i int) float64 {
	return 0
}
