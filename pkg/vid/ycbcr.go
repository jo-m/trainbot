package vid

import (
	"image"
	"image/color"
)

// YCbCr is a raw YCbCr image.
type YCbCr struct {
	Pix  []byte
	Rect image.Rectangle
}

// Compile time interface check.
var _ image.Image = (*YCbCr)(nil)

// ColorModel implements image.Image.
func (i *YCbCr) ColorModel() color.Model {
	return color.YCbCrModel
}

// Bounds implements image.Image.
func (i *YCbCr) Bounds() image.Rectangle {
	return i.Rect
}

// At implements image.Image.
func (i *YCbCr) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(i.Rect)) {
		return color.YCbCr{}
	}
	pixIx := (y*i.Rect.Max.X + x)
	// Two pixels = four bytes.
	cellIx := (pixIx / 2) * 4

	Y0, Cb, Y1, Cr := i.Pix[cellIx+0], i.Pix[cellIx+1], i.Pix[cellIx+2], i.Pix[cellIx+3]

	Y := Y0
	if pixIx%2 == 1 {
		Y = Y1
	}

	return color.YCbCr{
		Y, Cb, Cr,
	}
}
