package pmatch

import (
	"image"
	"testing"

	"github.com/jo-m/trainbot/pkg/imutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_SearchGray(t *testing.T) {
	img := imutil.ToGray(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	require.NoError(t, err)

	x, y, score := SearchGray(img, pat.(*image.Gray))
	assert.Equal(t, 1., score)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)

	// Also resets pat bounds origin to (0,0).
	patCopy := imutil.ToGray(pat.(*image.Gray))

	x, y, score = SearchGray(img, patCopy)
	assert.Equal(t, 1., score)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)
}

func Benchmark_SearchGray(b *testing.B) {
	img := imutil.ToGray(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	if err != nil {
		b.Error(err)
	}

	// Make sure pattern lives in a different memory region.
	pat = imutil.ToGray(pat.(*image.Gray))

	for i := 0; i < b.N; i++ {
		SearchGray(img, pat.(*image.Gray))
	}
}

func Test_CosSimGray(t *testing.T) {
	img := imutil.ToGray(LoadTestImg())
	im00, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	require.NoError(t, err)
	im10, err := imutil.Sub(img, image.Rect(x0+1, y0, x0+w+1, y0+h))
	require.NoError(t, err)
	im01, err := imutil.Sub(img, image.Rect(x0, y0+1, x0+w, y0+h+1))
	require.NoError(t, err)
	im11, err := imutil.Sub(img, image.Rect(x0+1, y0+1, x0+w+1, y0+h+1))
	require.NoError(t, err)
	im22, err := imutil.Sub(img, image.Rect(x0+2, y0+2, x0+w+2, y0+h+2))
	require.NoError(t, err)
	im44, err := imutil.Sub(img, image.Rect(x0+4, y0+4, x0+w+4, y0+h+4))
	require.NoError(t, err)
	im88, err := imutil.Sub(img, image.Rect(x0+8, y0+8, x0+w+8, y0+h+8))
	require.NoError(t, err)

	// We expect scores to lower as the offsets increase.

	score := CosSimGray(im00.(*image.Gray), im00.(*image.Gray))
	assert.Equal(t, 1., score)

	score = CosSimGray(im00.(*image.Gray), im10.(*image.Gray))
	assert.Equal(t, 0.9794014202411235, score)

	score = CosSimGray(im00.(*image.Gray), im01.(*image.Gray))
	assert.Equal(t, 0.9843413588376974, score)

	score = CosSimGray(im00.(*image.Gray), im11.(*image.Gray))
	assert.Equal(t, 0.9807568924870074, score)

	score = CosSimGray(im00.(*image.Gray), im22.(*image.Gray))
	assert.Equal(t, 0.9551280220660936, score)

	score = CosSimGray(im00.(*image.Gray), im44.(*image.Gray))
	assert.Equal(t, 0.8991417445923269, score)

	score = CosSimGray(im00.(*image.Gray), im88.(*image.Gray))
	assert.Equal(t, 0.7897102979887818, score)
}

func Benchmark_CosSimGray(b *testing.B) {
	img := imutil.ToGray(LoadTestImg())
	im00, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	if err != nil {
		b.Error(err)
	}
	im88, err := imutil.Sub(img, image.Rect(x0+8, y0+8, x0+w+8, y0+h+8))
	if err != nil {
		b.Error(err)
	}

	// Make sure pattern lives in a different memory region.
	im88 = imutil.ToGray(im88.(*image.Gray))

	for i := 0; i < b.N; i++ {
		CosSimGray(im00.(*image.Gray), im88.(*image.Gray))
	}
}

func Test_CosSimRGBA(t *testing.T) {
	img := imutil.ToRGBA(LoadTestImg())
	im00, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	require.NoError(t, err)
	im10, err := imutil.Sub(img, image.Rect(x0+1, y0, x0+w+1, y0+h))
	require.NoError(t, err)
	im01, err := imutil.Sub(img, image.Rect(x0, y0+1, x0+w, y0+h+1))
	require.NoError(t, err)
	im11, err := imutil.Sub(img, image.Rect(x0+1, y0+1, x0+w+1, y0+h+1))
	require.NoError(t, err)
	im22, err := imutil.Sub(img, image.Rect(x0+2, y0+2, x0+w+2, y0+h+2))
	require.NoError(t, err)
	im44, err := imutil.Sub(img, image.Rect(x0+4, y0+4, x0+w+4, y0+h+4))
	require.NoError(t, err)
	im88, err := imutil.Sub(img, image.Rect(x0+8, y0+8, x0+w+8, y0+h+8))
	require.NoError(t, err)

	// We expect scores to lower as the offsets increase.

	score := CosSimRGBA(im00.(*image.RGBA), im00.(*image.RGBA))
	assert.Equal(t, 1., score)

	score = CosSimRGBA(im00.(*image.RGBA), im10.(*image.RGBA))
	assert.Equal(t, 0.9989434681698076, score)

	score = CosSimRGBA(im00.(*image.RGBA), im01.(*image.RGBA))
	assert.Equal(t, 0.9975862650499367, score)

	score = CosSimRGBA(im00.(*image.RGBA), im11.(*image.RGBA))
	assert.Equal(t, 0.9959661661231728, score)

	score = CosSimRGBA(im00.(*image.RGBA), im22.(*image.RGBA))
	assert.Equal(t, 0.9936682566413522, score)

	score = CosSimRGBA(im00.(*image.RGBA), im44.(*image.RGBA))
	assert.Equal(t, 0.9799362645218272, score)

	score = CosSimRGBA(im00.(*image.RGBA), im88.(*image.RGBA))
	assert.Equal(t, 0.864122120376717, score)
}

func Benchmark_CosSimRGBA(b *testing.B) {
	img := imutil.ToRGBA(LoadTestImg())
	im00, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	if err != nil {
		b.Error(err)
	}
	im88, err := imutil.Sub(img, image.Rect(x0+8, y0+8, x0+w+8, y0+h+8))
	if err != nil {
		b.Error(err)
	}

	// Make sure pattern lives in a different memory region.
	im88 = imutil.ToRGBA(im88.(*image.RGBA))

	for i := 0; i < b.N; i++ {
		CosSimRGBA(im00.(*image.RGBA), im88.(*image.RGBA))
	}
}

func Test_SearchRGBA(t *testing.T) {
	img := imutil.ToRGBA(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	require.NoError(t, err)

	x, y, score := SearchRGBA(img, pat.(*image.RGBA))
	assert.InDelta(t, 1., score, delta)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)

	// Also resets pat bounds origin to (0,0).
	patCopy := imutil.ToRGBA(pat.(*image.RGBA))

	x, y, score = SearchRGBA(img, patCopy)
	assert.InDelta(t, 1., score, delta)
	assert.Equal(t, x0, x)
	assert.Equal(t, y0, y)
}

func Benchmark_SearchRGBA(b *testing.B) {
	img := imutil.ToRGBA(LoadTestImg())
	pat, err := imutil.Sub(img, image.Rect(x0, y0, x0+w, y0+h))
	if err != nil {
		b.Error(err)
	}

	// Make sure pattern lives in a different memory region.
	pat = imutil.ToRGBA(pat.(*image.RGBA))

	for i := 0; i < b.N; i++ {
		SearchRGBA(img, pat.(*image.RGBA))
	}
}
