package imutil

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math/rand"
	"os"
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

	return png.Encode(f, img)
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

func ToGray(img image.Image) *image.Gray {
	ret := image.NewGray(img.Bounds())
	draw.Draw(ret, ret.Bounds(), img, img.Bounds().Min, draw.Src)

	return ret
}

func GrayReset0(img *image.Gray) *image.Gray {
	ret := image.NewGray(img.Bounds().Sub(img.Bounds().Min))
	draw.Draw(ret, ret.Bounds(), img, img.Bounds().Min, draw.Src)

	return ret
}

func Rand(seed int64, w, h int) *image.Gray {
	src := rand.NewSource(seed)
	rnd := rand.New(src)

	rect := image.Rect(0, 0, w, h)
	img := image.NewGray(rect)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.Gray{Y: uint8(rnd.Int())})
		}
	}

	return img
}
