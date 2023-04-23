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

	runTestSimple(t, c, r, "testdata/set1/01.mkv", 138)
	runTestSimple(t, c, r, "testdata/set1/02.mkv", 100)
	runTestSimple(t, c, r, "testdata/set1/03.mkv", 91)
	runTestSimple(t, c, r, "testdata/set1/04.mkv", 91)
	runTestSimple(t, c, r, "testdata/set1/05.mkv", 91)
	runTestSimple(t, c, r, "testdata/set1/06.mkv", 204)
	runTestSimple(t, c, r, "testdata/set1/07.mkv", 181)
	runTestSimple(t, c, r, "testdata/set1/08.mkv", 196)
	runTestSimple(t, c, r, "testdata/set1/09.mkv", 133)
	runTestSimple(t, c, r, "testdata/set1/10.mkv", 168)
	runTestSimple(t, c, r, "testdata/set1/11.mkv", 186)
	runTestSimple(t, c, r, "testdata/set1/12.mkv", 182)
	runTestSimple(t, c, r, "testdata/set1/13.mkv", 203)
	runTestSimple(t, c, r, "testdata/set1/14.mkv", 193)
	runTestSimple(t, c, r, "testdata/set1/15.mkv", 277)
	runTestSimple(t, c, r, "testdata/set1/16.mkv", 203)
	runTestSimple(t, c, r, "testdata/set1/17.mkv", 172)
	runTestSimple(t, c, r, "testdata/set1/18.mkv", 253)
	runTestSimple(t, c, r, "testdata/set1/19.mkv", 267)
	runTestSimple(t, c, r, "testdata/set1/20.mkv", 90)
	runTestSimple(t, c, r, "testdata/set1/21.mkv", 300)
	runTestSimple(t, c, r, "testdata/set1/22.mkv", 268)
	runTestSimple(t, c, r, "testdata/set1/23.mkv", 177)
	runTestSimple(t, c, r, "testdata/set1/24.mkv", 178)
	runTestSimple(t, c, r, "testdata/set1/25.mkv", 91)
	runTestSimple(t, c, r, "testdata/set1/26.mkv", 92)
	runTestSimple(t, c, r, "testdata/set1/27.mkv", 92)
	runTestSimple(t, c, r, "testdata/set1/28.mkv", 18)
	runTestSimple(t, c, r, "testdata/set1/29.mkv", 92)
	runTestSimple(t, c, r, "testdata/set1/30.mkv", 98)
	runTestSimple(t, c, r, "testdata/set1/31.mkv", 17)
	runTestSimple(t, c, r, "testdata/set1/32.mkv", 183)

	runTestSimple(t, c, r, "testdata/set1/falsepositive.mkv", 0)
}
