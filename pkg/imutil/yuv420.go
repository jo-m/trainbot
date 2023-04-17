package imutil

import (
	"image"
)

// NewYuv420 creates a new image.Image from a raw Yuv420p buffer.
// The returned image will reference the passed buffer and no data is copied.
func NewYuv420(buf []byte, w, h int) image.Image {
	im := image.NewYCbCr(image.Rect(0, 0, w, h), image.YCbCrSubsampleRatio420)
	im.Y = buf
	im.Cb = buf[len(buf)/12*8:]
	im.Cr = buf[len(buf)/12*10:]

	return im
}
