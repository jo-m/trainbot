package avg

import (
	"image"
)

// GrayOpt computes the pixel average, and pixel mean deviation from average,
// on a (grayscale) image.
// Slightly optimized implementation.
// Scaled to [0, 1].
func GrayOpt(img *image.Gray) (avg, avgDev float64) {
	var sum int64

	m, n := img.Bounds().Dx(), img.Bounds().Dy()
	s := img.Stride

	for y := 0; y < n; y++ {
		ys := y * s
		for x := 0; x < m; x++ {
			ix := ys + x
			px := img.Pix[ix]
			sum += int64(px)
		}
	}

	cnt := int64(img.Bounds().Size().X * img.Bounds().Size().Y)
	avgPx := sum / cnt

	sum = 0
	for y := 0; y < n; y++ {
		ys := y * s
		for x := 0; x < m; x++ {
			ix := ys + x
			px := img.Pix[ix]
			sum += iabs(int64(px) - avgPx)
		}
	}

	return float64(avgPx) / 255, float64(sum) / float64(cnt) / 255
}
