package imutil

import (
	"image"
	"image/color"
	"math/rand"
)

// RandGray creates a random grayscale image with a given seed.
func RandGray(seed int64, w, h int) *image.Gray {
	src := rand.NewSource(seed)
	// #nosec G404
	rnd := rand.New(src)

	rect := image.Rect(0, 0, w, h)
	img := image.NewGray(rect)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			px := color.Gray{Y: uint8(rnd.Int())}
			img.Set(x, y, px)
		}
	}

	return img
}

// RandRGBA creates a random RGBA image with a given seed.
func RandRGBA(seed int64, w, h int) *image.RGBA {
	src := rand.NewSource(seed)
	// #nosec G404
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
