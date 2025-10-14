//go:build !vk
// +build !vk

package pmatch

import "image"

// PMatchC implements Instance.
type PMatchC struct{}

// Destroy implements Instance.
func (p *PMatchC) Destroy() {}

// Kind implements Instance.
func (p *PMatchC) Kind() string {
	return "C"
}

// SearchRGBA implements Instance.
func (p *PMatchC) SearchRGBA(img *image.RGBA, pat *image.RGBA) (int, int, float64) {
	return SearchRGBAC(img, pat)
}

// Compile time interface check.
var _ Instance = (*PMatchC)(nil)

// NewInstance instantiates PMatchC.
func NewInstance() Instance {
	return &PMatchC{}
}
