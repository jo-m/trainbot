package stitch

import (
	"math"
	"time"

	"github.com/jo-m/trainbot/pkg/ransac"
)

func poly(x float64, ps []float64) float64 {
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

// the resulting slice must have same length as the input
func fitDx(ts []time.Time, dx []int) ([]int, error) {
	if len(dx) != len(ts) {
		panic("this should not happen")
	}

	n := len(dx)
	t0 := ts[0]
	// convert dx to float and calculate time offset in seconds
	tsSec := make([]float64, n)  // will contain ts
	values := make([]float64, n) // will contain dx values
	for i := range tsSec {
		tsSec[i] = float64(ts[i].Sub(t0).Seconds())
		values[i] = float64(dx[i])
	}

	params := ransac.RansacParams{
		MinModelPoints:  3,
		MaxIter:         25,
		MinInliers:      len(values) / 2,
		InlierThreshold: 3.,
		Seed:            0,
	}
	fit, err := ransac.Ransac(tsSec, values, poly, params)
	if err != nil {
		return nil, err
	}

	var roundErr float64 // sum of values we have wrongly rounded away
	xfit := make([]int, n)
	for i, y := range tsSec {
		x := poly(y, fit.X)
		xRound := math.Round(x)
		roundErr += x - xRound

		if math.Abs(roundErr) >= 1 {
			xRound += roundErr
			roundErr -= sign(roundErr)
		}

		xfit[i] = int(xRound)
	}

	return xfit, nil
}
