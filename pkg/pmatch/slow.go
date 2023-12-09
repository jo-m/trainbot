package pmatch

import (
	"image"
	"math"
)

// imgPatchWindow computes the patch window on img.
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

// ScoreGrayCosSlow computes the cosine similarity score for a (grayscale) patch
// on a (grayscale) image.
// This a completely un-optimized and thus rather slow implementation.
func ScoreGrayCosSlow(img, pat *image.Gray, offset image.Point) (cos float64) {
	img = imgPatchWindow(img, pat, offset).(*image.Gray)

	var dot, absI2, absP2 uint64

	for y := 0; y < pat.Bounds().Dy(); y++ {
		for x := 0; x < pat.Bounds().Dx(); x++ {
			pxI := img.GrayAt(img.Bounds().Min.X+x, img.Bounds().Min.Y+y)
			pxP := pat.GrayAt(pat.Bounds().Min.X+x, pat.Bounds().Min.Y+y)

			dot += uint64(pxI.Y) * uint64(pxP.Y)
			absI2 += uint64(pxI.Y) * uint64(pxI.Y)
			absP2 += uint64(pxP.Y) * uint64(pxP.Y)
		}
	}

	abs2 := float64(absI2) * float64(absP2)
	if abs2 == 0 {
		return 1
	}
	return float64(dot) / math.Sqrt(abs2)
}

// ScoreRGBACosSlow is like ScoreGrayCosSlow() but for RGBA images.
// Note that the alpha channel is ignored.
func ScoreRGBACosSlow(img, pat *image.RGBA, offset image.Point) (cos float64) {
	img = imgPatchWindow(img, pat, offset).(*image.RGBA)

	var dot, absI2, absP2 uint64

	for y := 0; y < pat.Rect.Dy(); y++ {
		for x := 0; x < pat.Rect.Dx(); x++ {
			pxI := img.RGBAAt(img.Bounds().Min.X+x, img.Bounds().Min.Y+y)
			pxP := pat.RGBAAt(pat.Bounds().Min.X+x, pat.Bounds().Min.Y+y)

			dot += uint64(pxI.R) * uint64(pxP.R)
			dot += uint64(pxI.G) * uint64(pxP.G)
			dot += uint64(pxI.B) * uint64(pxP.B)

			absI2 += uint64(pxI.R) * uint64(pxI.R)
			absI2 += uint64(pxI.G) * uint64(pxI.G)
			absI2 += uint64(pxI.B) * uint64(pxI.B)

			absP2 += uint64(pxP.R) * uint64(pxP.R)
			absP2 += uint64(pxP.G) * uint64(pxP.G)
			absP2 += uint64(pxP.B) * uint64(pxP.B)
		}
	}

	abs2 := float64(absI2) * float64(absP2)
	if abs2 == 0 {
		return 1
	}
	return float64(dot) / math.Sqrt(abs2)
}

// SearchGraySlow searches for the position of a (grayscale) patch in a (grayscale) image,
// using cosine similarity.
// This a completely un-optimized and thus rather slow implementation.
// Panics (due to out of bounds errors) if the patch is larger than the image in any dimension.
func SearchGraySlow(img, pat *image.Gray) (maxX, maxY int, maxCos float64) {
	// Search rect in img coordinates.
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Bounds().Size()).Add(image.Pt(1, 1)),
	}

	for y := 0; y < searchRect.Dy(); y++ {
		for x := 0; x < searchRect.Dx(); x++ {
			cos := ScoreGrayCosSlow(img, pat, image.Pt(x, y))

			if cos > maxCos {
				maxCos = cos
				maxX, maxY = x, y
			}
		}
	}

	return
}

// SearchRGBASlow is like SearchGraySlow(), but for RGBA images.
// Note that the alpha channel is ignored.
func SearchRGBASlow(img, pat *image.RGBA) (maxX, maxY int, maxCos float64) {
	// Search rect in img coordinates.
	searchRect := image.Rectangle{
		Min: img.Bounds().Min,
		Max: img.Bounds().Max.Sub(pat.Bounds().Size()).Add(image.Pt(1, 1)),
	}

	for y := 0; y < searchRect.Dy(); y++ {
		for x := 0; x < searchRect.Dx(); x++ {
			cos := ScoreRGBACosSlow(img, pat, image.Pt(x, y))

			if cos > maxCos {
				maxCos = cos
				maxX, maxY = x, y
			}
		}
	}

	return
}
