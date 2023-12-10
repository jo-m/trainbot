package pmatch

// #cgo CFLAGS: -Wall -Wextra -pedantic -std=c99
// #cgo CFLAGS: -O2
//
// #cgo LDFLAGS: -lvulkan
//
// #include "vk.h"
import "C"
import (
	_ "embed"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/jo-m/trainbot/pkg/imutil"
)

//go:generate glslangValidator -V -o shader.spv shader.comp

//go:embed shader.spv
var shaderCode []byte

const (
	localSizeX = 4
	localSizeY = 4
	localSizeZ = 1
)

func bufsz(img *image.RGBA) int {
	return img.Bounds().Dy() * img.Stride
}

func SearchRGBAVk(img, pat *image.RGBA) (maxX, maxY int, maxCos float64) {
	if pat.Bounds().Size().X > img.Bounds().Size().X ||
		pat.Bounds().Size().Y > img.Bounds().Size().Y {
		panic("patch too large")
	}

	// Search rect in img coordinates.
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Bounds().Size()).Add(image.Pt(1, 1)),
	}

	search := image.NewRGBA(searchRect)

	specConstants := []C.int32_t{
		// Local size.
		C.int32_t(localSizeX),
		C.int32_t(localSizeY),
		C.int32_t(localSizeZ),
		// Dimensions constants.
		C.int32_t(searchRect.Dx()),
		C.int32_t(searchRect.Dy()),
		C.int32_t(pat.Bounds().Dx()),
		C.int32_t(pat.Bounds().Dy()),
		C.int32_t(img.Stride),
		C.int32_t(pat.Stride),
		C.int32_t(search.Stride),
	}

	C.prepare(
		C.size_t(bufsz(img)),
		C.size_t(bufsz(pat)),
		C.size_t(bufsz(search)),
		(*C.uint8_t)(unsafe.Pointer(&shaderCode[0])),
		C.size_t(uint64(len(shaderCode))),
		(*C.int32_t)(unsafe.Pointer(&specConstants[0])),
		C.uint32_t(len(specConstants)),
	)

	res := C.results{}

	C.run(
		(*C.results)(unsafe.Pointer(&res)),
		(*C.uint8_t)(unsafe.Pointer(&img.Pix[0])),
		(*C.uint8_t)(unsafe.Pointer(&pat.Pix[0])),
		(*C.uint8_t)(unsafe.Pointer(&search.Pix[0])),
		// Workgroup size.
		C.dim3{
			C.uint32_t(searchRect.Dx()/localSizeX + 1),
			C.uint32_t(searchRect.Dy()/localSizeY + 1),
			C.uint32_t(1),
		})

	C.cleanup()

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	imutil.Dump(filepath.Join(home, "Desktop/img.png"), img)
	imutil.Dump(filepath.Join(home, "Desktop/pat.png"), pat)
	imutil.Dump(filepath.Join(home, "Desktop/search.png"), search)

	fmt.Printf("%+v\n", res)
	return int(res.max_x), int(res.max_y), float64(res.max)
}
