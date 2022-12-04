package imutil

import (
	"image"
	"image/draw"
)

func ToRGBA(img image.Image) *image.RGBA {
	ret := image.NewRGBA(img.Bounds())
	draw.Draw(ret, ret.Bounds(), img, img.Bounds().Min, draw.Src)

	return ret
}

func RGBAReset0(img *image.RGBA) *image.RGBA {
	ret := image.NewRGBA(img.Bounds().Sub(img.Bounds().Min))
	draw.Draw(ret, ret.Bounds(), img, img.Bounds().Min, draw.Src)

	return ret
}
