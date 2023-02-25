package ransac

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_sample(t *testing.T) {
	x := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9}
	y := []float64{10, 20, 30, 40, 50, 60, 70, 80, 90}

	src := rand.NewSource(123)
	rnd := rand.New(src)

	for i := 0; i < 1000; i++ {
		xs, ys := sample(rnd, x, y, 4)

		xset := map[float64]struct{}{}
		for j := range xs {
			assert.Equal(t, xs[j]*10, ys[j])
			xset[xs[j]] = struct{}{}
		}

		assert.Equal(t, len(xs), len(xset))
	}
}

func Test_Ransac(t *testing.T) {
	testData := []int{
		34, 34, 34, 34, 34, 26, 0, 34, 1, 1, 0, 0, 20, 0, 34, 34, 34, 34, 25, 34, 34,
		34, 34, 34, 34, 34, 34, 34, 22, 0, 34, 34, 28, 34, 34, 26, 27, 34, 34, 34, 34,
		34, 34, 0, 0, 34, 34, 34, 0, 34, 34, 34, 34, 1, 34, 34, 22, 34, 34, 34, 34, 0,
		34, 0, 34, 34, 26, 34, 34, 34, 3, 34, 34, 32, 34, 34, 34, 7, 0, 34, 0, 34, 1,
		34, 34, 0, 34, 34, 5, 34, 5, 34, 27, 0, 0, 34, 34, 34, 34, 34, 32, 31, 34, 34,
		29, 25, 34, 10, 0, 6, 0, 34, 0, 34, 1, 24, 34, 34, 35,
	}

	poly := func(x float64, params []float64) float64 {
		return params[0] + params[1]*x*x
	}
	xf := make([]float64, len(testData))
	yf := make([]float64, len(testData))
	for i := range yf {
		xf[i] = float64(i)
		yf[i] = float64(testData[i])
	}

	fit, err := Ransac(xf, yf, poly, MetaParams{
		MinModelPoints:  3,
		MaxIter:         10,
		MinInliers:      len(xf) / 2,
		InlierThreshold: 2.,
		Seed:            123,
	})
	require.NoError(t, err)
	require.InDelta(t, 34, fit.X[0], 0.1)
	require.InDelta(t, 0, fit.X[1], 0.001)
}

func Benchmark_Ransac(b *testing.B) {
	testData := []int{
		34, 34, 34, 34, 34, 26, 0, 34, 1, 1, 0, 0, 20, 0, 34, 34, 34, 34, 25, 34, 34,
		34, 34, 34, 34, 34, 34, 34, 22, 0, 34, 34, 28, 34, 34, 26, 27, 34, 34, 34, 34,
		34, 34, 0, 0, 34, 34, 34, 0, 34, 34, 34, 34, 1, 34, 34, 22, 34, 34, 34, 34, 0,
		34, 0, 34, 34, 26, 34, 34, 34, 3, 34, 34, 32, 34, 34, 34, 7, 0, 34, 0, 34, 1,
		34, 34, 0, 34, 34, 5, 34, 5, 34, 27, 0, 0, 34, 34, 34, 34, 34, 32, 31, 34, 34,
		29, 25, 34, 10, 0, 6, 0, 34, 0, 34, 1, 24, 34, 34, 35,
	}

	poly := func(x float64, ps []float64) float64 {
		return ps[0] + ps[1]*x*x
	}
	xf := make([]float64, len(testData))
	yf := make([]float64, len(testData))
	for i := range yf {
		xf[i] = float64(i)
		yf[i] = float64(testData[i])
	}

	for i := 0; i < b.N; i++ {
		_, err := Ransac(xf, yf, poly, MetaParams{
			MinModelPoints:  3,
			MaxIter:         10,
			MinInliers:      len(xf) / 2,
			InlierThreshold: 2.,
			Seed:            123,
		})

		if err != nil {
			b.Error(err)
		}
	}
}
