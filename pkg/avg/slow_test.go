package avg

import (
	"testing"

	"github.com/jo-m/trainbot/pkg/imutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GraySlow(t *testing.T) {
	high, err := imutil.Load("testdata/high.jpg")
	require.NoError(t, err)
	highG := imutil.ToGray(high)
	avg, avgDev := GraySlow(highG)
	assert.InDelta(t, 0.41, avg, 0.01)
	assert.InDelta(t, 0.22, avgDev, 0.01)

	mid, err := imutil.Load("testdata/mid.jpg")
	require.NoError(t, err)
	midG := imutil.ToGray(mid)
	avg, avgDev = GraySlow(midG)
	assert.InDelta(t, 0.019, avg, 0.001)
	assert.InDelta(t, 0.018, avgDev, 0.001)

	low, err := imutil.Load("testdata/low.jpg")
	require.NoError(t, err)
	lowG := imutil.ToGray(low)
	avg, avgDev = GraySlow(lowG)
	assert.InDelta(t, 0., avg, 0.0001)
	assert.InDelta(t, 0.00324, avgDev, 0.0001)
}

func Benchmark_GraySlow(b *testing.B) {
	high, err := imutil.Load("testdata/high.jpg")
	if err != nil {
		b.Error(err)
	}
	highG := imutil.ToGray(high)

	for i := 0; i < b.N; i++ {
		GraySlow(highG)
	}
}
