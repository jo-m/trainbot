package pmatch

import (
	"bytes"
	_ "embed"
	"image"
	"image/jpeg"
)

//go:embed testdata/bird.jpg
var testdata []byte

// LoadTestImg returns a test image to be used for tests and examples.
// Image data is embedded in the binary, which makes this function independent of the environment.
// Always returns a newly allocated image.
func LoadTestImg() *image.YCbCr {
	buf := bytes.NewBuffer(testdata)
	img, err := jpeg.Decode(buf)
	if err != nil {
		panic("should not happen")
	}
	return img.(*image.YCbCr)
}
