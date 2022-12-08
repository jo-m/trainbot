package vid

import (
	"image"
	"image/color"
)

type YCbCr struct {
	rect image.Rectangle
	buf  []byte
}

// compile time interface check
var _ image.Image = (*YCbCr)(nil)

func (i *YCbCr) ColorModel() color.Model {
	return color.YCbCrModel
}
func (i *YCbCr) Bounds() image.Rectangle {
	return i.rect
}
func (i *YCbCr) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(i.rect)) {
		return color.YCbCr{}
	}
	pixIx := (y*i.rect.Max.X + x)
	// two pixels = four bytes
	cellIx := (pixIx / 2) * 4

	Y0, Cb, Y1, Cr := i.buf[cellIx+0], i.buf[cellIx+1], i.buf[cellIx+2], i.buf[cellIx+3]

	Y := Y0
	if pixIx%2 == 1 {
		Y = Y1
	}

	return color.YCbCr{
		Y, Cb, Cr,
	}
}
