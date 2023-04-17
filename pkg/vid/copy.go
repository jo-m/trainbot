package vid

import (
	"image"
)

func imCopy(in image.Image) (dst image.Image) {
	switch i := in.(type) {
	case *image.Gray:
		cp := *i
		cp.Pix = make([]uint8, len(i.Pix))
		copy(cp.Pix, i.Pix)
		dst = &cp
	case *image.RGBA:
		cp := *i
		cp.Pix = make([]uint8, len(i.Pix))
		copy(cp.Pix, i.Pix)
		dst = &cp
	case *image.YCbCr:
		cp := *i
		cp.Y = make([]uint8, len(i.Y))
		copy(cp.Y, i.Y)
		cp.Cb = make([]uint8, len(i.Cb))
		copy(cp.Cb, i.Cb)
		cp.Cr = make([]uint8, len(i.Cr))
		copy(cp.Cr, i.Cr)
		dst = &cp
	case *YCbCr:
		cp := *i
		cp.Pix = make([]uint8, len(i.Pix))
		copy(cp.Pix, i.Pix)
		dst = &cp
	default:
		panic("not implemented")
	}
	return
}
