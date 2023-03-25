package stitch

import (
	"errors"
	"math"
	"time"

	"github.com/jo-m/trainbot/pkg/ransac"
	"github.com/rs/zerolog/log"
)

const modelNParams = 3

// model computes current distance at given time t, assuming constant acceleration.
func model(t float64, params []float64) float64 {
	s0 := params[0]
	v0 := params[1]
	a := params[2]
	return s0 + v0*t + 0.5*a*t*t
}

// Returns fitted dx values. Length will always be the same as the input.
// Does not modify seq.
// Also returns estimated v0 [px/s] and acceleration [px/s^2].
func fitDx(seq sequence) ([]int, float64, float64, error) {
	start := time.Now()
	defer log.Trace().Dur("dur", time.Since(start)).Msg("fitDx() duration")

	// Sanity checks.
	if len(seq.dx) < (modelNParams+1)*3 {
		return nil, 0, 0, errors.New("sequence length too short")
	}

	// For fitting, we want 1. time [s] and total distance [px],
	// both as float. The first entries for both remain 0.
	n := len(seq.dx)
	t := make([]float64, n+1)
	x := make([]float64, n+1)
	dxSum := 0
	for i := range seq.dx {
		t[i+1] = seq.ts[i].Sub(*seq.startTS).Seconds()
		dxSum += seq.dx[i]
		x[i+1] = float64(dxSum)
	}

	// Fit model.
	params := ransac.MetaParams{
		MinModelPoints:  modelNParams + 1,
		MaxIter:         25,
		MinInliers:      len(x) / 2,
		InlierThreshold: 25., // TODO: should depend on pixel density
		Seed:            0,
	}
	fit, err := ransac.Ransac(t, x, model, modelNParams, params)
	if err != nil {
		return nil, 0, 0, err
	}

	// Generate dx from fit, optimized for accumulated rounding error.
	dxFit := make([]int, n)
	xSum := int(math.Round(model(0, fit.X)))
	for i := range seq.dx {
		x := math.Round(model(t[i+1], fit.X))
		dxFit[i] = int(x) - xSum
		xSum += dxFit[i]
	}

	a := fit.X[2]
	v0 := fit.X[1] + a*t[1] // Adjusted to first sample.
	return dxFit, v0, a, nil
}
