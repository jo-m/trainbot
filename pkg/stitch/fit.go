package stitch

import (
	"math"
	"time"

	"github.com/jo-m/trainbot/pkg/ransac"
	"github.com/rs/zerolog/log"
)

const modelNParams = 2

func model(x float64, ps []float64) float64 {
	return ps[0] + ps[1]*x*x
}

func sign(x float64) float64 {
	if x > 0 {
		return 1
	}
	if x < 0 {
		return -1
	}
	return 0
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

	var roundErr float64 // Sum of values we have rounded away.
	dxFit := make([]int, n)
	for i, tsF := range tsF {
		dxF := model(tsF, fit.X)
		dxRound := math.Round(dxF)
		roundErr += dxF - dxRound

		if math.Abs(roundErr) >= 1 {
			dxRound += roundErr
			roundErr -= sign(roundErr)
		}

		dxFit[i] = int(dxRound)
	}

	return dxFit, nil
}
