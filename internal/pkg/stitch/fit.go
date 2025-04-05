package stitch

import (
	"errors"
	"math"
	"time"

	"github.com/rs/zerolog/log"
	"jo-m.ch/go/trainbot/pkg/ransac"
)

const modelNParams = 2

// model computes velocity at a given time, assuming constant acceleration.
func model(t float64, params []float64) float64 {
	v0 := params[0]
	a := params[1]
	return v0 + a*t
}

// Returns fitted dx values. Length will always be the same as the input.
// Does not modify seq.
// Also returns estimated length [px], v0 [px/s] and acceleration [px/s^2].
func fitDx(seq sequence, maxSpeedPxS float64) ([]int, float64, float64, float64, error) {
	start := time.Now()
	defer func() {
		log.Trace().Dur("dur", time.Since(start)).Msg("fitDx() duration")
	}()

	// Sanity checks.
	if len(seq.dx) < (modelNParams+1)*3 {
		return nil, 0, 0, 0, errors.New("sequence length too short")
	}

	// Prepare data for fitting.
	n := len(seq.dx)
	dt := make([]float64, n) // Time since last data point [s].
	t := make([]float64, n)  // Time since start [s].
	v := make([]float64, n)  // Current velocity [px/s].
	for i := range seq.dx {
		if i == 0 {
			dt[i] = seq.ts[i].Sub(*seq.startTS).Seconds()
		} else {
			dt[i] = seq.ts[i].Sub(seq.ts[i-1]).Seconds()
		}
		t[i] = seq.ts[i].Sub(*seq.startTS).Seconds()
		v[i] = float64(seq.dx[i]) / dt[i]
	}

	// Fit.
	params := ransac.MetaParams{
		MinModelPoints:  modelNParams + 1,
		MaxIter:         25,
		MinInliers:      len(v) / 2,
		InlierThreshold: maxSpeedPxS * 0.05, // 5% of max speed.
		Seed:            0,
	}
	log.Debug().Floats64("t", t).Floats64("v", v).Ints("dx", seq.dx).Interface("params", params).Msg("RANSAC")
	fit, err := ransac.Ransac(t, v, model, modelNParams, params)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	// Generate dx from fit.
	dxFit := make([]int, n)
	var roundErr float64 // Sum of values we have rounded away.
	for i := range seq.dx {
		dxF := model(t[i], fit.X) * dt[i]
		dxRound := math.Round(dxF)
		roundErr += dxF - dxRound

		if math.Abs(roundErr) >= 0.5 {
			dxRound += roundErr
			roundErr -= sign(roundErr)
		}

		dxFit[i] = int(dxRound)
	}

	log.Debug().Floats64("fit", fit.X).Ints("dxFit", dxFit).Float64("roundErr", roundErr).Msg("RANSAC results")

	v0 := fit.X[0]
	a := fit.X[1]
	ds := v0*t[len(t)-1] + 0.5*a*t[len(t)-1]*t[len(t)-1]
	if ds < 0 {
		ds = -ds
	}
	return dxFit, ds, v0, a, nil
}
