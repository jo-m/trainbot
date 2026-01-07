package pmatch

import (
	"image"
)

// Instance is the common interface for SearchRGBA implementations.
type Instance interface {
	SearchRGBA(img, pat *image.RGBA) (int, int, float64)
	Kind() string
	Destroy()
}
