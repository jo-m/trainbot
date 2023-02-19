package pmatch

import (
	"image"
	"testing"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
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

	// also resets pat bounds origin to (0,0)
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

	// make sure pattern lives in a different memory region
	pat = imutil.ToGray(pat.(*image.Gray))

	for i := 0; i < b.N; i++ {
		SearchGray(img, pat.(*image.Gray))
	}
}
