package stitch

import (
	"math"
	"time"

	"github.com/jo-m/trainbot/pkg/ransac"
	"github.com/rs/zerolog/log"
)

const modelNParams = 2

func model(dx float64, ps []float64) float64 {
	v0 := ps[0]
	a := ps[1]
	return v0 + a*dx
}

// The resulting slice will have the same length as the input.
func fitDx(ts []time.Time, dx []int) ([]int, error) {
	start := time.Now()
	defer log.Trace().Dur("dur", time.Since(start)).Msg("fitDx() duration")

	if len(dx) != len(ts) {
		panic("this should not happen")
	}

	n := len(dx)
	t0 := ts[0]
	// Convert dx to float and calculate relative time.
	tsF := make([]float64, n) // Will contain zero-based time in seconds.
	dxF := make([]float64, n) // Will contain dx values.
	for i := range tsF {
		tsF[i] = float64(ts[i].Sub(t0).Seconds())
		dxF[i] = float64(dx[i])
	}

	params := ransac.MetaParams{
		MinModelPoints:  4,
		MaxIter:         25,
		MinInliers:      len(dxF) / 2,
		InlierThreshold: 3.,
		Seed:            0,
	}
	fit, err := ransac.Ransac(tsF, dxF, model, modelNParams, params)
	if err != nil {
		return nil, err
	}

	dxFit := make([]int, n)
	for i, tsF := range tsF {
		dxF := model(tsF, fit.X)
		dxFit[i] = int(math.Round(dxF))
	}

	return dxFit, nil
}
