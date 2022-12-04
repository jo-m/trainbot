// Package pmatch implements image patch matching and search.
package pmatch

import (
	"image"
	"math"
)

// patch window on img
func imgPatchWindow(img, pat image.Image, offset image.Point) image.Image {
	window := pat.Bounds().
		Sub(pat.Bounds().Min).
		Add(img.Bounds().Min).
		Add(offset).
		Intersect(img.Bounds())

	if window.Dx()*window.Dy() != pat.Bounds().Dx()*pat.Bounds().Dy() {
		panic("patch not fully contained in image")
	}
	iface, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if !ok {
		panic("img does not implement SubImage()")
	}
	return iface.SubImage(window)
}

func ScoreGrayCos(img, pat *image.Gray, offset image.Point) float64 {
	img = imgPatchWindow(img, pat, offset).(*image.Gray)

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
		return 1
	}
	cos := float64(dot) / (absI * absP)
	return cos
}

// Note that the alpha channel is ignored.
func ScoreRGBACos(img, pat *image.RGBA, offset image.Point) float64 {
	img = imgPatchWindow(img, pat, offset).(*image.RGBA)

	var dot, sqSumI, sqSumP uint64

	for y := 0; y < pat.Rect.Dy(); y++ {
		for x := 0; x < pat.Rect.Dx(); x++ {
			pxI := img.RGBAAt(img.Bounds().Min.X+x, img.Bounds().Min.Y+y)
			pxP := pat.RGBAAt(pat.Bounds().Min.X+x, pat.Bounds().Min.Y+y)

			dot += uint64(pxI.R) * uint64(pxP.R)
			dot += uint64(pxI.G) * uint64(pxP.G)
			dot += uint64(pxI.B) * uint64(pxP.B)

			sqSumI += uint64(pxI.R) * uint64(pxI.R)
			sqSumI += uint64(pxI.G) * uint64(pxI.G)
			sqSumI += uint64(pxI.B) * uint64(pxI.B)

			sqSumP += uint64(pxP.R) * uint64(pxP.R)
			sqSumP += uint64(pxP.G) * uint64(pxP.G)
			sqSumP += uint64(pxP.B) * uint64(pxP.B)
		}
	}

	absI := math.Sqrt(float64(sqSumI))
	absP := math.Sqrt(float64(sqSumP))
	if absI*absP == 0 {
		return 1
	}
	cos := float64(dot) / (absI * absP)
	return cos
}

func SearchGray(img, pat *image.Gray) (maxX, maxY int, maxScore float64) {
	// search rect in img coordinates
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Rect.Size()).Add(image.Pt(1, 1)),
	}

	for y := 0; y < searchRect.Dy(); y++ {
		for x := 0; x < searchRect.Dx(); x++ {
			score := ScoreGrayCos(img, pat, image.Pt(x, y))

			if score > maxScore {
				maxScore = score
				maxX, maxY = x, y
			}
		}
	}

	return
}

func SearchRGBA(img, pat *image.RGBA) (maxX, maxY int, maxScore float64) {
	// search rect in img coordinates
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Rect.Size()).Add(image.Pt(1, 1)),
	}

	for y := 0; y < searchRect.Dy(); y++ {
		for x := 0; x < searchRect.Dx(); x++ {
			score := ScoreRGBACos(img, pat, image.Pt(x, y))

			if score > maxScore {
				maxScore = score
				maxX, maxY = x, y
			}
		}
	}

	return
}
