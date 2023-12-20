package pmatch

import (
	"image"
	"image/color"
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

// ScoreRGBACosSlow computes the cosine similarity score for an (RGBA) patch
// on an (RGBA) image.
// This a slow implementation useful as ground truth for testing.
// The alpha channel is ignored.
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

func colscale(val float64) color.RGBA {
	ret := color.RGBA{A: 255}

	if val < 0.8 {
		ret.R = uint8(val / 0.8 * 255)
	} else if val < 0.9 {
		ret.B = uint8((val - 0.8) / 0.1 * 255)
	} else {
		ret.G = uint8((val - 0.9) / 0.1 * 255)
	}

	return ret
}

// SearchRGBASlow searches for the position of an (RGBA) patch in an (RGBA) image,
// using cosine similarity.
// This a slow implementation useful as ground truth for testing.
// Panics (due to out of bounds errors) if the patch is larger than the image in any dimension.
// The alpha channel is ignored.
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
