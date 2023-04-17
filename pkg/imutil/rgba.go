package imutil

import (
	"image"
	"image/draw"
)

// ToRGBA returns a new copy of img, converted to RGBA, with the origin of its bounds reset to 0.
func ToRGBA(img image.Image) *image.RGBA {
	ret := image.NewRGBA(img.Bounds().Sub(img.Bounds().Min))
	draw.Draw(ret, ret.Bounds(), img, img.Bounds().Min, draw.Src)

	return ret
}
