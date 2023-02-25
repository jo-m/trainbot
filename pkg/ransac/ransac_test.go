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

	const modelNParams = 2
	model := func(x float64, params []float64) float64 {
		return params[0] + params[1]*x*x
	}
	xf := make([]float64, len(testData))
	yf := make([]float64, len(testData))
	for i := range yf {
		xf[i] = float64(i)
		yf[i] = float64(testData[i])
	}

	fit, err := Ransac(xf, yf, model, modelNParams, MetaParams{
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

func Test_Ransac2(t *testing.T) {
	x := []float64{
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
		39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50,
	}
	y := []float64{
		-97.78, 28.13, -168.58, 50.80, 61.93, 56.98, 82.35, 301.81, 106.61, 41.08,
		129.21, 140.78, 155.03, 167.81, 180.02, 382.00, 204.80, 218.80, 230.70,
		247.93, 262.02, 275.77, 286.56, 301.00, 315.51, 333.54, 347.28, 361.53,
		377.77, 393.67, 411.54, 424.81, 444.14, 456.44, 476.06, 494.24, 511.96,
		289.42, 584.95, 562.87, 465.45, 599.09, 617.92, 634.42, 653.89, 674.67,
		690.20, 709.56, 980.63, 749.58, 572.2,
	}

	const modelNParams = 3
	model := func(x float64, params []float64) float64 {
		return params[0] + params[1]*x + 0.5*params[2]*x*x
	}

	fit, err := Ransac(x, y, model, modelNParams, MetaParams{
		MinModelPoints:  4,
		MaxIter:         20,
		MinInliers:      len(x) / 2,
		InlierThreshold: 5.,
		Seed:            123,
	})
	require.NoError(t, err)
	require.InDelta(t, 20, fit.X[0], 5)
	require.InDelta(t, 10, fit.X[1], 1)
	require.InDelta(t, 0.2, fit.X[2], 0.015)
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
		_, err := Ransac(xf, yf, poly, 2, MetaParams{
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
