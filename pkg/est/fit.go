package est

import (
	"errors"
	"math"

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

func fitDx(dx []int) ([]int, error) {
	n := len(dx)
	// convert x to float and generate y values
	xf := make([]float64, n)
	yf := make([]float64, n)
	for i := range yf {
		xf[i] = float64(dx[i])
		yf[i] = float64(i)
	}

	params := ransac.RansacParams{
		MinModelPoints:  3,
		MaxIter:         25,
		MinInliers:      len(xf) * 3 / 2,
		InlierThreshold: 2.,
		Seed:            0,
	}
	// note that x and y are swapped
	fit, err := ransac.Ransac(yf, xf, poly, params)
	if err != nil {
		return nil, err
	}

	var roundErr float64 // sum of values we have wrongly rounded away
	xfit := make([]int, n)
	for i, y := range yf {
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

func cleanupDx(dx []int) ([]int, error) {
	if len(dx) < 10 {
		return nil, errors.New("len(x) must be >= 10")
	}
	if dx[0] == 0 {
		return nil, errors.New("first dx value cannot be 0")
	}

	// remove trailing zeros
	for dx[len(dx)-1] == 0 {
		dx = dx[:len(dx)-1]
	}

	dxFit, err := fitDx(dx)
	if err != nil {
		return nil, err
	}

	return dxFit, nil
}
