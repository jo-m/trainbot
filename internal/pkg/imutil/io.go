package imutil

import (
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
)

func Load(path string) (image.Image, error) {
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

func Dump(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if strings.HasSuffix(path, ".png") {
		return png.Encode(f, img)
	}

	if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
	}

	return errors.New("unknown image suffix")
}

func Sub(img image.Image, r image.Rectangle) (image.Image, error) {
	iface, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	})

	if !ok {
		return nil, errors.New("img does not implement SubImage()")
	}

	return iface.SubImage(r), nil
}
