//go:build vk
// +build vk

package pmatch

import "image"

type PMatchVk struct {
	vk *SearchVk
}

// Destroy implements Instance.
func (p *PMatchVk) Destroy() {
	p.vk.Destroy()
}

// Kind implements Instance.
func (p *PMatchVk) Kind() string {
	return "Vk"
}

// SearchRGBA implements Instance.
func (p *PMatchVk) SearchRGBA(img *image.RGBA, pat *image.RGBA) (int, int, float64) {
	maxX, maxY, maxCos, err := p.vk.Run(img, pat)
	if err != nil {
		panic(err)
	}
	return maxX, maxY, maxCos
}

// Compile time interface check.
var _ Instance = (*PMatchVk)(nil)

func NewInstance(imgBounds, patBounds image.Rectangle, imgStride, patStride int) Instance {
	vk, err := NewSearchVk(imgBounds, patBounds, imgStride, patStride, true) // TODO: false.
	if err != nil {
		panic(err)
	}

	return &PMatchVk{vk: vk}
}
