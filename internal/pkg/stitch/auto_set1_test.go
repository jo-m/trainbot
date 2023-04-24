//go:build moretests
// +build moretests

package stitch

import (
	"image"
	"testing"
)

func Test_AutoStitcher_Set1(t *testing.T) {
	c := Config{
		PixelsPerM:  42,
		MinSpeedKPH: 10,
		MaxSpeedKPH: 120,
		MinLengthM:  10,
	}
	r := image.Rect(0, 0, 206, 290)

	runTestSimple(t, c, r, "testdata/set1/train001.mkv", 174)
	runTestSimple(t, c, r, "testdata/set1/train002.mkv", 100)
	runTestSimple(t, c, r, "testdata/set1/train003.mkv", 91)
	runTestSimple(t, c, r, "testdata/set1/train004.mkv", 91)
	runTestSimple(t, c, r, "testdata/set1/train005.mkv", 91)
	runTestSimple(t, c, r, "testdata/set1/train006.mkv", 204)
	runTestSimple(t, c, r, "testdata/set1/train007.mkv", 181)
	runTestSimple(t, c, r, "testdata/set1/train008.mkv", 196)
	runTestSimple(t, c, r, "testdata/set1/train009.mkv", 138)
	runTestSimple(t, c, r, "testdata/set1/train010.mkv", 177)
	runTestSimple(t, c, r, "testdata/set1/train011.mkv", 186)
	runTestSimple(t, c, r, "testdata/set1/train012.mkv", 187)
	runTestSimple(t, c, r, "testdata/set1/train013.mkv", 203)
	runTestSimple(t, c, r, "testdata/set1/train014.mkv", 193)
	runTestSimple(t, c, r, "testdata/set1/train015.mkv", 277)
	runTestSimple(t, c, r, "testdata/set1/train016.mkv", 203)
	runTestSimple(t, c, r, "testdata/set1/train017.mkv", 178)
	runTestSimple(t, c, r, "testdata/set1/train018.mkv", 271)
	runTestSimple(t, c, r, "testdata/set1/train019.mkv", 273)
	runTestSimple(t, c, r, "testdata/set1/train020.mkv", 90)
	runTestSimple(t, c, r, "testdata/set1/train021.mkv", 300)
	runTestSimple(t, c, r, "testdata/set1/train022.mkv", 275)
	runTestSimple(t, c, r, "testdata/set1/train023.mkv", 177)
	runTestSimple(t, c, r, "testdata/set1/train024.mkv", 184)
	runTestSimple(t, c, r, "testdata/set1/train025.mkv", 91)
	runTestSimple(t, c, r, "testdata/set1/train026.mkv", 92)
	runTestSimple(t, c, r, "testdata/set1/train027.mkv", 92)
	runTestSimple(t, c, r, "testdata/set1/train028.mkv", 18)
	runTestSimple(t, c, r, "testdata/set1/train029.mkv", 92)
	runTestSimple(t, c, r, "testdata/set1/train030.mkv", 98)
	runTestSimple(t, c, r, "testdata/set1/train031.mkv", 17)
	runTestSimple(t, c, r, "testdata/set1/train032.mkv", 183)

	runTestSimple(t, c, r, "testdata/set1/negative001.mkv", 0)
}
