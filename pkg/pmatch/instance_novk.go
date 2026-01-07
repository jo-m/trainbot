//go:build !vk
// +build !vk

package pmatch

import "image"

// C implements Instance.
type C struct{}

// Destroy implements Instance.
func (p *C) Destroy() {}

// Kind implements Instance.
func (p *C) Kind() string {
	return "C"
}

// SearchRGBA implements Instance.
func (p *C) SearchRGBA(img *image.RGBA, pat *image.RGBA) (int, int, float64) {
	return SearchRGBAC(img, pat)
}

// Compile time interface check.
var _ Instance = (*C)(nil)

// NewInstance instantiates C.
func NewInstance() Instance {
	return &C{}
}
