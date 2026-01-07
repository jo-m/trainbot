//go:build vk
// +build vk

package pmatch

import (
	"image"

	"github.com/rs/zerolog/log"
)

type params struct {
	ImgBounds, PatBounds image.Rectangle
	ImgStride, PatStride int
}

type PMatchVk struct {
	pool map[params]*SearchVk
}

// Destroy implements Instance.
func (p *PMatchVk) Destroy() {
	for _, inst := range p.pool {
		inst.Destroy()
	}
}

// Kind implements Instance.
func (p *PMatchVk) Kind() string {
	return "Vk"
}

// SearchRGBA implements Instance.
func (p *PMatchVk) SearchRGBA(img *image.RGBA, pat *image.RGBA) (int, int, float64) {
	params := params{img.Bounds(), pat.Bounds(), img.Stride, pat.Stride}

	inst, ok := p.pool[params]

	if !ok {
		log.Warn().Interface("params", params).Msg("allocating instance") // TODO: Remove.

		var err error
		inst, err = NewSearchVk(params.ImgBounds, params.PatBounds, params.ImgStride, params.PatStride, true) // TODO: false.
		if err != nil {
			panic(err)
		}
		p.pool[params] = inst

		if len(p.pool) > 50 {
			panic("too many instances are being allocated, check your image pipeline")
		}
	}

	maxX, maxY, maxCos, err := inst.Run(img, pat)
	if err != nil {
		panic(err)
	}
	return maxX, maxY, maxCos
}

// Compile time interface check.
var _ Instance = (*PMatchVk)(nil)

func NewInstance() Instance {
	return &PMatchVk{pool: map[params]*SearchVk{}}
}
