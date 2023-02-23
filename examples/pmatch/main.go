// Package main (pmatch) is a simple binary to demonstrate the usage of the pmatch package.
package main

import (
	"fmt"
	"image"
	"time"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/pkg/pmatch"
)

const (
	px      = 20
	py      = 40
	patchSz = 50
)

func bench(fn func(), count int) {
	t0 := time.Now()
	for i := 0; i < count; i++ {
		fn()
	}
	dt := time.Since(t0)
	fmt.Printf("%d iterations took %fs -> %fms/iter\n", count, dt.Seconds(), dt.Seconds()/float64(count)*1000)
}

func main() {
	img := imutil.ToGray(pmatch.LoadTestImg())
	rect := image.Rect(px, py, patchSz+px, patchSz+py)
	pat, err := imutil.Sub(img, rect)
	if err != nil {
		panic(err)
	}

	fn := pmatch.SearchGraySlow

	// test
	x, y, score := fn(img, pat.(*image.Gray))
	fmt.Printf("x=%d y=%d score=%f\n", x, y, score)
	if x != px {
		panic("x detected incorrectly")
	}
	if y != py {
		panic("y detected incorrectly")
	}
	if score != 1 {
		panic("invalid score")
	}

	// bench
	bench(func() {
		fn(img, pat.(*image.Gray))
	}, 100)
}
