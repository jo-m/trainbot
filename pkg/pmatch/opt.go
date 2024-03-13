package pmatch

import (
	"image"
	"math"
)

const four = 4

// SearchRGBA searches for the position of an (RGBA) patch in an (RGBA) image,
// using cosine similarity.
// Slightly optimized implementation.
// Panics (due to out of bounds errors) if the patch is larger than the image in any dimension.
// The alpha channel is ignored.
func SearchRGBA(img, pat *image.RGBA) (maxX, maxY int, maxCos float64) {
	if pat.Bounds().Size().X > img.Bounds().Size().X ||
		pat.Bounds().Size().Y > img.Bounds().Size().Y {
		panic("patch too large")
	}

	// Search rect in img coordinates.
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Bounds().Size()).Add(image.Pt(1, 1)),
	}

	m, n := searchRect.Dx(), searchRect.Dy()
	du, dv := pat.Bounds().Dx(), pat.Bounds().Dy()

	is, ps := img.Stride, pat.Stride

	var maxCos2 float64
	for y := 0; y < n; y++ {
		for x := 0; x < m; x++ {

			imgPatStartIx := y*is + x*four

			var dot, absI2, absP2 uint64

			for v := 0; v < dv; v++ {
				pxIi := v * is
				pxPi := v * ps

				for u := 0; u < du; u++ {
					for rgb := 0; rgb < 3; rgb++ {
						pxI := img.Pix[imgPatStartIx+pxIi+u*four+rgb]
						pxP := pat.Pix[pxPi+u*four+rgb]

						dot += uint64(pxI) * uint64(pxP)
						absI2 += uint64(pxI) * uint64(pxI)
						absP2 += uint64(pxP) * uint64(pxP)
					}

				}
			}

			abs2 := float64(absI2) * float64(absP2)
			var cos2 float64
			if abs2 == 0 {
				cos2 = 1
			} else {
				cos2 = float64(dot) * float64(dot) / abs2
			}

			if cos2 > maxCos2 {
				maxCos2 = cos2
				maxX, maxY = x, y
			}
		}
	}

	// This was left out above.
	maxCos = math.Sqrt(maxCos2)

	return
}
