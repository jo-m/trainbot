package pmatch

import (
	"image"
	"math"
)

func SearchGrayOpt(img, pat *image.Gray) (maxX, maxY int, maxScore float64) {
	if pat.Bounds().Size().X > img.Bounds().Size().X ||
		pat.Bounds().Size().Y > img.Bounds().Size().Y {
		panic("patch too large")
	}

	// search rect in img coordinates
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Rect.Size()),
	}

	m, n := searchRect.Dx(), searchRect.Dy()
	du, dv := pat.Rect.Dx(), pat.Rect.Dy()

	for y := 0; y < n; y++ {
		for x := 0; x < m; x++ {
			imgWinRect := pat.Bounds().
				Sub(pat.Bounds().Min).
				Add(img.Bounds().Min).
				Add(image.Pt(x, y))
			imgPat := img.SubImage(imgWinRect).(*image.Gray)

			var dot, sqSumI, sqSumP uint64

			for v := 0; v < dv; v++ {
				for u := 0; u < du; u++ {
					pxI := imgPat.GrayAt(
						imgPat.Bounds().Min.X+u,
						imgPat.Bounds().Min.Y+v,
					)
					pxP := pat.GrayAt(
						pat.Bounds().Min.X+u,
						pat.Bounds().Min.Y+v,
					)

					dot += uint64(pxI.Y) * uint64(pxP.Y)
					sqSumI += uint64(pxI.Y) * uint64(pxI.Y)
					sqSumP += uint64(pxP.Y) * uint64(pxP.Y)
				}
			}

			abs := math.Sqrt(float64(sqSumI) * float64(sqSumP))
			var score float64
			if abs == 0 {
				score = 1
			} else {
				score = float64(dot) / abs
			}

			if score > maxScore {
				maxScore = score
				maxX, maxY = x, y
			}
		}
	}

	return
}
