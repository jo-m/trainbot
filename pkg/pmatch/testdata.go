package pmatch

import (
	"bytes"
	_ "embed"
	"image"
	"image/jpeg"
)

// We could just load testdata directly from the file system during tests,
// but this allows to build test binaries which are easy to ship.

//go:embed testdata/bird.jpg
var testdata []byte

func LoadTestImg() *image.YCbCr {
	buf := bytes.NewBuffer(testdata)
	img, err := jpeg.Decode(buf)
	if err != nil {
		panic("should not happen")
	}
	return img.(*image.YCbCr)
}
