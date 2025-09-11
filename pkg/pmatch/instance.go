package pmatch

import (
	"image"
)

type Instance interface {
	SearchRGBA(img, pat *image.RGBA) (int, int, float64)
	Kind() string
	Destroy()
}
