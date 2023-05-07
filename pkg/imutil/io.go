package imutil

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
)

// Load tries to load an image from a file.
func Load(path string) (image.Image, error) {
	// #nosec 304
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// Dump will dump an image to a file.
// Format is determined by file ending, PNG and JPG are supported.
func Dump(path string, img image.Image) error {
	// #nosec 304
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if strings.HasSuffix(path, ".png") {
		return png.Encode(f, img)
	}

	if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 95})
	}

	return errors.New("unknown image suffix")
}

// Dump dumps a JPEG to a file with a given quality.
func DumpJPEG(path string, img image.Image, quality int) error {
	// #nosec 304
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
}

// DumpGIF will dump a GIF image to a file.
func DumpGIF(path string, img *gif.GIF) error {
	// #nosec 304
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return gif.EncodeAll(f, img)
}

// Sub tries to call SubImage on the given image.
func Sub(img image.Image, r image.Rectangle) (image.Image, error) {
	iface, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	})

	if !ok {
		return nil, errors.New("img does not implement SubImage()")
	}

	return iface.SubImage(r), nil
}
