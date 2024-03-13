package imutil

import (
	"image"
)

// Copy creates a deep copy of an image (copies underlying buffer).
func Copy(in image.Image) (ret image.Image) {
	switch i := in.(type) {
	case *image.Gray:
		cp := *i
		cp.Pix = make([]uint8, len(i.Pix))
		copy(cp.Pix, i.Pix)
		ret = &cp
	case *image.RGBA:
		cp := *i
		cp.Pix = make([]uint8, len(i.Pix))
		copy(cp.Pix, i.Pix)
		ret = &cp
	case *image.YCbCr:
		cp := *i
		cp.Y = make([]uint8, len(i.Y))
		copy(cp.Y, i.Y)
		cp.Cb = make([]uint8, len(i.Cb))
		copy(cp.Cb, i.Cb)
		cp.Cr = make([]uint8, len(i.Cr))
		copy(cp.Cr, i.Cr)
		ret = &cp
	case *YCbCr:
		cp := *i
		cp.Pix = make([]uint8, len(i.Pix))
		copy(cp.Pix, i.Pix)
		ret = &cp
	default:
		panic("not implemented")
	}
	return
}
