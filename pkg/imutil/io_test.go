package imutil

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getFileSize(t *testing.T, path string) int {
	stat, err := os.Stat(path)
	require.NoError(t, err)

	return int(stat.Size())
}

func Test_Dump_Size_PNG(t *testing.T) {
	dir := t.TempDir()
	fname := filepath.Join(dir, "out.png")
	img := RandRGBA(123, 200, 500)

	err := Dump(fname, img)
	assert.NoError(t, err)
	assert.Equal(t, 400772, getFileSize(t, fname))
}
func Test_Dump_Size_JPG(t *testing.T) {
	dir := t.TempDir()
	fname := filepath.Join(dir, "out.jpg")
	img := RandRGBA(123, 200, 500)

	err := Dump(fname, img)
	assert.NoError(t, err)
	err = Dump(fname, img)
	assert.NoError(t, err)
	assert.Equal(t, 121978, getFileSize(t, fname))
}

func Benchmark_Dump_PNG(b *testing.B) {
	dir := b.TempDir()
	img := RandRGBA(123, 200, 500)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fname := filepath.Join(dir, fmt.Sprintf("out_%05d.png", i))

		b.StartTimer()
		err := Dump(fname, img)
		b.StopTimer()

		if err != nil {
			b.Error(err)
		}
	}
}

func Benchmark_Dump_JPG(b *testing.B) {
	dir := b.TempDir()
	img := RandRGBA(123, 200, 500)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fname := filepath.Join(dir, fmt.Sprintf("out_%05d.jpg", i))

		b.StartTimer()
		err := Dump(fname, img)
		b.StopTimer()

		if err != nil {
			b.Error(err)
		}
	}
}
