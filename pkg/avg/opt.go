package avg

import (
	"image"
)

const four = 4

// RGBA computes the pixel average, and pixel mean deviation from average,
// on an RGBA image, per channel.
// The alpha channel is ignored.
// Scaled to [0, 1].
// Slightly optimized implementation.
func RGBA(img *image.RGBA) ([3]float64, [3]float64) {
	var sum [3]int64

	m, n := img.Bounds().Dx(), img.Bounds().Dy()
	s := img.Stride

	for y := 0; y < n; y++ {
		ys := y * s
		for x := 0; x < m; x++ {
			ix := ys + x*four
			sum[0] += int64(img.Pix[ix+0])
			sum[1] += int64(img.Pix[ix+1])
			sum[2] += int64(img.Pix[ix+2])
		}
	}

	cnt := int64(m * n)
	avgPx := [3]int64{
		sum[0] / cnt,
		sum[1] / cnt,
		sum[2] / cnt,
	}

	sum = [3]int64{0, 0, 0}
	for y := 0; y < n; y++ {
		ys := y * s
		for x := 0; x < m; x++ {
			ix := ys + x*four

			sum[0] += iabs(int64(img.Pix[ix+0]) - avgPx[0])
			sum[1] += iabs(int64(img.Pix[ix+1]) - avgPx[1])
			sum[2] += iabs(int64(img.Pix[ix+2]) - avgPx[2])
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
