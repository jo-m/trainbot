package pmatch

import (
	"image"
	"math"
)

// SearchGray searches for the position of a (grayscale) patch in a (grayscale) image,
// using cosine similarity.
// Slightly optimized implementation.
// Panics (due to out of bounds errors) if the patch is larger than the image in any dimension.
func SearchGray(img, pat *image.Gray) (maxX, maxY int, maxCos float64) {
	if pat.Bounds().Size().X > img.Bounds().Size().X ||
		pat.Bounds().Size().Y > img.Bounds().Size().Y {
		panic("patch too large")
	}

	// search rect in img coordinates
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

			imgPatStartIx := y*is + x

			var dot, sqSumI, sqSumP uint64

			for v := 0; v < dv; v++ {
				pxIi := v * is
				pxPi := v * ps

				for u := 0; u < du; u++ {
					pxI := img.Pix[imgPatStartIx+pxIi+u]
					pxP := pat.Pix[pxPi+u]

					dot += uint64(pxI) * uint64(pxP)
					sqSumI += uint64(pxI) * uint64(pxI)
					sqSumP += uint64(pxP) * uint64(pxP)
				}
			}

			abs := float64(sqSumI) * float64(sqSumP)
			var cos2 float64
			if abs == 0 {
				cos2 = 1
			} else {
				cos2 = float64(dot*dot) / abs
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

// CosSimGray returns the cosine similarity score for two (grayscale) images of the same size.
// Slightly optimized implementation.
// Panics (due to out of bounds errors) if the sizes don't match.
func CosSimGray(im0, im1 *image.Gray) (cos float64) {
	if im0.Bounds().Size() != im1.Bounds().Size() {
		panic("image sizes do not match")
	}

	du, dv := im1.Bounds().Dx(), im1.Bounds().Dy()
	is, ps := im0.Stride, im1.Stride

	var dot, sqSum0, sqSum1 uint64

	for v := 0; v < dv; v++ {
		px0i := v * is
		px1i := v * ps

		for u := 0; u < du; u++ {
			px0 := im0.Pix[px0i]
			px1 := im1.Pix[px1i]

			dot += uint64(px0) * uint64(px1)
			sqSum0 += uint64(px0) * uint64(px0)
			sqSum1 += uint64(px1) * uint64(px1)

			px0i++
			px1i++
		}
	}

	abs := float64(sqSum0) * float64(sqSum1)
	if abs == 0 {
		return 1
	}

	return math.Sqrt(float64(dot*dot) / abs)
}

// CosSimRGBA is like CosSimGray() but for RGBA images.
// Note that the alpha channel is ignored.
func CosSimRGBA(im0, im1 *image.RGBA) (cos float64) {
	if im0.Bounds().Size() != im1.Bounds().Size() {
		panic("image sizes do not match")
	}

	du, dv := im1.Bounds().Dx(), im1.Bounds().Dy()
	is, ps := im0.Stride, im1.Stride

	var dot, sqSum0, sqSum1 uint64

	for v := 0; v < dv; v++ {
		px0i := v * is
		px1i := v * ps

		for u := 0; u < du; u++ {
			// R
			px0R := im0.Pix[px0i+0]
			px1R := im1.Pix[px1i+0]
			dot += uint64(px0R) * uint64(px1R)
			sqSum0 += uint64(px0R) * uint64(px0R)
			sqSum1 += uint64(px1R) * uint64(px1R)

			// B
			px0B := im0.Pix[px0i+1]
			px1B := im1.Pix[px1i+1]
			dot += uint64(px0B) * uint64(px1B)
			sqSum0 += uint64(px0B) * uint64(px0B)
			sqSum1 += uint64(px1B) * uint64(px1B)

			// G
			px0G := im0.Pix[px0i+2]
			px1G := im1.Pix[px1i+2]
			dot += uint64(px0G) * uint64(px1G)
			sqSum0 += uint64(px0G) * uint64(px0G)
			sqSum1 += uint64(px1G) * uint64(px1G)

			px0i++
			px1i++
		}
	}

	abs := float64(sqSum0) * float64(sqSum1)
	if abs == 0 {
		return 1
	}

	return math.Sqrt(float64(dot*dot) / abs)
}

const four = 4

// SearchRGBA is like SearchGray, but for RGBA images.
// Note that the alpha channel is ignored.
// Slightly optimized implementation.
// Panics (due to out of bounds errors) if the patch is larger than the image in any dimension.
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

			var dot, sqSumI, sqSumP uint64

			for v := 0; v < dv; v++ {
				pxIi := v * is
				pxPi := v * ps

				for u := 0; u < du; u++ {
					for rgb := 0; rgb < 3; rgb++ {
						pxI := img.Pix[imgPatStartIx+pxIi+u*four+rgb]
						pxP := pat.Pix[pxPi+u*four+rgb]

						dot += uint64(pxI) * uint64(pxP)
						sqSumI += uint64(pxI) * uint64(pxI)
						sqSumP += uint64(pxP) * uint64(pxP)
					}

				}
			}

			abs := float64(sqSumI) * float64(sqSumP)
			var cos2 float64
			if abs == 0 {
				cos2 = 1
			} else {
				cos2 = float64(dot*dot) / abs
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
