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

// GraySlow computes the pixel average, and pixel mean deviation from average,
// on a (grayscale) image.
// This a completely un-optimized and thus rather slow implementation.
// Scaled to [0, 1].
func GraySlow(img *image.Gray) (avg, avgDev float64) {
	var sum int64
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			px := img.GrayAt(x, y).Y
			sum += int64(px)
		}
	}

	cnt := int64(img.Bounds().Size().X * img.Bounds().Size().Y)
	avgPx := sum / cnt

	sum = 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			px := img.GrayAt(x, y).Y
			sum += iabs(int64(px) - avgPx)
		}
	}

	return float64(avgPx) / 255, float64(sum) / float64(cnt) / 255
}
