package cqoi

// See pmatch/c.go for details on the chosen compiler flags.

// #cgo CFLAGS: -Wall -Werror -Wextra -pedantic -std=c99
// #cgo CFLAGS: -O2
// #cgo CFLAGS: -DQOI_IMPLEMENTATION
//
// #cgo amd64 CFLAGS: -march=native
//
// #cgo arm64 CFLAGS: -mcpu=cortex-a72 -mtune=cortex-a72
//
// #include "qoi.h"
import "C"
import (
	"errors"
	"image"
	"unsafe"
)

// Dump wraps qoi_write().
func Dump(path string, img image.Image) error {
	rgba, ok := img.(*image.RGBA)
	if !ok {
		return errors.New("img must be image.RGBA")
	}

	sz := rgba.Rect.Size()
	if rgba.Stride != sz.X*4 {
		return errors.New("invalid stride, must be width * 4")
	}

	desc := C.qoi_desc{
		width:      C.uint(sz.X),
		height:     C.uint(sz.Y),
		channels:   C.uchar(4), // RGBA
		colorspace: C.QOI_LINEAR,
	}

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	written := C.qoi_write(cpath, unsafe.Pointer(&rgba.Pix[0]), &desc)

	if written == 0 {
		return errors.New("qoi_write() returned NULL")
	}
	return nil
}

// Load wraps qoi_read().
func Load(path string) (image.Image, error) {
	desc := C.qoi_desc{}

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	pix := C.qoi_read(cpath, &desc, 4) // RGBA
	if pix == nil {
		return nil, errors.New("qoi_read() returned NULL")
	}
	defer C.free(pix)

	ret := image.NewRGBA(image.Rect(0, 0, int(desc.width), int(desc.height)))
	if ret.Stride != int(desc.width)*4 {
		panic("invalid stride")
	}

	slc := unsafe.Slice((*byte)(pix), desc.width*desc.height*4)
	copy(ret.Pix, slc)

	return ret, nil
}
