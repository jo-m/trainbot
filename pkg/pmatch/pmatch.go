// Package pmatch implements image patch matching and search.
package pmatch

import (
	"image"
	"math"
)

type ScoreGrayFn func(img, pat *image.Gray, offset image.Point) float64

func ScoreGrayCos(img, pat *image.Gray, offset image.Point) float64 {
	// patch comparison rect in img coordinates
	imgRect := pat.Bounds().
		Sub(pat.Bounds().Min).
		Add(img.Bounds().Min).
		Add(offset).
		Intersect(img.Bounds())

	if imgRect.Dx()*imgRect.Dy() != pat.Rect.Dx()*pat.Rect.Dy() {
		panic("patch not contained in image")
	}
	img = img.SubImage(imgRect).(*image.Gray)

	var dot, sqSum0, sqSum1 uint64

	for y := 0; y < pat.Rect.Dy(); y++ {
		for x := 0; x < pat.Rect.Dx(); x++ {
			px0 := img.GrayAt(img.Bounds().Min.X+x, img.Bounds().Min.Y+y).Y
			px1 := pat.GrayAt(pat.Bounds().Min.X+x, pat.Bounds().Min.Y+y).Y

			dot += uint64(px0) * uint64(px1)
			sqSum0 += uint64(px0) * uint64(px0)
			sqSum1 += uint64(px1) * uint64(px1)
		}
	}

	abs0 := math.Sqrt(float64(sqSum0))
	abs1 := math.Sqrt(float64(sqSum1))
	if abs0*abs1 == 0 {
		return 1
	}
	cos := float64(dot) / (abs0 * abs1)
	return cos
}

func Search(img, pat *image.Gray, scoreFn ScoreGrayFn) (maxX, maxY int, maxScore float64) {
	// search rect in img coordinates
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Rect.Size()),
	}

	for y := 0; y < searchRect.Dy(); y++ {
		for x := 0; x < searchRect.Dx(); x++ {
			score := scoreFn(img, pat, image.Pt(x, y))

			if score > maxScore {
				maxScore = score
				maxX, maxY = x, y
			}
		}
	}

	return
}

// Note that the alpha channel is ignored.
func ScoreRGBCos(img, pat *image.Gray, offset image.Point) float64 {
	panic("not implemented")
}
