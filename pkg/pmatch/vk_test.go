//go:build vk
// +build vk

package pmatch

import (
	"image"
	"testing"

	"github.com/jo-m/trainbot/pkg/imutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SearchRGBAVk_Simple(t *testing.T) {
	img := imutil.ToRGBA(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	require.NoError(t, err)

	SearchRGBAVk(img, pat.(*image.RGBA))
	x, y, score := SearchRGBAVk(img, pat.(*image.RGBA))
	assert.InDelta(t, 1., score, delta)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)

	// Also resets pat bounds origin to (0,0).
	patCopy := imutil.ToRGBA(pat.(*image.RGBA))

	x, y, score = SearchRGBAVk(img, patCopy)
	assert.InDelta(t, 1., score, delta)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)
}

func Test_SearchRGBAVk_Instance(t *testing.T) {
	img := imutil.ToRGBA(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	require.NoError(t, err)

	search, err := NewSearchVk(img.Bounds(), pat.Bounds(), img.Stride, pat.(*image.RGBA).Stride, true)
	assert.NoError(t, err)
	defer search.Destroy()

	search.Run(img, pat.(*image.RGBA))
	x, y, score, err := search.Run(img, pat.(*image.RGBA))
	assert.NoError(t, err)
	assert.InDelta(t, 1., score, delta)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)

	// Also resets pat bounds origin to (0,0).
	patCopy := imutil.ToRGBA(pat.(*image.RGBA))

	x, y, score, err = search.Run(img, patCopy)
	assert.NoError(t, err)
	assert.InDelta(t, 1., score, delta)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)
}

func Benchmark_SearchRGBAVk(b *testing.B) {
	img := imutil.ToRGBA(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	if err != nil {
		b.Error(err)
	}

	// Make sure pattern lives in a different memory region.
	pat = imutil.ToRGBA(pat.(*image.RGBA))

	search, err := NewSearchVk(img.Bounds(), pat.Bounds(), img.Stride, pat.(*image.RGBA).Stride, true)
	if err != nil {
		b.Error(err)
	}
	defer search.Destroy()

	for i := 0; i < b.N; i++ {
		search.Run(img, pat.(*image.RGBA))
	}
}

func Benchmark_SearchRGBAVkNoVal(b *testing.B) {
	img := imutil.ToRGBA(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	if err != nil {
		b.Error(err)
	}

	// Make sure pattern lives in a different memory region.
	pat = imutil.ToRGBA(pat.(*image.RGBA))

	search, err := NewSearchVk(img.Bounds(), pat.Bounds(), img.Stride, pat.(*image.RGBA).Stride, false)
	if err != nil {
		b.Error(err)
	}
	defer search.Destroy()

	for i := 0; i < b.N; i++ {
		search.Run(img, pat.(*image.RGBA))
	}
}
