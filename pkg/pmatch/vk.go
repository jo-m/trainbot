package pmatch

/*

intel-gpu-tools

go generate ./...; go test -run Test_Run ./pkg/pmatch
*/

// #cgo CFLAGS: -std=c99
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

func bufsz(img *image.RGBA) int {
	return img.Bounds().Dy() * img.Stride
}

const (
	localSizeX = 4
	localSizeY = 4
	localSizeZ = 1
)

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

	C.prepare(
		C.size_t(bufsz(img)),
		C.size_t(bufsz(pat)),
		C.size_t(bufsz(search)),
		(*C.uint8_t)(unsafe.Pointer(&shaderCode[0])),
		C.size_t(uint64(len(shaderCode))),
		C.dim3{
			C.uint32_t(searchRect.Dx()/localSizeX + 1),
			C.uint32_t(searchRect.Dy()/localSizeY + 1),
			C.uint32_t(1),
		})

	dims := C.dimensions{
		m:        C.uint32_t(searchRect.Dx()),
		n:        C.uint32_t(searchRect.Dy()),
		du:       C.uint32_t(pat.Bounds().Dx()),
		dv:       C.uint32_t(pat.Bounds().Dy()),
		is:       C.uint32_t(img.Stride),
		ps:       C.uint32_t(pat.Stride),
		ss:       C.uint32_t(search.Stride),
		max_uint: C.uint32_t(0),
		max:      C.float(0),
		max_x:    C.uint32_t(0),
		max_y:    C.uint32_t(0),
	}

	C.run(
		(*C.dimensions)(unsafe.Pointer(&dims)),
		(*C.uint8_t)(unsafe.Pointer(&img.Pix[0])),
		(*C.uint8_t)(unsafe.Pointer(&pat.Pix[0])),
		(*C.uint8_t)(unsafe.Pointer(&search.Pix[0])),
		C.dim3{C.uint32_t(localSizeX), C.uint32_t(localSizeY), C.uint32_t(localSizeZ)})

	C.cleanup()

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	imutil.Dump(filepath.Join(home, "Desktop/img.png"), img)
	imutil.Dump(filepath.Join(home, "Desktop/pat.png"), pat)
	imutil.Dump(filepath.Join(home, "Desktop/search.png"), search)

	fmt.Printf("%+v\n", dims)

	return int(dims.max_x), int(dims.max_y), float64(dims.max)
}
