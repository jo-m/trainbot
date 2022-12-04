package pmatch

import (
	"image"
	"math"
)

func SearchGrayOpt(img, pat *image.Gray) (maxX, maxY int, maxScore float64) {
	// search rect in img coordinates
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Rect.Size()),
	}

	for y := 0; y < searchRect.Dy(); y++ {
		for x := 0; x < searchRect.Dx(); x++ {
			offset := image.Pt(x, y)
			var score float64
			{
				img := imgPatchWindow(img, pat, offset).(*image.Gray)

				var dot, sqSumI, sqSumP uint64

				for y := 0; y < pat.Rect.Dy(); y++ {
					for x := 0; x < pat.Rect.Dx(); x++ {
						pxI := img.GrayAt(img.Bounds().Min.X+x, img.Bounds().Min.Y+y)
						pxP := pat.GrayAt(pat.Bounds().Min.X+x, pat.Bounds().Min.Y+y)

						dot += uint64(pxI.Y) * uint64(pxP.Y)
						sqSumI += uint64(pxI.Y) * uint64(pxI.Y)
						sqSumP += uint64(pxP.Y) * uint64(pxP.Y)
					}
				}

				absI := math.Sqrt(float64(sqSumI))
				absP := math.Sqrt(float64(sqSumP))
				if absI*absP == 0 {
					score = 1
				} else {
					score = float64(dot) / (absI * absP)
				}
			}

			if score > maxScore {
				maxScore = score
				maxX, maxY = x, y
			}
		}
	}

	return
}
