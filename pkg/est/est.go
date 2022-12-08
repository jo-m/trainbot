package est

import (
	"fmt"
	"image"
	"time"

	"github.com/jo-m/trainbot/internal/pkg/imutil"
	"github.com/jo-m/trainbot/pkg/pmatch"
	"github.com/rs/zerolog/log"
)

const (
	magicBad        = 0xDEADDEAD
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
	c     Config
	maxDx int

	prevCount int
	prevFrame *image.Gray
}

func NewEstimator(c Config) *Estimator {
	return &Estimator{
		c:     c,
		maxDx: c.MaxPxPerFrame(),
	}
}

func findOffset(prev, curr *image.Gray, maxDx int) (score float64, dx int) {
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
	return score, x - xZero
}

// will NOT make a copy of the image
func (r *Estimator) Frame(frame *image.Gray, ts time.Time) error {
	frame = imutil.ToGray(frame)
	defer func() {
		r.prevFrame = frame
		r.prevCount++
	}()

	if r.prevFrame == nil {
		// first time
		return nil
	}

	score, dx := findOffset(r.prevFrame, frame, r.maxDx)
	fmt.Println(r.prevCount, score, dx)
	imutil.Dump(fmt.Sprintf("imgs/frame%05d.jpg", r.prevCount), r.prevFrame)

	return nil
}
