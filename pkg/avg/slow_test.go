package avg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"jo-m.ch/go/trainbot/pkg/imutil"
)

func Test_RGBASlow(t *testing.T) {
	high, err := imutil.Load("testdata/high.jpg")
	require.NoError(t, err)
	highRGB := imutil.ToRGBA(high)
	avg, avgDev := RGBASlow(highRGB)
	assert.InDelta(t, 0.45, avg[0], 0.01)
	assert.InDelta(t, 0.38, avg[1], 0.01)
	assert.InDelta(t, 0.44, avg[2], 0.01)
	assert.InDelta(t, 0.25, avgDev[0], 0.01)
	assert.InDelta(t, 0.22, avgDev[1], 0.01)
	assert.InDelta(t, 0.18, avgDev[2], 0.01)

	mid, err := imutil.Load("testdata/mid.jpg")
	require.NoError(t, err)
	midRGB := imutil.ToRGBA(mid)
	avg, avgDev = RGBASlow(midRGB)
	assert.InDelta(t, 0.019, avg[0], 0.001)
	assert.InDelta(t, 0.015, avg[1], 0.001)
	assert.InDelta(t, 0.007, avg[2], 0.001)
	assert.InDelta(t, 0.023, avgDev[0], 0.001)
	assert.InDelta(t, 0.016, avgDev[1], 0.001)
	assert.InDelta(t, 0.009, avgDev[2], 0.001)

	low, err := imutil.Load("testdata/low.jpg")
	require.NoError(t, err)
	lowRGB := imutil.ToRGBA(low)
	avg, avgDev = RGBASlow(lowRGB)
	assert.InDelta(t, 0., avg[0], 0.0001)
	assert.InDelta(t, 0., avg[1], 0.0001)
	assert.InDelta(t, 0., avg[2], 0.0001)
	assert.InDelta(t, 0.0038, avgDev[0], 0.0001)
	assert.InDelta(t, 0.0029, avgDev[1], 0.0001)
	assert.InDelta(t, 0.0022, avgDev[2], 0.0001)
}

func Benchmark_RGBASlow(b *testing.B) {
	high, err := imutil.Load("testdata/high.jpg")
	if err != nil {
		b.Error(err)
	}
	highRGB := imutil.ToRGBA(high)

	for i := 0; i < b.N; i++ {
		RGBASlow(highRGB)
	}
}
