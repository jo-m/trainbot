package pmatch

// The flags chosen below optimize for the following platforms:
// - amd64: Compiler host
// - arm64: Raspberry Pi 4
//
// To show the flags which -march=native would produce, run
//
// 	gcc -march=native -E -v - </dev/null 2>&1 | grep cc1
//
// For more details, see
// - https://gist.github.com/fm4dd/c663217935dc17f0fc73c9c81b0aa845
// - https://gcc.gnu.org/onlinedocs/gcc/x86-Options.html
// - https://gcc.gnu.org/onlinedocs/gcc/AArch64-Options.html

// #cgo CFLAGS: -Wall -Werror -Wextra -pedantic -std=c99
// #cgo CFLAGS: -O2
//
// #cgo amd64 CFLAGS: -march=x86-64 -mtune=generic
// #cgo amd64 CFLAGS: -fopenmp
// #cgo amd64 LDFLAGS: -fopenmp
//
// #cgo arm64 CFLAGS: -mcpu=cortex-a72 -mtune=cortex-a72
// #cgo arm64 CFLAGS: -fopenmp
// #cgo arm64 LDFLAGS: -fopenmp
//
// #include "c.h"
import "C"
import (
	"image"
	"math"
)

// SearchRGBAC searches for the position of an (RGBA) patch in an (RGBA) image,
// using cosine similarity.
// Implemented in Cgo.
// Panics if the patch is larger than the image in any dimension.
// The alpha channel is ignored.
func SearchRGBAC(img, pat *image.RGBA) (int, int, float64) {
	if pat.Bounds().Size().X > img.Bounds().Size().X ||
		pat.Bounds().Size().Y > img.Bounds().Size().Y {
		panic("patch too large")
	}

	// Search rect in img coordinates.
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Bounds().Size()).Add(image.Pt(1, 1)),
	}

	m, n := searchRect.Dx(), searchRect.Dy()
	du, dv := pat.Bounds().Dx(), pat.Bounds().Dy()

	is, ps := img.Stride, pat.Stride

	var maxX, maxY C.int
	var maxCos2 C.float64

	C.SearchRGBAC(
		C.int(m), C.int(n), C.int(du), C.int(dv), C.int(is), C.int(ps),
		(*C.uint8_t)(&img.Pix[0]),
		(*C.uint8_t)(&pat.Pix[0]),
		(*C.int)(&maxX),
		(*C.int)(&maxY),
		(*C.float64)(&maxCos2),
	)

	// This was left out above.
	cos := math.Sqrt(float64(maxCos2))

	return int(maxX), int(maxY), cos
}
