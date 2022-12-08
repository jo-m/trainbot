package est

import (
	"fmt"
	"image"
	"math"
	"time"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/pkg/pmatch"
	"github.com/rs/zerolog/log"
)

const (
	goodScoreNoMove = 0.99
	goodScoreMove   = 0.95
)

type Config struct {
	PixelsPerM  float64
	MinSpeedKPH float64
	MaxSpeedKPH float64

	VideoFPS float64
}

func (e *Config) MinPxPerFrame() int {
	return int(e.MinSpeedKPH*e.PixelsPerM/e.VideoFPS) - 1
}

func (e *Config) MaxPxPerFrame() int {
	return int(e.MaxSpeedKPH*e.PixelsPerM/e.VideoFPS) + 1
}

type Estimator struct {
	c            Config
	minDx, maxDx int

	prevCount int
	prevFrame *image.Gray

	// Those slices always have the same length.
	// dx[i] is the assumed offset between frames[i] and frames[i+1].
	// scores[i] is the score of that assumed offset.
	dx           []int
	scores       []float64 // TODO: maybe remove
	frames       []image.Image
	dxAbsLowPass float64
}

func NewEstimator(c Config) *Estimator {
	return &Estimator{
		c:     c,
		minDx: c.MinPxPerFrame(),
		maxDx: c.MaxPxPerFrame(),
	}
}

func findOffset(prev, curr *image.Gray, maxDx int) (dx int, score float64) {
	if prev.Rect.Size() != curr.Rect.Size() {
		panic("inconsistent size, this should not happen")
	}

	// centered crop from prev frame,
	// width is 3x max pixels per frame given by max velocity
	w := maxDx * 3
	// and 3/4 of frame height
	h := int(float64(prev.Rect.Dy())*3/4 + 1)
	subRect := image.Rect(0, 0, w, h).
		Add(curr.Rect.Min).
		Add(
			curr.Rect.Size().
				Sub(image.Pt(int(w), h)).
				Div(2),
		)
	sub, err := imutil.Sub(prev, subRect)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	// centered slice crop from next frame,
	// width is 1x max pixels per frame given by max velocity
	// and 3/4 of frame height
	sliceRect := image.Rect(0, 0, maxDx, h).
		Add(curr.Rect.Min).
		Add(
			curr.Rect.Size().
				Sub(image.Pt(w, h)).
				Div(2),
		)

	slice, err := imutil.Sub(curr, sliceRect)
	if err != nil {
		log.Panic().Err(err).Send()
	}

	// we expect this x value found by search
	// if nothing has moved
	xZero := sliceRect.Min.Sub(subRect.Min).X

	x, _, score := pmatch.SearchGrayC(sub.(*image.Gray), slice.(*image.Gray))
	return x - xZero, score
}

func (r *Estimator) reset() {
	r.dx = nil
	r.scores = nil
	r.frames = nil
	r.dxAbsLowPass = 0
}

func (r *Estimator) record(dx int, score float64, frame *image.Gray) {
	imutil.Dump(fmt.Sprintf("imgs/frame%05d.jpg", r.prevCount), r.prevFrame) // TODO

	r.dx = append(r.dx, dx)
	r.scores = append(r.scores, score)
	r.frames = append(r.frames, frame)
}

func iabs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func (r *Estimator) process() {
	fmt.Println(r.dx)
	fmt.Println(r.scores)
	dx, err := cleanupDx(r.dx)
	if err != nil {
		log.Panic().Err(err).Send() // TODO: handle properly
	}

	fmt.Println(dx) // TODO: assemble
}

// will NOT make a copy of the image
func (r *Estimator) Frame(frame *image.Gray, ts time.Time) {
	frame = imutil.ToGray(frame)
	defer func() {
		r.prevFrame = frame
		r.prevCount++
	}()

	if r.prevFrame == nil {
		// first time
		return
	}

	dx, score := findOffset(r.prevFrame, frame, r.maxDx)
	log.Debug().Int("prevCount", r.prevCount).Int("dx", dx).Float64("score", score).Msg("received frame")

	isActive := len(r.dx) > 0
	if isActive {
		r.dxAbsLowPass = r.dxAbsLowPass*0.9 + math.Abs(float64(dx))*0.1

		if r.dxAbsLowPass < r.c.MinSpeedKPH {
			log.Info().Msg("stop recording")
			r.process()
			r.reset()
			return
		}

		r.record(dx, score, r.prevFrame)
		return
	} else {
		if score >= goodScoreNoMove && iabs(dx) < r.minDx {
			log.Debug().Msg("not moving")
			return
		}

		if score >= goodScoreMove && iabs(dx) >= r.maxDx {
			log.Info().Msg("start recording")
			r.record(dx, score, r.prevFrame)
			r.dxAbsLowPass = math.Abs(float64(dx))
			return
		}
	}

	log.Panic().Msg("inconclusive")
}
