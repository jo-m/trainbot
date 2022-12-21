package cqoi

import (
	"fmt"
	"path"
	"testing"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/stretchr/testify/assert"
)

func Test_Dump_Load(t *testing.T) {
	img := imutil.RandRGBA(123, 200, 500)
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
	img := imutil.RandRGBA(123, 200, 500)

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
