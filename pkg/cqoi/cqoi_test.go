package cqoi

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func randRGBA(seed int64, w, h int) *image.RGBA {
	src := rand.NewSource(seed)
	rnd := rand.New(src)

	rect := image.Rect(0, 0, w, h)
	img := image.NewRGBA(rect)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			px := color.RGBA{uint8(rnd.Int()), uint8(rnd.Int()), uint8(rnd.Int()), uint8(rnd.Int())}
			img.Set(x, y, px)
		}
	}

	return img
}
func Test_Dump_Load(t *testing.T) {
	img := randRGBA(123, 200, 500)
	dir := t.TempDir()
	fname := path.Join(dir, "img.qoi")

	err := Dump(fname, img)
	assert.NoError(t, err)

	dec, err := Load(fname)
	assert.NoError(t, err)

	assert.Equal(t, img, dec)
}

func Benchmark_Dump_QOI(b *testing.B) {
	dir := b.TempDir()
	img := randRGBA(123, 200, 500)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fname := path.Join(dir, fmt.Sprintf("out_%05d.qoi", i))

		b.StartTimer()
		err := Dump(fname, img)
		b.StopTimer()

		if err != nil {
			b.Error(err)
		}
	}
}
