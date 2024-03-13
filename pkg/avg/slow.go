package avg

import (
	"image"
)

func iabs(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

// RGBASlow computes the pixel average, and pixel mean deviation from average,
// on an RGBA image, per channel.
// The alpha channel is ignored.
// Scaled to [0, 1].
// This a slow implementation useful as ground truth for testing.
func RGBASlow(img *image.RGBA) ([3]float64, [3]float64) {
	var sum [3]int64
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			px := img.RGBAAt(x, y)
			sum[0] += int64(px.R)
			sum[1] += int64(px.G)
			sum[2] += int64(px.B)
		}
	}

	cnt := int64(img.Bounds().Size().X * img.Bounds().Size().Y)
	avgPx := [3]int64{
		sum[0] / cnt,
		sum[1] / cnt,
		sum[2] / cnt,
	}

	sum = [3]int64{0, 0, 0}
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			px := img.RGBAAt(x, y)
			sum[0] += iabs(int64(px.R) - avgPx[0])
			sum[1] += iabs(int64(px.G) - avgPx[1])
			sum[2] += iabs(int64(px.B) - avgPx[2])
		}
	}

	avg := [3]float64{
		float64(avgPx[0]) / 255,
		float64(avgPx[1]) / 255,
		float64(avgPx[2]) / 255,
	}
	avgDev := [3]float64{
		float64(sum[0]) / float64(cnt) / 255,
		float64(sum[1]) / float64(cnt) / 255,
		float64(sum[2]) / float64(cnt) / 255,
	}

	return avg, avgDev
}
