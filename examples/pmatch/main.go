// Package main (pmatch) is a simple binary to demonstrate the usage of the pmatch package.
package main

import (
	"fmt"
	"image"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/pkg/pmatch"
)

const (
	px      = 20
	py      = 40
	patchSz = 50
)

func main() {
	img := imutil.ToGray(pmatch.LoadTestImg())
	rect := image.Rect(px, py, patchSz+px, patchSz+py)
	pat, err := imutil.Sub(img, rect)
	if err != nil {
		panic(err)
	}

	x, y, score := pmatch.SearchGrayC(img, pat.(*image.Gray))
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
}
