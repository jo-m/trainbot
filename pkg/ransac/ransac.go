// Package ransac contains an implementation of the RANSAC algorithm.
package ransac

import (
	"errors"
	"image/color"
	"math"
	"math/rand"

	"github.com/rs/zerolog/log"
	"go-hep.org/x/hep/fit"
	"go-hep.org/x/hep/hplot"
	"gonum.org/v1/gonum/optimize"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func sample(rnd *rand.Rand, x, y []float64, n int) ([]float64, []float64) {
	if n < 1 {
		panic("n must be > 0")
	}
	if len(x) != len(y) {
		panic("x and y must have same length")
	}

	isel := map[int]struct{}{}
	xret, yret := make([]float64, n), make([]float64, n)
	for {
		// Random index.
		i := int(rnd.Intn(len(x)))

		// Make sure is new.
		if _, ok := isel[i]; ok {
			continue
		}
		isel[i] = struct{}{}

		n--
		xret[n] = x[i]
		yret[n] = y[i]
		if n == 0 {
			return xret, yret
		}
	}
}

// ModelFn represents the model to be fitted during RANSAC.
// It has to be implemented by the user.
// Example (poly2):
//
//	func(x float64, params []float64) float64 {
//		return params[0] + params[1]*x*x
//	}
type ModelFn func(x float64, params []float64) float64

// MetaParams contains the meta-parameters for a RANSAC search.
type MetaParams struct {
	MinModelPoints  int
	MaxIter         int
	MinInliers      int
	InlierThreshold float64
	Seed            int64
}

// Check validates RANSAC params and will panic if there are invalid settings.
func (p *MetaParams) Check(nx int) {
	if p.MinModelPoints == 0 {
		panic("MinModelPoints cannot be 0")
	}
	if p.MinModelPoints > nx*3/4 {
		panic("MinModelPoints should be <= len(x)*2/3")
	}
	if p.MaxIter == 0 {
		panic("MaxIter must be > 0")
	}
	if p.MinInliers < p.MinModelPoints {
		panic("MinInliers must be at least MinModelPoints")
	}
}

// Ransac runs the RANSAC algorithm, trying to find model parameters for ModelFn
// according to the meta parameters.
func Ransac(x, y []float64, model ModelFn, p MetaParams) (*optimize.Location, error) {
	if len(x) != len(y) {
		panic("x and y must have same length")
	}
	p.Check(len(x))

	src := rand.NewSource(p.Seed)
	// #nosec G404
	rnd := rand.New(src)

	bestFit := optimize.Location{
		F: math.MaxFloat64,
	}

	for i := 0; i < p.MaxIter; i++ {
		// Sample
		xS, yS := sample(rnd, x, y, p.MinModelPoints)

		// Fit
		params, err := fit.Curve1D(
			fit.Func1D{
				F:  model,
				X:  xS,
				Y:  yS,
				Ps: []float64{1, 1},
			},
			nil, &optimize.NelderMead{},
		)
		if err != nil {
			log.Err(err).Msg("fit did not converge (sample)")
			continue
		}

		// Plot(
		// 	fmt.Sprintf("~/Desktop/fit_sample_%03d_%f.png", i, params.F),
		// 	xS, yS, params.X, model, "f(x) = a + b*x*x")

		// Select inliers
		xIn, yIn := []float64{}, []float64{}
		for j := range x {
			yModel := model(x[j], params.X)
			if math.Abs(yModel-y[j]) < p.InlierThreshold {
				xIn = append(xIn, x[j])
				yIn = append(yIn, y[j])
			}
		}
		if len(xIn) < p.MinInliers {
			continue
		}

		// Fit inliers
		params, err = fit.Curve1D(
			fit.Func1D{
				F:  model,
				X:  xIn,
				Y:  yIn,
				Ps: []float64{1, 1},
			},
			nil, &optimize.NelderMead{},
		)
		if err != nil {
			log.Err(err).Msg("fit did not converge (inliers)")
			continue
		}

		// TODO: require from input distribution of inliers that they are somewhat linearly distributed

		// Plot(
		// 	fmt.Sprintf("~/Desktop/fit_inliers_%03d_%f.png", i, params.F),
		// 	xIn, yIn, params.X, model, "f(x) = a + b*x*x")

		if params.F < bestFit.F {
			bestFit = params.Location
		}
	}

	if bestFit.F == math.MaxFloat64 {
		return nil, errors.New("RANSAC unsuccessful")
	}

	return &bestFit, nil
}

// Plot is a helper to plot the results of a RANSAC iteration.
func Plot(path string, x, y []float64, ps []float64, fn ModelFn, labelX string) {
	p := hplot.New()
	p.X.Label.Text = labelX
	p.Y.Label.Text = "y-data"
	p.X.Min = -10
	p.X.Max = +10
	p.Y.Min = 0
	p.Y.Max = 220

	s := hplot.NewS2D(hplot.ZipXY(x, y))
	s.Color = color.RGBA{0, 0, 255, 255}
	p.Add(s)

	f := plotter.NewFunction(func(x float64) float64 {
		return fn(x, ps)
	})
	f.Color = color.RGBA{255, 0, 0, 255}
	f.Samples = 1000
	p.Add(f)

	p.Add(plotter.NewGrid())

	err := p.Save(20*vg.Centimeter, -1, path)
	if err != nil {
		panic(err)
	}
}
