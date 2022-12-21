package cqoi

import (
	"image"
	"image/color"
	"math/rand"
)

func RandRGBA(seed int64, w, h int) *image.RGBA {
	src := rand.NewSource(seed)
	rnd := rand.New(src)

	rect := image.Rect(0, 0, w, h)
	img := image.NewRGBA(rect)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			px := color.RGBA{uint8(rnd.Int()), uint8(rnd.Int()), uint8(rnd.Int()), uint8(rnd.Int())}
			img.Set(x, y, px)
		}
	}

	return img
}
