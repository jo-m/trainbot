package pmatch

import (
	"image"
	"testing"

	"github.com/jo-m/trainbot/pkg/imutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SearchRGBAVk(t *testing.T) {
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

func Benchmark_SearchRGBAVk(b *testing.B) {
	img := imutil.ToRGBA(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	if err != nil {
		b.Error(err)
	}

	// Make sure pattern lives in a different memory region.
	pat = imutil.ToRGBA(pat.(*image.RGBA))

	for i := 0; i < b.N; i++ {
		SearchRGBAVk(img, pat.(*image.RGBA))
	}
}
