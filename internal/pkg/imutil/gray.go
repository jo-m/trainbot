package imutil

import (
	"image"
	"image/draw"
)

// ToGray returns a new copy of img, converted to grayscale, with the origin of its bounds reset to 0.
func ToGray(img image.Image) *image.Gray {
	ret := image.NewGray(img.Bounds().Sub(img.Bounds().Min))
	draw.Draw(ret, ret.Bounds(), img, img.Bounds().Min, draw.Src)

	return ret
}
