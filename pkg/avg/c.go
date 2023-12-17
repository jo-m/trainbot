package avg

// See pkg/pmatch/c.go for docs on CC flags.

// #cgo CFLAGS: -Wall -Wextra -pedantic -std=c99
// #cgo CFLAGS: -O2
//
// #cgo amd64 CFLAGS: -march=native
//
// #cgo arm64 CFLAGS: -mcpu=cortex-a72 -mtune=cortex-a72
//
// #include "c.h"
import "C"
import "image"

// RGBAC computes the pixel average, and pixel mean deviation from average,
// on an RGBA image, per channel.
// Note that the alpha channel is ignored.
// Scaled to [0, 1].
// Implemented in Cgo.
func RGBAC(img *image.RGBA) ([3]float64, [3]float64) {

	m, n := img.Bounds().Dx(), img.Bounds().Dy()
	s := img.Stride

	ret := C.retData{}
	C.RGBAC(C.int(m), C.int(n), C.int(s),
		(*C.uint8_t)(&img.Pix[0]), (*C.retData)(&ret))

	return [3]float64{
			float64(ret.avg[0]),
			float64(ret.avg[1]),
			float64(ret.avg[2]),
		},
		[3]float64{
			float64(ret.avgDev[0]),
			float64(ret.avgDev[1]),
			float64(ret.avgDev[2]),
		}
}
