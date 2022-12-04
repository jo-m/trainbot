package imutil

import (
	"image"
	"image/draw"
)

func ToGray(img image.Image) *image.Gray {
	ret := image.NewGray(img.Bounds())
	draw.Draw(ret, ret.Bounds(), img, img.Bounds().Min, draw.Src)

	return ret
}

func GrayReset0(img *image.Gray) *image.Gray {
	ret := image.NewGray(img.Bounds().Sub(img.Bounds().Min))
	draw.Draw(ret, ret.Bounds(), img, img.Bounds().Min, draw.Src)

	return ret
}
