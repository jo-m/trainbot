package imutil

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path"
	"strings"
	"testing"

	qoi3 "github.com/arian/go-qoi"
	"github.com/jo-m/trainbot/pkg/cqoi"
	qoi2 "github.com/takeyourhatoff/qoi"
	qoi1 "github.com/xfmoulet/qoi"
)

func dumpQOI(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if strings.HasSuffix(path, ".qoi1") {
		return qoi1.Encode(f, img)
	}

	if strings.HasSuffix(path, ".qoi2") {
		return qoi2.Encode(f, img)
	}

	if strings.HasSuffix(path, ".qoi3") {
		return qoi3.Encode(f, img)
	}

	return errors.New("unknown image suffix")
}

func Benchmark_Dump_QOI1(b *testing.B) {
	dir := b.TempDir()
	img := RandRGBA(123, 200, 500)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fname := path.Join(dir, fmt.Sprintf("out_%05d.qoi1", i))

		b.StartTimer()
		err := dumpQOI(fname, img)
		b.StopTimer()

		if err != nil {
			b.Error(err)
		}
	}
}

func Benchmark_Dump_QOI2(b *testing.B) {
	dir := b.TempDir()
	img := RandRGBA(123, 200, 500)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fname := path.Join(dir, fmt.Sprintf("out_%05d.qoi2", i))

		b.StartTimer()
		err := dumpQOI(fname, img)
		b.StopTimer()

		if err != nil {
			b.Error(err)
		}
	}
}

func Benchmark_Dump_QOI3(b *testing.B) {
	dir := b.TempDir()
	img := RandRGBA(123, 200, 500)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fname := path.Join(dir, fmt.Sprintf("out_%05d.qoi3", i))

		b.StartTimer()
		err := dumpQOI(fname, img)
		b.StopTimer()

		if err != nil {
			b.Error(err)
		}
	}
}

func Benchmark_Dump_CQOI(b *testing.B) {
	dir := b.TempDir()
	img := RandRGBA(123, 200, 500)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fname := path.Join(dir, fmt.Sprintf("out_%05d.qoi3", i))

		b.StartTimer()
		err := cqoi.Dump(fname, img)
		b.StopTimer()

		if err != nil {
			b.Error(err)
		}
	}
}
